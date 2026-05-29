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
        }
        else {
            Write-Info "  already absent $Skill"
        }
    }

    $MissingKeep = New-Object System.Collections.Generic.List[string]
    foreach ($Skill in $Policy.skills.keep) {
        $SkillPath = Join-Path $TargetDir ([string]$Skill)
        if (-not (Test-Path -LiteralPath $SkillPath)) {
            [void]$MissingKeep.Add([string]$Skill)
        }
    }

    if ($MissingKeep.Count -gt 0) {
        Write-Info ("  warning: keep skills missing in {0}: {1}" -f $TargetDir, ($MissingKeep -join ', '))
    }
}

if (-not (Test-Path -LiteralPath $OpenCodeConfig)) {
    Write-Info "- skip missing OpenCode config: $OpenCodeConfig"
    Write-Info 'Done.'
    exit 0
}

$Config = Get-Content -LiteralPath $OpenCodeConfig -Raw | ConvertFrom-Json
if (-not $Config.agent) {
    throw 'OpenCode config does not contain an agent map'
}

$ConfigChanged = $false
$GeneratedCount = 0
$SkippedCount = 0

foreach ($Override in $Policy.agent_overrides) {
    $Key = [string]$Override.key
    $Model = [string]$Override.model
    $Variant = [string]$Override.variant
    if (-not ($Config.agent.$Key -is [PSCustomObject])) {
        $Config.agent | Add-Member -NotePropertyName $Key -NotePropertyValue ([PSCustomObject]@{}) -Force
        Write-Info "  agent override $Key reset to object before applying model"
    }

    $CurrentModel = [string]$Config.agent.$Key.model
    $CurrentVariant = [string]$Config.agent.$Key.variant
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

if (-not (Test-Path -LiteralPath $GeneratedDir)) {
    New-Item -ItemType Directory -Path $GeneratedDir -Force | Out-Null
}
if (-not (Test-Path -LiteralPath $SnapshotDir)) {
    New-Item -ItemType Directory -Path $SnapshotDir -Force | Out-Null
}

$AgentKeys = $Config.agent.PSObject.Properties.Name
foreach ($AgentKey in $AgentKeys) {
    if (-not (Is-OrchestratorAgent -AgentKey $AgentKey -ExactKeys $Policy.opencode.orchestrator_agent_keys -Prefixes $Policy.opencode.orchestrator_agent_prefixes)) {
        continue
    }

    $Agent = $Config.agent.$AgentKey
    if (-not ($Agent -is [PSCustomObject])) {
        Write-Info "  skip $AgentKey: agent entry is not an object"
        $SkippedCount++
        continue
    }

    $PromptValue = $Agent.prompt
    if ($PromptValue -isnot [string] -or [string]::IsNullOrWhiteSpace($PromptValue)) {
        Write-Info "  skip $AgentKey: prompt missing or not a string"
        $SkippedCount++
        continue
    }

    $GeneratedPath = Join-Path $GeneratedDir ($AgentKey + '.overlay.md')
    $DesiredPrompt = '{file:' + $GeneratedPath + '}'
    $SnapshotPath = Join-Path $SnapshotDir ($AgentKey + '.last.md')

    if ($PromptValue -eq $DesiredPrompt -and (Test-Path -LiteralPath $GeneratedPath -PathType Leaf)) {
        Write-Info "  keep $AgentKey: already points to generated overlay prompt"
        continue
    }

    if ($PromptValue.StartsWith('{file:') -and $PromptValue.EndsWith('}')) {
        Write-Info "  skip $AgentKey: prompt is external file ref and no inline content is available"
        $SkippedCount++
        continue
    }

    Write-Utf8NoBomAtomic -Path $SnapshotPath -Content (Ensure-LfTerminated $PromptValue)
    $SanitizedPrompt = Sanitize-OrchestratorPrompt -Prompt $PromptValue -SanitizerPolicy $Policy.sanitizer
    Write-Utf8NoBomAtomic -Path $GeneratedPath -Content (Ensure-LfTerminated $SanitizedPrompt)

    $Agent.prompt = $DesiredPrompt
    $ConfigChanged = $true
    $GeneratedCount++
    Write-Info "  generated $AgentKey -> $GeneratedPath"
}

if ($ConfigChanged) {
    $Json = $Config | ConvertTo-Json -Depth 100
    Write-Utf8NoBomAtomic -Path $OpenCodeConfig -Content (Ensure-LfTerminated $Json)
    $ConfigStatus = 'updated'
}
else {
    $ConfigStatus = 'unchanged'
}

Write-Info "- OpenCode config status: $ConfigStatus"
Write-Info "- generated orchestrator prompts: $GeneratedCount"
Write-Info "- skipped orchestrator prompts: $SkippedCount"
Write-Info 'Done. Restart OpenCode if opencode.json changed.'
