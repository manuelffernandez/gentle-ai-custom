#Requires -Version 5.1
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Write-Info {
    param([string]$Message)
    Write-Host $Message
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

function Remove-ExactOnce {
    param(
        [string]$Text,
        [string]$Old,
        [string]$New,
        [string]$Label
    )

    if ($Text.IndexOf($Old, [System.StringComparison]::Ordinal) -lt 0) {
        throw "Missing expected text: $Label"
    }

    return $Text.Replace($Old, $New)
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
        throw "Missing expected block: $Label"
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

function Sanitize-OrchestratorPrompt {
    param(
        [string]$Prompt,
        [object]$SanitizerPolicy
    )

    foreach ($Marker in $SanitizerPolicy.required_markers) {
        if ($Prompt.IndexOf([string]$Marker, [System.StringComparison]::Ordinal) -lt 0) {
            throw "Missing required marker before sanitizing: $Marker"
        }
    }

    $Text = $Prompt
    $Text = Remove-ExactOnce $Text '3. **Chained PR strategy**: `auto-forecast`, `ask-always`, `single-pr-default`, or `force-chained`.' '' 'preflight PR choice item'
    $Text = Remove-ExactOnce $Text '4. **Review budget**: maximum changed lines before stopping for reviewer-burden approval.' '' 'preflight review choice item'
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
    $Text = Remove-ExactOnce $Text '3. The user explicitly asked to apply/continue implementation, or the prior SDD planning phase completed and the orchestrator has passed the review workload guard.' '3. The user explicitly asked to apply/continue implementation, or the prior SDD planning phase completed.' 'apply routing clause'

    foreach ($Marker in $SanitizerPolicy.required_markers) {
        if ($Text.IndexOf([string]$Marker, [System.StringComparison]::Ordinal) -lt 0) {
            throw "Missing required marker after sanitizing: $Marker"
        }
    }
    foreach ($Marker in $SanitizerPolicy.forbidden_markers) {
        if ($Text.IndexOf([string]$Marker, [System.StringComparison]::Ordinal) -ge 0) {
            throw "Forbidden marker still present after sanitizing: $Marker"
        }
    }

    return $Text
}

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$OverlayRoot = Split-Path -Parent $ScriptDir
$RepoRoot = Split-Path -Parent (Split-Path -Parent $OverlayRoot)
$PolicyFile = Join-Path $OverlayRoot 'policy/gentle-ai-policy.json'

if (-not (Test-Path -LiteralPath $PolicyFile)) {
    throw "Policy file not found: $PolicyFile"
}

$Policy = Get-Content -LiteralPath $PolicyFile -Raw | ConvertFrom-Json
$OpenCodeConfig = Resolve-UserPath $Policy.opencode.config_path
$GeneratedDir = Resolve-UserPath $Policy.opencode.generated_orchestrators_dir
$SnapshotDir = Join-Path $RepoRoot $Policy.opencode.orchestrator_snapshot_dir

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
    throw ("OpenCode config at {0} is not valid JSON: {1}. Restore it from a backup under {2}/.gentle-ai/backups/ or re-run ``gentle-ai sync`` to regenerate it." -f $OpenCodeConfig, $_.Exception.Message, $HOME)
}

if (-not $Config.agent) {
    throw 'OpenCode config does not contain an agent map'
}

$ConfigChanged = $false
$GeneratedCount = 0
$RecoveredCount = 0
$SkippedCount = 0
$SnapshotNew = 0
$SnapshotChanged = 0
$SnapshotUnchanged = 0
$TopologyWarnings = New-Object System.Collections.Generic.List[string]
$WrittenOrchestratorKeys = New-Object System.Collections.Generic.HashSet[string]

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

    $AgentExists = $OriginalAgentKeys.Contains($Key)
    if (-not ($Config.agent.$Key -is [PSCustomObject])) {
        $Config.agent | Add-Member -NotePropertyName $Key -NotePropertyValue ([PSCustomObject]@{}) -Force
        if (-not $AgentExists) {
            [void]$CreatedOverrides.Add($Key)
        }
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
if (-not (Test-Path -LiteralPath $SnapshotDir)) {
    New-Item -ItemType Directory -Path $SnapshotDir -Force | Out-Null
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

    $GeneratedPath = Join-Path $GeneratedDir ($AgentKey + '.overlay.md')
    $DesiredPrompt = '{file:' + $GeneratedPath + '}'
    $SnapshotPath = Join-Path $SnapshotDir ($AgentKey + '.last.md')

    # Fully-applied state — already pointing at our generated overlay file,
    # and that file exists on disk. Nothing to do.
    if ($PromptValue -eq $DesiredPrompt -and (Test-Path -LiteralPath $GeneratedPath -PathType Leaf)) {
        Write-Info "  keep $AgentKey`: already points to generated overlay prompt"
        [void]$WrittenOrchestratorKeys.Add($AgentKey)
        continue
    }

    $RecoveredFromSnapshot = $false
    $InlinePrompt = $null

    if ($PromptValue.StartsWith('{file:') -and $PromptValue.EndsWith('}')) {
        # Prompt is a file reference but either the target file is missing
        # or it points somewhere different from our desired path.
        # Recover from snapshot if available; fail loud if not.
        if (-not (Test-Path -LiteralPath $SnapshotPath -PathType Leaf)) {
            throw ("broken state for orchestrator '{0}': opencode.json prompt is '{1}' but the target file is missing and no snapshot exists at {2}. Run ``gentle-ai sync`` to reset the orchestrator prompt to inline content, then re-run this script." -f $AgentKey, $PromptValue, $SnapshotPath)
        }
        $InlinePrompt = (Get-Content -LiteralPath $SnapshotPath -Raw).TrimEnd("`n", "`r")
        $RecoveredFromSnapshot = $true
        Write-Info "  recovering $AgentKey from snapshot (target file missing or path drift)"
    }
    else {
        $InlinePrompt = $PromptValue
    }

    # Snapshot drift tracking — only when capturing fresh inline content.
    if ($RecoveredFromSnapshot) {
        $SnapshotStatus = 'recovered'
    }
    else {
        $Normalized = Ensure-LfTerminated $InlinePrompt
        if (Test-Path -LiteralPath $SnapshotPath -PathType Leaf) {
            $OldSnapshotRaw = Get-Content -LiteralPath $SnapshotPath -Raw
            if ((Normalize-Lf $OldSnapshotRaw) -ne (Normalize-Lf $Normalized)) {
                $SnapshotStatus = 'changed'
                $SnapshotChanged++
            }
            else {
                $SnapshotStatus = 'unchanged'
                $SnapshotUnchanged++
            }
        }
        else {
            $SnapshotStatus = 'new'
            $SnapshotNew++
        }
        Write-Utf8NoBomAtomic -Path $SnapshotPath -Content $Normalized
    }

    $SanitizedPrompt = Sanitize-OrchestratorPrompt -Prompt $InlinePrompt -SanitizerPolicy $Policy.sanitizer
    Write-Utf8NoBomAtomic -Path $GeneratedPath -Content (Ensure-LfTerminated $SanitizedPrompt)

    if ((Get-AgentPropertyValue -Agent $Agent -Name 'prompt') -ne $DesiredPrompt) {
        $Agent | Add-Member -NotePropertyName 'prompt' -NotePropertyValue $DesiredPrompt -Force
        $ConfigChanged = $true
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

# --- Atomic write of opencode.json + post-write verification ---
if ($ConfigChanged) {
    $Json = $Config | ConvertTo-Json -Depth 100
    Write-Utf8NoBomAtomic -Path $OpenCodeConfig -Content (Ensure-LfTerminated $Json)
    $ConfigStatus = 'updated'

    $VerifyConfig = Get-Content -LiteralPath $OpenCodeConfig -Raw | ConvertFrom-Json

    foreach ($Override in $Policy.agent_overrides) {
        $Key = [string]$Override.key
        $ExpectedModel = [string]$Override.model
        $ExpectedVariant = [string]$Override.variant

        if (-not ($VerifyConfig.agent.PSObject.Properties.Name -contains $Key)) {
            throw "post-write verification failed: agent '$Key' is missing from $OpenCodeConfig after write"
        }
        $ActualAgent = $VerifyConfig.agent.$Key
        $ActualModel = [string](Get-AgentPropertyValue -Agent $ActualAgent -Name 'model')
        if ($ActualModel -ne $ExpectedModel) {
            throw "post-write verification failed: agent '$Key' model is '$ActualModel' after write, expected '$ExpectedModel'"
        }
        if ($ExpectedVariant) {
            $ActualVariant = [string](Get-AgentPropertyValue -Agent $ActualAgent -Name 'variant')
            if ($ActualVariant -ne $ExpectedVariant) {
                throw "post-write verification failed: agent '$Key' variant is '$ActualVariant' after write, expected '$ExpectedVariant'"
            }
        }
    }

    foreach ($Key in ($WrittenOrchestratorKeys | Sort-Object)) {
        $ExpectedRef = '{file:' + (Join-Path $GeneratedDir ($Key + '.overlay.md')) + '}'
        if (-not ($VerifyConfig.agent.PSObject.Properties.Name -contains $Key)) {
            throw "post-write verification failed: orchestrator '$Key' is missing from $OpenCodeConfig after write"
        }
        $ActualPrompt = [string](Get-AgentPropertyValue -Agent $VerifyConfig.agent.$Key -Name 'prompt')
        if ($ActualPrompt -ne $ExpectedRef) {
            throw "post-write verification failed: orchestrator '$Key' prompt is '$ActualPrompt' after write, expected '$ExpectedRef'"
        }
        $OverlayPath = Join-Path $GeneratedDir ($Key + '.overlay.md')
        if (-not (Test-Path -LiteralPath $OverlayPath -PathType Leaf)) {
            throw "post-write verification failed: overlay file missing for '$Key' at $OverlayPath"
        }
    }
}
else {
    $ConfigStatus = 'unchanged'
}

Write-Info ''
Write-Info 'Summary:'
Write-Info "  OpenCode config status: $ConfigStatus"
Write-Info "  skills pruned this run: $PrunedCount"
Write-Info "  orchestrators generated (fresh): $GeneratedCount"
Write-Info "  orchestrators recovered from snapshot: $RecoveredCount"
Write-Info "  orchestrators skipped: $SkippedCount"
Write-Info "  snapshots - new: $SnapshotNew, changed: $SnapshotChanged, unchanged: $SnapshotUnchanged"
Write-Info "  topology warnings: $($TopologyWarnings.Count)"

if ($MissingKeepSummary.Count -gt 0) {
    Write-Info ''
    Write-Info 'WARNING - keep skills missing (expected but absent):'
    foreach ($Entry in $MissingKeepSummary) {
        Write-Info "  - $Entry"
    }
}

if ($SnapshotChanged -gt 0) {
    Write-Info ''
    Write-Info 'NOTE: upstream orchestrator prompts drifted. Review with:'
    Write-Info '  git diff overlay/gentle-ai/snapshots/'
}

if ($TopologyWarnings.Count -gt 0) {
    Write-Info ''
    Write-Info 'NOTE: topology drift detected. Review the topology: warnings above and update policy/intent if needed.'
}

Write-Info ''
Write-Info 'Done. Restart OpenCode if opencode.json changed.'
