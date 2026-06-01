#Requires -Version 5.1
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Write-Info {
    param([string]$Message)
    Write-Host $Message
}

function Die {
    param([string]$Message)
    [Console]::Error.WriteLine("ERROR: $Message")
    exit 1
}

function Resolve-UserPath {
    param([string]$PathValue)

    if ([string]::IsNullOrWhiteSpace($PathValue)) { return $PathValue }
    if ($PathValue.StartsWith('~/') -or $PathValue.StartsWith('~\')) {
        return Join-Path $HOME $PathValue.Substring(2)
    }
    if ($PathValue -eq '~') { return $HOME }
    return $PathValue
}

function Assert-SafeSnapshotKey {
    param([string]$Key)

    if ([string]::IsNullOrWhiteSpace($Key)) {
        Die "unsafe agent key for snapshot path: empty"
    }
    if ($Key.Contains('/') -or $Key.Contains('\') -or $Key.Contains('..') -or $Key.Contains([char]0)) {
        Die "unsafe agent key for snapshot path: '$Key'"
    }
    return $Key
}

function Write-Utf8NoBomAtomic {
    param(
        [string]$Path,
        [string]$Content
    )

    $Directory = Split-Path -Parent $Path
    if (-not (Test-Path -LiteralPath $Directory)) {
        New-Item -ItemType Directory -Path $Directory -Force | Out-Null
    }

    $TempPath = Join-Path $Directory ([System.IO.Path]::GetRandomFileName())
    $Utf8NoBom = New-Object System.Text.UTF8Encoding($false)
    [System.IO.File]::WriteAllText($TempPath, $Content, $Utf8NoBom)

    if (Test-Path -LiteralPath $Path) {
        [System.IO.File]::Replace($TempPath, $Path, $null)
    }
    else {
        [System.IO.File]::Move($TempPath, $Path)
    }
}

function Copy-Utf8NoBomFile {
    param(
        [string]$SourcePath,
        [string]$DestinationPath
    )

    $Content = (Get-Content -LiteralPath $SourcePath -Raw).TrimEnd("`n", "`r")
    Write-Utf8NoBomAtomic -Path $DestinationPath -Content $Content
}

function Ensure-LfTerminated {
    param([string]$Content)

    if (-not $Content.EndsWith("`n")) {
        return $Content + "`n"
    }
    return $Content
}

function Normalize-Lf {
    param([string]$Content)

    if ($null -eq $Content) { return '' }
    return ($Content -replace "`r`n", "`n")
}

function Get-Sha256Hex {
    param([string]$Content)

    $Normalized = Ensure-LfTerminated (Normalize-Lf $Content)
    $Bytes = [System.Text.Encoding]::UTF8.GetBytes($Normalized)
    $Sha256 = [System.Security.Cryptography.SHA256]::Create()
    try {
        return (($Sha256.ComputeHash($Bytes) | ForEach-Object { $_.ToString('x2') }) -join '')
    }
    finally {
        $Sha256.Dispose()
    }
}

function Read-SimpleYamlMap {
    param([string]$Path)

    $Map = @{}
    try {
        $Lines = Get-Content -LiteralPath $Path
    }
    catch {
        Die ("Cannot read audited snapshot metadata at {0}: {1}" -f $Path, $_.Exception.Message)
    }

    for ($i = 0; $i -lt $Lines.Count; $i++) {
        $RawLine = [string]$Lines[$i]
        $Line = $RawLine.Trim()
        if ([string]::IsNullOrWhiteSpace($Line) -or $Line.StartsWith('#')) {
            continue
        }
        $Parts = $RawLine.Split(':', 2)
        if ($Parts.Count -ne 2) {
            Die ("invalid metadata line {0} in {1}: missing ':' separator" -f ($i + 1), $Path)
        }
        $Map[$Parts[0].Trim()] = $Parts[1].Trim()
    }

    return $Map
}

function Unescape-NonAsciiUnicode {
    param([string]$Json)

    # ConvertTo-Json on PS5.1 escapes non-ASCII as \uXXXX. Bash/Python uses
    # ensure_ascii=False and keeps raw UTF-8. To produce byte-identical
    # opencode.json across both helpers, unescape \uXXXX back to UTF-8 here.
    #
    # Limitation: uses a negative lookbehind to skip \\uXXXX (literal backslash-u
    # in JSON). This handles the common case but may not be correct for sequences
    # with an odd number of consecutive backslashes before \uXXXX. Orchestrator
    # prompts are markdown documents that don't contain literal \uXXXX sequences,
    # so this is safe for our workload.
    if ([string]::IsNullOrEmpty($Json)) { return $Json }
    $Regex = New-Object System.Text.RegularExpressions.Regex '(?<!\\)\\u([0-9a-fA-F]{4})'
    $Evaluator = [System.Text.RegularExpressions.MatchEvaluator] {
        param($Match)
        $CodePoint = [Convert]::ToInt32($Match.Groups[1].Value, 16)
        return [char]$CodePoint
    }
    return $Regex.Replace($Json, $Evaluator)
}

function Remove-ExactOnce {
    param(
        [string]$Text,
        [string]$Old,
        [string]$New,
        [string]$Label
    )

    $Index = $Text.IndexOf($Old, [System.StringComparison]::Ordinal)
    if ($Index -lt 0) {
        Die "missing expected text: $Label"
    }

    # Single-occurrence replace, mirroring Python's str.replace(old, new, 1).
    return $Text.Substring(0, $Index) + $New + $Text.Substring($Index + $Old.Length)
}

function Remove-RegexOnce {
    param(
        [string]$Text,
        [string]$Pattern,
        [string]$Replacement,
        [string]$Label
    )

    $Regex = New-Object System.Text.RegularExpressions.Regex($Pattern)
    $Matches = $Regex.Matches($Text)
    if ($Matches.Count -eq 0) {
        Die "missing expected block: $Label"
    }

    return $Regex.Replace($Text, $Replacement, 1)
}

function Is-OrchestratorAgent {
    param(
        [string]$AgentKey,
        [string[]]$ExactKeys,
        [string[]]$Prefixes
    )

    if ($ExactKeys -contains $AgentKey) { return $true }
    foreach ($Prefix in $Prefixes) {
        if ($AgentKey.StartsWith($Prefix)) { return $true }
    }
    return $false
}

function Is-ProfileOrchestratorKey {
    param(
        [string]$AgentKey,
        [string]$ProfilePrefix
    )

    if ([string]::IsNullOrEmpty($ProfilePrefix)) { return $false }
    return $AgentKey.StartsWith($ProfilePrefix)
}

function Get-ProfileNameFromOrchestratorKey {
    param(
        [string]$AgentKey,
        [string]$ProfilePrefix
    )

    if ([string]::IsNullOrEmpty($ProfilePrefix)) { return '' }
    if (-not $AgentKey.StartsWith($ProfilePrefix)) { return '' }
    return $AgentKey.Substring($ProfilePrefix.Length)
}

function Get-AgentPropertyValue {
    param(
        [PSCustomObject]$Agent,
        [string]$Name
    )

    if ($null -eq $Agent) { return $null }
    if ($Agent.PSObject.Properties.Name -contains $Name) {
        return $Agent.$Name
    }
    return $null
}

function Test-RepoSnapshotEligible {
    param(
        [string]$AgentKey,
        [object]$PolicyObject
    )

    return $PolicyObject.opencode.orchestrator_agent_keys -contains $AgentKey
}

function Write-SnapshotWithStatus {
    param(
        [string]$Path,
        [string]$Content,
        [ref]$NewCounter,
        [ref]$ChangedCounter,
        [ref]$UnchangedCounter
    )

    $Normalized = Ensure-LfTerminated $Content
    if (Test-Path -LiteralPath $Path -PathType Leaf) {
        $OldSnapshotRaw = Get-Content -LiteralPath $Path -Raw
        if ((Normalize-Lf $OldSnapshotRaw) -ne (Normalize-Lf $Normalized)) {
            $ChangedCounter.Value++
            $Status = 'changed'
        }
        else {
            $UnchangedCounter.Value++
            $Status = 'unchanged'
        }
    }
    else {
        $NewCounter.Value++
        $Status = 'new'
    }

    Write-Utf8NoBomAtomic -Path $Path -Content $Normalized
    return $Status
}

function Ensure-LocalSnapshotFromRepo {
    param(
        [string]$AgentKey,
        [string]$RepoSnapshotPath,
        [string]$LocalSnapshotPath,
        [ref]$MigrationCounter
    )

    if ((Test-Path -LiteralPath $LocalSnapshotPath -PathType Leaf) -or -not (Test-Path -LiteralPath $RepoSnapshotPath -PathType Leaf)) {
        return $false
    }

    Copy-Utf8NoBomFile -SourcePath $RepoSnapshotPath -DestinationPath $LocalSnapshotPath
    $MigrationCounter.Value++
    Write-Info "  migrated snapshot $AgentKey -> $LocalSnapshotPath (from repo versioned snapshot)"
    return $true
}

function Ensure-RepoSnapshotFromLocal {
    param(
        [string]$AgentKey,
        [string]$RepoSnapshotPath,
        [string]$LocalSnapshotPath,
        [object]$PolicyObject,
        [ref]$BackfillCounter
    )

    if (-not (Test-RepoSnapshotEligible -AgentKey $AgentKey -PolicyObject $PolicyObject)) {
        return $false
    }
    if ((Test-Path -LiteralPath $RepoSnapshotPath -PathType Leaf) -or -not (Test-Path -LiteralPath $LocalSnapshotPath -PathType Leaf)) {
        return $false
    }

    Copy-Utf8NoBomFile -SourcePath $LocalSnapshotPath -DestinationPath $RepoSnapshotPath
    $BackfillCounter.Value++
    Write-Info "  backfilled repo snapshot $AgentKey -> $RepoSnapshotPath (from local operational snapshot)"
    return $true
}

function Validate-Assignment {
    param(
        [string]$Label,
        [object]$Value
    )

    if (-not ($Value -is [PSCustomObject])) {
        Die "$Label`: must be an object with 'model' and 'variant'"
    }
    $AllowedFields = @('model', 'variant')
    foreach ($PropName in $Value.PSObject.Properties.Name) {
        if (-not ($AllowedFields -contains $PropName)) {
            Die "$Label`: unexpected field '$PropName'; only 'model' and 'variant' are allowed"
        }
    }
    if (-not ($Value.PSObject.Properties.Name -contains 'model')) {
        Die "$Label`: missing required field 'model'"
    }
    if (-not ($Value.PSObject.Properties.Name -contains 'variant')) {
        Die "$Label`: missing required field 'variant' (use '' if the assignment has no variant)"
    }
    if ($Value.model -isnot [string] -or [string]::IsNullOrEmpty($Value.model)) {
        Die "$Label`: field 'model' must be a non-empty string"
    }
    if ($Value.variant -isnot [string]) {
        Die "$Label`: field 'variant' must be a string (use '' for no variant)"
    }
}

function Sanitize-OrchestratorPrompt {
    param(
        [string]$Prompt,
        [object]$SanitizerPolicy
    )

    foreach ($Marker in $SanitizerPolicy.required_markers) {
        if ($Prompt.IndexOf([string]$Marker, [System.StringComparison]::Ordinal) -lt 0) {
            Die "missing required marker before sanitizing: $Marker"
        }
    }

    $Text = $Prompt

    # Bash removes preflight choices 3 and 4 as a single block (one replace_once
    # with both lines concatenated). Mirror that here so failure semantics are
    # identical: if upstream reorders/interleaves these lines, BOTH scripts
    # fail with "preflight PR/review choices" rather than silently sanitizing
    # half of them.
    $PreflightBlock = '3. **Chained PR strategy**: `auto-forecast`, `ask-always`, `single-pr-default`, or `force-chained`.' + "`n" + '4. **Review budget**: maximum changed lines before stopping for reviewer-burden approval.' + "`n"
    $Text = Remove-ExactOnce $Text $PreflightBlock '' 'preflight PR/review choices'

    $Text = Remove-ExactOnce $Text 'Reply with "use recommended" or with codes like: A1, B1, C1, D1.' 'Reply with "use recommended" or with codes like: A1, B1.' 'english preflight codes'
    $Text = Remove-ExactOnce $Text 'Respondé con "usar recomendado" o con códigos como: A1, B1, C1, D1.' 'Respondé con "usar recomendado" o con códigos como: A1, B1.' 'spanish preflight codes'
    $Text = Remove-RegexOnce $Text '(?ms)^C\. PRs\n.*?^   D3 Other: ask for the number afterwards\.\n' '' 'english PR/review prompt block'
    $Text = Remove-RegexOnce $Text '(?ms)^C\. PRs\n.*?^   D3 Otro: preguntar el número después\.\n' '' 'spanish PR/review prompt block'
    $Text = Remove-RegexOnce $Text '(?m)^- PRs:.*\n' '' 'PR answer mapping'
    $Text = Remove-RegexOnce $Text '(?m)^- Review:.*\n' '' 'review answer mapping'
    $Text = Remove-ExactOnce $Text 'If the user explicitly provided all four choices in the current conversation, summarize them as the session preflight block and continue.' 'If the user explicitly provided both choices in the current conversation, summarize them as the session preflight block and continue.' 'all four choices wording'
    $Text = Remove-RegexOnce $Text '(?ms)^### Delivery Strategy\n.*?(?=^### Chain Strategy\n|^### Dependency Graph\n)' '' 'Delivery Strategy section'
    $Text = Remove-RegexOnce $Text '(?ms)^### Chain Strategy\n.*?(?=^### Dependency Graph\n)' '' 'Chain Strategy section'
    $Text = Remove-RegexOnce $Text '(?ms)^### Review Workload Guard \(MANDATORY\)\n.*?(?=^<!-- gentle-ai:sdd-model-assignments -->\n)' '' 'Review Workload Guard section'
    $Text = Remove-ExactOnce $Text '3. The user explicitly asked to apply/continue implementation, or the prior SDD planning phase completed and the orchestrator has passed the review workload guard.' '3. The user explicitly asked to apply/continue implementation, or the prior SDD planning phase completed.' 'apply routing review-workload clause'

    foreach ($Marker in $SanitizerPolicy.required_markers) {
        if ($Text.IndexOf([string]$Marker, [System.StringComparison]::Ordinal) -lt 0) {
            Die "missing required marker after sanitizing: $Marker"
        }
    }
    foreach ($Marker in $SanitizerPolicy.forbidden_markers) {
        if ($Text.IndexOf([string]$Marker, [System.StringComparison]::Ordinal) -ge 0) {
            Die "forbidden marker still present after sanitizing: $Marker"
        }
    }

    return $Text
}

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$OverlayRoot = Split-Path -Parent $ScriptDir
$RepoRoot = Split-Path -Parent (Split-Path -Parent $OverlayRoot)
$PolicyFile = Join-Path $OverlayRoot 'policy/gentle-ai-policy.json'

if (-not (Test-Path -LiteralPath $PolicyFile)) {
    Die "Policy file not found: $PolicyFile"
}

$Policy = Get-Content -LiteralPath $PolicyFile -Raw | ConvertFrom-Json
$OpenCodeConfig = Resolve-UserPath $Policy.opencode.config_path
$BaseOrchestratorKey = [string]$Policy.opencode.base_orchestrator_key
$GeneratedDir = Resolve-UserPath $Policy.opencode.generated_orchestrators_dir
$RepoSnapshotDir = Join-Path $RepoRoot $Policy.opencode.orchestrator_snapshot_dir
$RepoSnapshotMetadataFile = Join-Path $RepoRoot $Policy.opencode.orchestrator_snapshot_metadata_file
$LocalSnapshotDir = Resolve-UserPath $Policy.opencode.local_orchestrator_snapshot_dir
$LocalProfilesConfig = Resolve-UserPath $Policy.opencode.sdd_profiles_local_config_path
$StateFile = Join-Path $RepoRoot $Policy.maintenance.state_file
$ProfileOrchPrefix = [string]$Policy.opencode.profile_orchestrator_prefix
$SddPhases = @($Policy.opencode.sdd_phases)
if ($SddPhases.Count -eq 0) {
    Die "policy.opencode.sdd_phases is empty or missing; cannot reconcile SDD profiles"
}
$SddPhasesSet = New-Object System.Collections.Generic.HashSet[string]
foreach ($P in $SddPhases) { [void]$SddPhasesSet.Add([string]$P) }

Write-Info 'Applying Gentle AI overlay policy...'

# --- Phase 1: skills pruning ---

$PrunedCount = 0
$MissingKeepSummary = New-Object System.Collections.Generic.List[string]

foreach ($TargetDirRaw in $Policy.skills.targets) {
    $TargetDir = Resolve-UserPath ([string]$TargetDirRaw)
    if (-not (Test-Path -LiteralPath $TargetDir)) {
        Write-Info "- skip missing skills dir: $TargetDir"
        continue
    }

    Write-Info "- pruning unwanted skills in $TargetDir"
    foreach ($Skill in $Policy.skills.prune) {
        $SkillPath = Join-Path $TargetDir ([string]$Skill)
        if (Test-Path -LiteralPath $SkillPath) {
            Remove-Item -LiteralPath $SkillPath -Recurse -Force
            Write-Info "  removed $Skill"
            $PrunedCount++
        }
        else {
            Write-Info "  already absent $Skill"
        }
    }

    foreach ($Skill in $Policy.skills.keep) {
        $SkillPath = Join-Path $TargetDir ([string]$Skill)
        if (-not (Test-Path -LiteralPath $SkillPath)) {
            [void]$MissingKeepSummary.Add("$TargetDir -> $Skill")
        }
    }
}

# --- Phase 2: OpenCode config ---

if (-not (Test-Path -LiteralPath $OpenCodeConfig)) {
    Write-Info "- skip missing OpenCode config: $OpenCodeConfig"
    Write-Info ''
    Write-Info 'Summary:'
    Write-Info "  skills pruned this run: $PrunedCount"
    if ($MissingKeepSummary.Count -gt 0) {
        Write-Info '  WARNING - keep skills missing (expected but absent):'
        foreach ($Entry in $MissingKeepSummary) {
            Write-Info "    - $Entry"
        }
    }
    Write-Info 'Done.'
    exit 0
}

# Validate that the config is parseable JSON before mutating anything.
try {
    $RawConfig = Get-Content -LiteralPath $OpenCodeConfig -Raw
    $Config = $RawConfig | ConvertFrom-Json
}
catch {
    Die ("OpenCode config at {0} is not valid JSON: {1}. Restore it from a backup under {2}/.gentle-ai/backups/ or re-run ``gentle-ai sync`` to regenerate it." -f $OpenCodeConfig, $_.Exception.Message, $HOME)
}

if (-not $Config.agent) {
    Die 'OpenCode config does not contain an agent map'
}

try {
    $State = (Get-Content -LiteralPath $StateFile -Raw) | ConvertFrom-Json
}
catch {
    Die ("state file at {0} is not valid JSON: {1}" -f $StateFile, $_.Exception.Message)
}

$RepoSnapshotBaselinePath = Join-Path $RepoSnapshotDir ($BaseOrchestratorKey + '.last.md')
if (-not (Test-Path -LiteralPath $RepoSnapshotBaselinePath -PathType Leaf)) {
    Die ("audited base snapshot missing for orchestrator '{0}' at {1}. Restore the committed baseline before re-running apply." -f $BaseOrchestratorKey, $RepoSnapshotBaselinePath)
}
$RepoSnapshotBaseline = (Get-Content -LiteralPath $RepoSnapshotBaselinePath -Raw).TrimEnd("`n", "`r")
$SnapshotMetadata = Read-SimpleYamlMap -Path $RepoSnapshotMetadataFile
$ExpectedMetadata = [ordered]@{
    schema_version = '1'
    snapshot_file = [System.IO.Path]::GetFileName($RepoSnapshotBaselinePath)
    snapshot_source = 'upstream-opencode-inline-asset'
    state_file = [string]$Policy.maintenance.state_file
    upstream_repo_name = [System.IO.Path]::GetFileName((Resolve-UserPath ([string]$Policy.upstream.repo_path)).TrimEnd('\', '/'))
    upstream_prompt_rel_path = [string]$Policy.upstream.orchestrator_prompt_path
    upstream_inject_source_rel_path = 'internal/components/sdd/inject.go'
    upstream_profiles_source_rel_path = 'internal/components/sdd/profiles.go'
    last_maintained_version = [string]$State.last_maintained_version
    last_maintained_tag = [string]$State.last_maintained_tag
    last_maintained_commit = [string]$State.last_maintained_commit
    last_reviewed_at = [string]$State.last_reviewed_at
    base_orchestrator_key = $BaseOrchestratorKey
    profile_orchestrator_prefix = $ProfileOrchPrefix
    profile_phase_order_csv = ($SddPhases -join ',')
    profile_task_scope_rule = 'deny-all-then-allow-suffixed-phases-and-global-jd'
}
foreach ($Entry in $ExpectedMetadata.GetEnumerator()) {
    $ActualValue = if ($SnapshotMetadata.ContainsKey($Entry.Key)) { [string]$SnapshotMetadata[$Entry.Key] } else { $null }
    if ($ActualValue -ne [string]$Entry.Value) {
        Die ("audited snapshot metadata mismatch: field '{0}' in {1} is '{2}', expected '{3}'. Repair the committed baseline before re-running apply." -f $Entry.Key, $RepoSnapshotMetadataFile, $ActualValue, [string]$Entry.Value)
    }
}
$ActualSnapshotHash = Get-Sha256Hex -Content $RepoSnapshotBaseline
if ([string]$SnapshotMetadata['snapshot_sha256'] -ne $ActualSnapshotHash) {
    Die ("audited snapshot metadata mismatch: snapshot_sha256 in {0} is '{1}', expected '{2}' from {3}. Repair the committed baseline before re-running apply." -f $RepoSnapshotMetadataFile, [string]$SnapshotMetadata['snapshot_sha256'], $ActualSnapshotHash, $RepoSnapshotBaselinePath)
}

$ConfigChanged = $false
$GeneratedCount = 0
$RecoveredCount = 0
$KeptCount = 0
$SkippedCount = 0
$RepoSnapshotNew = 0
$RepoSnapshotChanged = 0
$RepoSnapshotUnchanged = 0
$LocalSnapshotNew = 0
$LocalSnapshotChanged = 0
$LocalSnapshotUnchanged = 0
$LocalSnapshotMigrations = 0
$RepoSnapshotBackfills = 0
$TopologyWarnings = New-Object System.Collections.Generic.List[string]
$WrittenOrchestratorKeys = New-Object System.Collections.Generic.HashSet[string]

# Profile reconciliation counters
$ProfilesManagedCount = 0
$ProfileAgentsCreated = 0
$ProfileAgentsUpdated = 0
$ProfileAgentsUnchanged = 0
$UnmanagedProfileNames = New-Object System.Collections.Generic.List[string]
$ManagedProfileNames = New-Object System.Collections.Generic.HashSet[string]
$BaseRuntimePrompt = $null
$BaseGeneratedPath = $null

# Snapshot original agent keys BEFORE the override loop creates any stubs,
# so topology drift checks can tell which override targets had to be invented.
$OriginalAgentKeys = New-Object System.Collections.Generic.HashSet[string]
foreach ($Name in $Config.agent.PSObject.Properties.Name) {
    [void]$OriginalAgentKeys.Add($Name)
}

$CreatedOverrides = New-Object System.Collections.Generic.List[string]

foreach ($Override in $Policy.agent_overrides) {
    $Key = [string]$Override.key
    $Model = [string]$Override.model
    $Variant = [string]$Override.variant

    # Track EVERY non-object reset, matching bash behavior. Two cases land here:
    #   - the key did not exist upstream (purely missing)
    #   - the key existed but as a non-object (string/null/list)
    # Both deserve a topology warning so the maintainer can review whether the
    # upstream agent shape changed.
    $AgentValue = if ($OriginalAgentKeys.Contains($Key)) { $Config.agent.$Key } else { $null }
    if (-not ($AgentValue -is [PSCustomObject])) {
        $Config.agent | Add-Member -NotePropertyName $Key -NotePropertyValue ([PSCustomObject]@{}) -Force
        [void]$CreatedOverrides.Add($Key)
        Write-Info "  agent override $Key reset to object before applying model"
    }

    $CurrentModel = [string](Get-AgentPropertyValue -Agent $Config.agent.$Key -Name 'model')
    $CurrentVariant = [string](Get-AgentPropertyValue -Agent $Config.agent.$Key -Name 'variant')
    if ($CurrentModel -ne $Model) {
        $Config.agent.$Key | Add-Member -NotePropertyName 'model' -NotePropertyValue $Model -Force
        $ConfigChanged = $true
    }
    if ($Variant -and $CurrentVariant -ne $Variant) {
        $Config.agent.$Key | Add-Member -NotePropertyName 'variant' -NotePropertyValue $Variant -Force
        $ConfigChanged = $true
    }

    $OverrideSuffix = if ($Variant) { " ($Variant)" } else { '' }
    Write-Info "  agent override $Key -> $Model$OverrideSuffix"
}

# --- SDD profile reconciliation (strict, fail-closed) ---
# Mirrors the bash helper. Contract:
#   - If the local config file does NOT exist: do not touch SDD profiles.
#   - If it exists: parse + validate STRICTLY before any mutation.
#   - For each managed profile: create/update orchestrator + 10 phase agents
#     with the configured model/variant. Do NOT touch prompts here.
#   - Profiles present in opencode.json but absent from local config are left
#     untouched but surfaced as warnings + counter.
#   - No automatic deletion of unmanaged profiles.

if (Test-Path -LiteralPath $LocalProfilesConfig -PathType Leaf) {
    try {
        $LocalCfgRaw = Get-Content -LiteralPath $LocalProfilesConfig -Raw
        $LocalCfg = $LocalCfgRaw | ConvertFrom-Json
    }
    catch {
        Die ("local SDD profile config at {0} is not valid JSON: {1}. Fix or remove the file before re-running this script." -f $LocalProfilesConfig, $_.Exception.Message)
    }

    if (-not ($LocalCfg -is [PSCustomObject])) {
        Die ("local SDD profile config at {0} must be a JSON object at the top level" -f $LocalProfilesConfig)
    }
    $AllowedLocalTopFields = @('version', 'profiles')
    foreach ($PropName in $LocalCfg.PSObject.Properties.Name) {
        if (-not ($AllowedLocalTopFields -contains $PropName)) {
            Die ("local SDD profile config at {0} has unexpected top-level field '{1}'; only 'version' and 'profiles' are allowed" -f $LocalProfilesConfig, $PropName)
        }
    }

    $LocalVersion = Get-AgentPropertyValue -Agent $LocalCfg -Name 'version'
    if ($LocalVersion -ne 1) {
        Die ("local SDD profile config at {0} has unsupported 'version' '{1}'; expected 1" -f $LocalProfilesConfig, $LocalVersion)
    }
    if (-not ($LocalCfg.PSObject.Properties.Name -contains 'profiles')) {
        Die ("local SDD profile config at {0} must contain a non-empty 'profiles' array" -f $LocalProfilesConfig)
    }
    $ProfilesValue = $LocalCfg.profiles
    if ($null -eq $ProfilesValue) {
        Die ("local SDD profile config at {0} must contain a non-empty 'profiles' array" -f $LocalProfilesConfig)
    }
    # Force-wrap valid arrays to preserve the single-element-array case. Reject
    # true non-arrays before any mutation.
    if ($ProfilesValue -isnot [System.Array]) {
        Die ("local SDD profile config at {0} must contain a non-empty 'profiles' array (got non-array type {1})" -f $LocalProfilesConfig, $ProfilesValue.GetType().FullName)
    }
    $ProfilesRaw = @($ProfilesValue)
    if ($ProfilesRaw.Count -eq 0) {
        Die ("local SDD profile config at {0} must contain a non-empty 'profiles' array" -f $LocalProfilesConfig)
    }

    $SeenProfileNames = New-Object System.Collections.Generic.HashSet[string]
    $ValidatedProfiles = New-Object System.Collections.Generic.List[object]
    for ($i = 0; $i -lt $ProfilesRaw.Count; $i++) {
        $Prefix = "profiles[$i]"
        $ProfileEntry = $ProfilesRaw[$i]
        if (-not ($ProfileEntry -is [PSCustomObject])) {
            Die "$Prefix`: must be an object"
        }
        $AllowedTopFields = @('name', 'orchestrator', 'phases')
        foreach ($PropName in $ProfileEntry.PSObject.Properties.Name) {
            if (-not ($AllowedTopFields -contains $PropName)) {
                Die "$Prefix`: unexpected field '$PropName'; only 'name', 'orchestrator', 'phases' are allowed"
            }
        }
        $ProfileName = Get-AgentPropertyValue -Agent $ProfileEntry -Name 'name'
        if ($ProfileName -isnot [string] -or [string]::IsNullOrEmpty($ProfileName)) {
            Die "$Prefix`: 'name' must be a non-empty string"
        }
        if ($ProfileName -cnotmatch '^[a-z0-9][a-z0-9._-]*$') {
            Die "$Prefix`: 'name' '$ProfileName' must match ^[a-z0-9][a-z0-9._-]*$ to be safe as an agent-key suffix"
        }
        if ($SeenProfileNames.Contains($ProfileName)) {
            Die "$Prefix`: duplicate profile name '$ProfileName'"
        }
        [void]$SeenProfileNames.Add($ProfileName)

        if (-not ($ProfileEntry.PSObject.Properties.Name -contains 'orchestrator')) {
            Die "$Prefix`: missing required field 'orchestrator'"
        }
        Validate-Assignment -Label "$Prefix.orchestrator" -Value $ProfileEntry.orchestrator

        if (-not ($ProfileEntry.PSObject.Properties.Name -contains 'phases')) {
            Die "$Prefix`: missing required field 'phases'"
        }
        $PhasesObj = $ProfileEntry.phases
        if (-not ($PhasesObj -is [PSCustomObject])) {
            Die "$Prefix.phases: must be an object keyed by SDD phase name"
        }
        $PhaseKeysSet = New-Object System.Collections.Generic.HashSet[string]
        foreach ($PK in $PhasesObj.PSObject.Properties.Name) { [void]$PhaseKeysSet.Add($PK) }

        $Missing = New-Object System.Collections.Generic.List[string]
        foreach ($P in $SddPhases) {
            if (-not $PhaseKeysSet.Contains($P)) { [void]$Missing.Add($P) }
        }
        if ($Missing.Count -gt 0) {
            $MissingSorted = ($Missing | Sort-Object) -join ', '
            Die "$Prefix.phases: missing required phases [$MissingSorted] (no defaults are inherited)"
        }
        $Unknown = New-Object System.Collections.Generic.List[string]
        foreach ($PK in $PhaseKeysSet) {
            if (-not $SddPhasesSet.Contains($PK)) { [void]$Unknown.Add($PK) }
        }
        if ($Unknown.Count -gt 0) {
            $UnknownSorted = ($Unknown | Sort-Object) -join ', '
            $AllowedSorted = $SddPhases -join ', '
            Die "$Prefix.phases: unknown phases [$UnknownSorted]; allowed: [$AllowedSorted]"
        }
        foreach ($P in $SddPhases) {
            Validate-Assignment -Label "$Prefix.phases.$P" -Value $PhasesObj.$P
        }

        [void]$ValidatedProfiles.Add([PSCustomObject]@{
            Name = $ProfileName
            Orchestrator = $ProfileEntry.orchestrator
            Phases = $PhasesObj
        })
    }

    # --- All validation passed. Apply. ---
    foreach ($Vp in $ValidatedProfiles) {
        $Pname = $Vp.Name
        [void]$ManagedProfileNames.Add($Pname)
        $ProfilesManagedCount++

        # Orchestrator agent
        $OrchKey = "sdd-orchestrator-$Pname"
        $OrchAssignment = $Vp.Orchestrator
        $ExistingOrch = if ($Config.agent.PSObject.Properties.Name -contains $OrchKey) { $Config.agent.$OrchKey } else { $null }
        if (-not ($ExistingOrch -is [PSCustomObject])) {
            $NewObj = [PSCustomObject]@{ model = [string]$OrchAssignment.model; variant = [string]$OrchAssignment.variant }
            $Config.agent | Add-Member -NotePropertyName $OrchKey -NotePropertyValue $NewObj -Force
            $ProfileAgentsCreated++
            $ConfigChanged = $true
            Write-Info "  profile $Pname`: created orchestrator agent $OrchKey (no prompt; run ``gentle-ai sync`` to materialize)"
        }
        else {
            $ChangedHere = $false
            $CurModel = [string](Get-AgentPropertyValue -Agent $ExistingOrch -Name 'model')
            $CurVariant = [string](Get-AgentPropertyValue -Agent $ExistingOrch -Name 'variant')
            $WantModel = [string]$OrchAssignment.model
            $WantVariant = [string]$OrchAssignment.variant
            if ($CurModel -ne $WantModel) {
                $ExistingOrch | Add-Member -NotePropertyName 'model' -NotePropertyValue $WantModel -Force
                $ChangedHere = $true
            }
            if ($CurVariant -ne $WantVariant) {
                $ExistingOrch | Add-Member -NotePropertyName 'variant' -NotePropertyValue $WantVariant -Force
                $ChangedHere = $true
            }
            if ($ChangedHere) {
                $ProfileAgentsUpdated++
                $ConfigChanged = $true
                $Suffix = if ($WantVariant) { " ($WantVariant)" } else { '' }
                Write-Info "  profile $Pname`: updated orchestrator agent $OrchKey -> $WantModel$Suffix"
            }
            else {
                $ProfileAgentsUnchanged++
            }
        }

        # Phase agents
        foreach ($P in $SddPhases) {
            $PhaseKey = "$P-$Pname"
            $Assignment = $Vp.Phases.$P
            $ExistingPhase = if ($Config.agent.PSObject.Properties.Name -contains $PhaseKey) { $Config.agent.$PhaseKey } else { $null }
            if (-not ($ExistingPhase -is [PSCustomObject])) {
                $NewObj = [PSCustomObject]@{ model = [string]$Assignment.model; variant = [string]$Assignment.variant }
                $Config.agent | Add-Member -NotePropertyName $PhaseKey -NotePropertyValue $NewObj -Force
                $ProfileAgentsCreated++
                $ConfigChanged = $true
                $Suffix = if ([string]$Assignment.variant) { " ($([string]$Assignment.variant))" } else { '' }
                Write-Info "  profile $Pname`: created phase agent $PhaseKey -> $([string]$Assignment.model)$Suffix"
            }
            else {
                $ChangedHere = $false
                $CurModel = [string](Get-AgentPropertyValue -Agent $ExistingPhase -Name 'model')
                $CurVariant = [string](Get-AgentPropertyValue -Agent $ExistingPhase -Name 'variant')
                $WantModel = [string]$Assignment.model
                $WantVariant = [string]$Assignment.variant
                if ($CurModel -ne $WantModel) {
                    $ExistingPhase | Add-Member -NotePropertyName 'model' -NotePropertyValue $WantModel -Force
                    $ChangedHere = $true
                }
                if ($CurVariant -ne $WantVariant) {
                    $ExistingPhase | Add-Member -NotePropertyName 'variant' -NotePropertyValue $WantVariant -Force
                    $ChangedHere = $true
                }
                if ($ChangedHere) {
                    $ProfileAgentsUpdated++
                    $ConfigChanged = $true
                    $Suffix = if ($WantVariant) { " ($WantVariant)" } else { '' }
                    Write-Info "  profile $Pname`: updated phase agent $PhaseKey -> $WantModel$Suffix"
                }
                else {
                    $ProfileAgentsUnchanged++
                }
            }
        }
    }

    # Detect unmanaged profiles already present in opencode.json (warn-only).
    $DiscoveredProfileNames = New-Object System.Collections.Generic.HashSet[string]
    foreach ($K in $OriginalAgentKeys) {
        if (Is-ProfileOrchestratorKey -AgentKey $K -ProfilePrefix $ProfileOrchPrefix) {
            $Pn = Get-ProfileNameFromOrchestratorKey -AgentKey $K -ProfilePrefix $ProfileOrchPrefix
            if (-not [string]::IsNullOrEmpty($Pn)) {
                [void]$DiscoveredProfileNames.Add($Pn)
            }
        }
    }
    $UnmanagedList = New-Object System.Collections.Generic.List[string]
    foreach ($Pn in $DiscoveredProfileNames) {
        if (-not $ManagedProfileNames.Contains($Pn)) {
            [void]$UnmanagedList.Add($Pn)
        }
    }
    foreach ($Pn in ($UnmanagedList | Sort-Object)) {
        [void]$UnmanagedProfileNames.Add($Pn)
        Write-Info "  unmanaged SDD profile present in opencode.json (left untouched): $Pn"
    }
}
else {
    Write-Info "  no local SDD profile config at $LocalProfilesConfig - SDD profiles untouched"
}

# --- Topology drift checks (non-fatal warnings) ---
$KnownOrchestratorKeys = New-Object System.Collections.Generic.HashSet[string]
foreach ($K in $Policy.opencode.orchestrator_agent_keys) {
    [void]$KnownOrchestratorKeys.Add([string]$K)
}

$OrchestratorsInConfig = New-Object System.Collections.Generic.HashSet[string]
foreach ($K in $OriginalAgentKeys) {
    if (Is-OrchestratorAgent -AgentKey $K -ExactKeys $Policy.opencode.orchestrator_agent_keys -Prefixes $Policy.opencode.orchestrator_agent_prefixes) {
        [void]$OrchestratorsInConfig.Add($K)
    }
}

$UnknownList = New-Object System.Collections.Generic.List[string]
foreach ($K in $OrchestratorsInConfig) {
    if (-not $KnownOrchestratorKeys.Contains($K)) {
        # EXCEPTION: profile-managed orchestrators (sdd-orchestrator-<name>) are
        # deliberately not in orchestrator_agent_keys; they are managed via the
        # local SDD profile config. Don't warn about them.
        if (Is-ProfileOrchestratorKey -AgentKey $K -ProfilePrefix $ProfileOrchPrefix) {
            continue
        }
        [void]$UnknownList.Add($K)
    }
}
foreach ($K in ($UnknownList | Sort-Object)) {
    $Msg = "unknown orchestrator matched by prefix only: $K"
    [void]$TopologyWarnings.Add($Msg)
    Write-Info "  topology: $Msg"
}

$MissingOrchestratorList = New-Object System.Collections.Generic.List[string]
foreach ($K in $KnownOrchestratorKeys) {
    if (-not $OriginalAgentKeys.Contains($K)) {
        [void]$MissingOrchestratorList.Add($K)
    }
}
foreach ($K in ($MissingOrchestratorList | Sort-Object)) {
    $Msg = "expected orchestrator missing from opencode.json: $K"
    [void]$TopologyWarnings.Add($Msg)
    Write-Info "  topology: $Msg"
}

foreach ($K in ($CreatedOverrides | Sort-Object)) {
    $Msg = "agent_override target was missing from upstream (created): $K"
    [void]$TopologyWarnings.Add($Msg)
    Write-Info "  topology: $Msg"
}

# --- Generate orchestrator overlays ---
if (-not (Test-Path -LiteralPath $GeneratedDir)) {
    New-Item -ItemType Directory -Path $GeneratedDir -Force | Out-Null
}
if (-not (Test-Path -LiteralPath $RepoSnapshotDir)) {
    New-Item -ItemType Directory -Path $RepoSnapshotDir -Force | Out-Null
}
if (-not (Test-Path -LiteralPath $LocalSnapshotDir)) {
    New-Item -ItemType Directory -Path $LocalSnapshotDir -Force | Out-Null
}

$AgentKeys = $Config.agent.PSObject.Properties.Name | Sort-Object
foreach ($AgentKey in $AgentKeys) {
    if (-not (Is-OrchestratorAgent -AgentKey $AgentKey -ExactKeys $Policy.opencode.orchestrator_agent_keys -Prefixes $Policy.opencode.orchestrator_agent_prefixes)) {
        continue
    }

    $Agent = $Config.agent.$AgentKey
    if (-not ($Agent -is [PSCustomObject])) {
        Write-Info "  skip $AgentKey`: agent entry is not an object"
        $SkippedCount++
        continue
    }

    $PromptValue = Get-AgentPropertyValue -Agent $Agent -Name 'prompt'
    if ($PromptValue -isnot [string] -or [string]::IsNullOrWhiteSpace($PromptValue)) {
        Write-Info "  skip $AgentKey`: prompt missing or not a string"
        $SkippedCount++
        continue
    }

    $SafeKey = Assert-SafeSnapshotKey -Key $AgentKey
    $GeneratedPath = Join-Path $GeneratedDir ($SafeKey + '.overlay.md')
    $DesiredPrompt = '{file:' + $GeneratedPath + '}'
    $RepoSnapshotPath = Join-Path $RepoSnapshotDir ($SafeKey + '.last.md')
    $LocalSnapshotPath = Join-Path $LocalSnapshotDir ($SafeKey + '.last.md')

    [void](Ensure-LocalSnapshotFromRepo -AgentKey $AgentKey -RepoSnapshotPath $RepoSnapshotPath -LocalSnapshotPath $LocalSnapshotPath -MigrationCounter ([ref]$LocalSnapshotMigrations))
    [void](Ensure-RepoSnapshotFromLocal -AgentKey $AgentKey -RepoSnapshotPath $RepoSnapshotPath -LocalSnapshotPath $LocalSnapshotPath -PolicyObject $Policy -BackfillCounter ([ref]$RepoSnapshotBackfills))

    # Fully-applied state — already pointing at our generated overlay file,
    # and that file exists on disk. Nothing to do.
    if ($PromptValue -eq $DesiredPrompt -and (Test-Path -LiteralPath $GeneratedPath -PathType Leaf)) {
        if (-not (Test-Path -LiteralPath $LocalSnapshotPath -PathType Leaf)) {
            Die ("local operational snapshot missing for orchestrator '{0}' at {1}. Run ``gentle-ai sync`` to reset the orchestrator prompt to inline content, then re-run this script to capture a fresh snapshot." -f $AgentKey, $LocalSnapshotPath)
        }
        if ((Test-RepoSnapshotEligible -AgentKey $AgentKey -PolicyObject $Policy) -and -not (Test-Path -LiteralPath $RepoSnapshotPath -PathType Leaf)) {
            [void](Ensure-RepoSnapshotFromLocal -AgentKey $AgentKey -RepoSnapshotPath $RepoSnapshotPath -LocalSnapshotPath $LocalSnapshotPath -PolicyObject $Policy -BackfillCounter ([ref]$RepoSnapshotBackfills))
            if (-not (Test-Path -LiteralPath $RepoSnapshotPath -PathType Leaf)) {
                Die ("versioned repo snapshot missing for orchestrator '{0}' at {1}. Run ``gentle-ai sync`` to capture fresh upstream, then re-run this script." -f $AgentKey, $RepoSnapshotPath)
            }
        }
        Write-Info "  keep $AgentKey`: already points to generated overlay prompt"
        if ($AgentKey -eq $BaseOrchestratorKey) {
            $BaseRuntimePrompt = (Get-Content -LiteralPath $LocalSnapshotPath -Raw).TrimEnd("`n", "`r")
            $BaseGeneratedPath = $GeneratedPath
        }
        [void]$WrittenOrchestratorKeys.Add($AgentKey)
        $KeptCount++
        continue
    }

    $RecoveredFromSnapshot = $false
    $InlinePrompt = $null

    if ($PromptValue.StartsWith('{file:') -and $PromptValue.EndsWith('}')) {
        # Prompt is a file reference but either the target file is missing
        # or it points somewhere different from our desired path.
        # Recover from snapshot if available; fail loud if not.
        if (-not (Test-Path -LiteralPath $LocalSnapshotPath -PathType Leaf)) {
            [void](Ensure-LocalSnapshotFromRepo -AgentKey $AgentKey -RepoSnapshotPath $RepoSnapshotPath -LocalSnapshotPath $LocalSnapshotPath -MigrationCounter ([ref]$LocalSnapshotMigrations))
        }
        if (-not (Test-Path -LiteralPath $LocalSnapshotPath -PathType Leaf)) {
            if (Test-RepoSnapshotEligible -AgentKey $AgentKey -PolicyObject $Policy) {
                $MissingDetail = "no local operational snapshot exists at $LocalSnapshotPath and no repo snapshot exists at $RepoSnapshotPath"
            }
            else {
                $MissingDetail = "no local operational snapshot exists at $LocalSnapshotPath"
            }
            Die ("broken state for orchestrator '{0}': opencode.json prompt is '{1}' but the target file is missing and {2}. Run ``gentle-ai sync`` to reset the orchestrator prompt to inline content, then re-run this script." -f $AgentKey, $PromptValue, $MissingDetail)
        }
        if ((Test-RepoSnapshotEligible -AgentKey $AgentKey -PolicyObject $Policy) -and -not (Test-Path -LiteralPath $RepoSnapshotPath -PathType Leaf)) {
            [void](Ensure-RepoSnapshotFromLocal -AgentKey $AgentKey -RepoSnapshotPath $RepoSnapshotPath -LocalSnapshotPath $LocalSnapshotPath -PolicyObject $Policy -BackfillCounter ([ref]$RepoSnapshotBackfills))
        }
        $InlinePrompt = (Get-Content -LiteralPath $LocalSnapshotPath -Raw).TrimEnd("`n", "`r")
        $RecoveredFromSnapshot = $true
        Write-Info "  WARNING recovering $AgentKey from local snapshot - content may pre-date current upstream; run ``gentle-ai sync`` then re-run this script to capture fresh upstream into the snapshot"
    }
    else {
        $InlinePrompt = $PromptValue
    }

    # Sanitize FIRST so a failure does not leave a stale-but-overwritten
    # snapshot. The snapshot is only updated after sanitization succeeds.
    $SanitizedPrompt = Sanitize-OrchestratorPrompt -Prompt $InlinePrompt -SanitizerPolicy $Policy.sanitizer

    # Snapshot drift tracking — only when capturing fresh inline content.
    if ($RecoveredFromSnapshot) {
        $SnapshotStatus = 'recovered'
    }
    else {
        $LocalSnapshotStatus = Write-SnapshotWithStatus -Path $LocalSnapshotPath -Content $InlinePrompt -NewCounter ([ref]$LocalSnapshotNew) -ChangedCounter ([ref]$LocalSnapshotChanged) -UnchangedCounter ([ref]$LocalSnapshotUnchanged)
        if (Test-RepoSnapshotEligible -AgentKey $AgentKey -PolicyObject $Policy) {
            $RepoSnapshotStatus = Write-SnapshotWithStatus -Path $RepoSnapshotPath -Content $InlinePrompt -NewCounter ([ref]$RepoSnapshotNew) -ChangedCounter ([ref]$RepoSnapshotChanged) -UnchangedCounter ([ref]$RepoSnapshotUnchanged)
            $SnapshotStatus = "local: $LocalSnapshotStatus, repo: $RepoSnapshotStatus"
        }
        else {
            $SnapshotStatus = "local: $LocalSnapshotStatus"
        }
    }

    Write-Utf8NoBomAtomic -Path $GeneratedPath -Content (Ensure-LfTerminated $SanitizedPrompt)

    if ((Get-AgentPropertyValue -Agent $Agent -Name 'prompt') -ne $DesiredPrompt) {
        $Agent | Add-Member -NotePropertyName 'prompt' -NotePropertyValue $DesiredPrompt -Force
        $ConfigChanged = $true
    }

    if ($AgentKey -eq $BaseOrchestratorKey) {
        $BaseRuntimePrompt = $InlinePrompt
        $BaseGeneratedPath = $GeneratedPath
    }

    [void]$WrittenOrchestratorKeys.Add($AgentKey)
    if ($RecoveredFromSnapshot) {
        $RecoveredCount++
        Write-Info "  recovered $AgentKey -> $GeneratedPath (from snapshot)"
    }
    else {
        $GeneratedCount++
        Write-Info "  generated $AgentKey -> $GeneratedPath (snapshot: $SnapshotStatus)"
    }
}

# --- Atomic write of opencode.json ---
if ($ConfigChanged) {
    $Json = $Config | ConvertTo-Json -Depth 100
    # PS 5.1 ConvertTo-Json escapes non-ASCII to \uXXXX. Bash/Python writes raw
    # UTF-8 via ensure_ascii=False. Unescape here so both helpers produce
    # byte-identical opencode.json output.
    $Json = Unescape-NonAsciiUnicode $Json
    Write-Utf8NoBomAtomic -Path $OpenCodeConfig -Content (Ensure-LfTerminated $Json)
    $ConfigStatus = 'updated'

}
else {
    $ConfigStatus = 'unchanged'
}

# --- Verification from persisted opencode.json ---
$VerifyConfig = Get-Content -LiteralPath $OpenCodeConfig -Raw | ConvertFrom-Json

foreach ($Override in $Policy.agent_overrides) {
    $Key = [string]$Override.key
    $ExpectedModel = [string]$Override.model
    $ExpectedVariant = [string]$Override.variant

    if (-not ($VerifyConfig.agent.PSObject.Properties.Name -contains $Key)) {
        Die "post-write verification failed: agent '$Key' is missing from $OpenCodeConfig after write"
    }
    $ActualAgent = $VerifyConfig.agent.$Key
    $ActualModel = [string](Get-AgentPropertyValue -Agent $ActualAgent -Name 'model')
    if ($ActualModel -ne $ExpectedModel) {
        Die "post-write verification failed: agent '$Key' model is '$ActualModel' after write, expected '$ExpectedModel'"
    }
    if ($ExpectedVariant) {
        $ActualVariant = [string](Get-AgentPropertyValue -Agent $ActualAgent -Name 'variant')
        if ($ActualVariant -ne $ExpectedVariant) {
            Die "post-write verification failed: agent '$Key' variant is '$ActualVariant' after write, expected '$ExpectedVariant'"
        }
    }
}

foreach ($Key in ($WrittenOrchestratorKeys | Sort-Object)) {
    $ExpectedRef = '{file:' + (Join-Path $GeneratedDir ($Key + '.overlay.md')) + '}'
    if (-not ($VerifyConfig.agent.PSObject.Properties.Name -contains $Key)) {
        Die "post-write verification failed: orchestrator '$Key' is missing from $OpenCodeConfig after write"
    }
    $ActualPrompt = [string](Get-AgentPropertyValue -Agent $VerifyConfig.agent.$Key -Name 'prompt')
    if ($ActualPrompt -ne $ExpectedRef) {
        Die "post-write verification failed: orchestrator '$Key' prompt is '$ActualPrompt' after write, expected '$ExpectedRef'"
    }
    $OverlayPath = Join-Path $GeneratedDir ($Key + '.overlay.md')
    if (-not (Test-Path -LiteralPath $OverlayPath -PathType Leaf)) {
        Die "post-write verification failed: overlay file missing for '$Key' at $OverlayPath"
    }
}

# Profile reconciliation post-write verification.
foreach ($Pname in ($ManagedProfileNames | Sort-Object)) {
    $OrchKey = "sdd-orchestrator-$Pname"
    if (-not ($VerifyConfig.agent.PSObject.Properties.Name -contains $OrchKey)) {
        Die "post-write verification failed: profile '$Pname' orchestrator agent '$OrchKey' missing from $OpenCodeConfig after write"
    }
    foreach ($P in $SddPhases) {
        $PhaseKey = "$P-$Pname"
        if (-not ($VerifyConfig.agent.PSObject.Properties.Name -contains $PhaseKey)) {
            Die "post-write verification failed: profile '$Pname' phase agent '$PhaseKey' missing from $OpenCodeConfig after write"
        }
    }
}

if ($null -eq $BaseRuntimePrompt -or [string]::IsNullOrEmpty([string]$BaseGeneratedPath)) {
    Die ("audited baseline verification failed: orchestrator '{0}' was not materialized during apply. Run ``gentle-ai sync`` to restore the inline upstream prompt, then re-run this script." -f $BaseOrchestratorKey)
}

if ((Normalize-Lf (Ensure-LfTerminated $BaseRuntimePrompt)) -ne (Normalize-Lf (Ensure-LfTerminated $RepoSnapshotBaseline))) {
    Die ("audited baseline mismatch for orchestrator '{0}': runtime source prompt does not match {1}. Run ``bash audit-gentle-ai-upstream.sh`` before adopting a new upstream baseline, then re-run ``gentle-ai sync`` and this script." -f $BaseOrchestratorKey, $RepoSnapshotBaselinePath)
}

$ExpectedBaseOverlay = Sanitize-OrchestratorPrompt -Prompt $RepoSnapshotBaseline -SanitizerPolicy $Policy.sanitizer
$ActualBaseOverlay = (Get-Content -LiteralPath $BaseGeneratedPath -Raw).TrimEnd("`n", "`r")
if ((Normalize-Lf (Ensure-LfTerminated $ActualBaseOverlay)) -ne (Normalize-Lf (Ensure-LfTerminated $ExpectedBaseOverlay))) {
    Die ("audited baseline mismatch for orchestrator '{0}': generated overlay at {1} does not match the sanitized audited snapshot. Re-run apply after restoring the audited baseline, or run ``gentle-ai sync`` if local runtime state is stale." -f $BaseOrchestratorKey, $BaseGeneratedPath)
}

Write-Info ''
Write-Info 'Summary:'
Write-Info "  OpenCode config status: $ConfigStatus"
Write-Info "  skills pruned this run: $PrunedCount"
Write-Info "  orchestrators generated (fresh): $GeneratedCount"
Write-Info "  orchestrators recovered from snapshot: $RecoveredCount"
Write-Info "  orchestrators kept (already applied): $KeptCount"
Write-Info "  orchestrators skipped: $SkippedCount"
Write-Info "  repo snapshots - new: $RepoSnapshotNew, changed: $RepoSnapshotChanged, unchanged: $RepoSnapshotUnchanged"
Write-Info "  local snapshots - new: $LocalSnapshotNew, changed: $LocalSnapshotChanged, unchanged: $LocalSnapshotUnchanged"
Write-Info "  local snapshot migrations from repo: $LocalSnapshotMigrations"
Write-Info "  repo snapshot backfills from local: $RepoSnapshotBackfills"
Write-Info "  topology warnings: $($TopologyWarnings.Count)"
Write-Info "  SDD profiles managed: $ProfilesManagedCount"
Write-Info "  SDD profile agents created: $ProfileAgentsCreated"
Write-Info "  SDD profile agents updated: $ProfileAgentsUpdated"
Write-Info "  SDD profile agents unchanged: $ProfileAgentsUnchanged"
Write-Info "  SDD profiles unmanaged (present in opencode.json, absent from local config): $($UnmanagedProfileNames.Count)"
Write-Info '  audited base baseline verification: ok'

if ($UnmanagedProfileNames.Count -gt 0) {
    Write-Info ''
    Write-Info 'WARNING - unmanaged SDD profiles left untouched (add them to the local SDD profile config to manage):'
    foreach ($Entry in $UnmanagedProfileNames) {
        Write-Info "  - $Entry"
    }
}

if ($MissingKeepSummary.Count -gt 0) {
    Write-Info ''
    Write-Info 'WARNING - keep skills missing (expected but absent):'
    foreach ($Entry in $MissingKeepSummary) {
        Write-Info "  - $Entry"
    }
}

if ($RepoSnapshotChanged -gt 0) {
    Write-Info ''
    Write-Info 'NOTE: versioned orchestrator snapshots drifted. Review with:'
    Write-Info '  git diff overlay/gentle-ai/snapshots/'
}

if ($LocalSnapshotChanged -gt 0) {
    Write-Info ''
    Write-Info 'NOTE: local operational orchestrator snapshots drifted under:'
    Write-Info "  $LocalSnapshotDir"
}

if ($LocalSnapshotMigrations -gt 0) {
    Write-Info ''
    Write-Info "NOTE: migrated $LocalSnapshotMigrations legacy snapshot(s) from the repo into the local operational snapshot dir."
}

if ($RepoSnapshotBackfills -gt 0) {
    Write-Info ''
    Write-Info "NOTE: backfilled $RepoSnapshotBackfills versioned repo snapshot(s) from local operational snapshots."
}

if ($RecoveredCount -gt 0) {
    Write-Info ''
    Write-Info "NOTE: $RecoveredCount orchestrator(s) recovered from snapshot."
    Write-Info '  The snapshot content may pre-date the current upstream version.'
    Write-Info '  Run `gentle-ai sync` then re-run this script to capture fresh upstream.'
}

if ($TopologyWarnings.Count -gt 0) {
    Write-Info ''
    Write-Info 'NOTE: topology drift detected. Review the topology: warnings above and update policy/intent if needed.'
}

Write-Info ''
Write-Info 'Done. Restart OpenCode if opencode.json changed.'
