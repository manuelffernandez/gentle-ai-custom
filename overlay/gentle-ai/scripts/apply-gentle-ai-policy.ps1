Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Write-Info {
    param([string]$Message)
    Write-Host $Message
}

function Resolve-UserPath {
    param([string]$PathValue)

    if ([string]::IsNullOrWhiteSpace($PathValue)) {
        return $PathValue
    }

    if ($PathValue.StartsWith('~/') -or $PathValue.StartsWith('~\')) {
        return Join-Path $HOME $PathValue.Substring(2)
    }

    if ($PathValue -eq '~') {
        return $HOME
    }

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

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$OverlayRoot = Split-Path -Parent $ScriptDir
$RepoRoot = Split-Path -Parent (Split-Path -Parent $OverlayRoot)
$PolicyFile = Join-Path $OverlayRoot 'policy/gentle-ai-policy.json'

if (-not (Test-Path -LiteralPath $PolicyFile)) {
    throw "Policy file not found: $PolicyFile"
}

$Policy = Get-Content -LiteralPath $PolicyFile -Raw | ConvertFrom-Json

$DerivedPrompt = Join-Path $RepoRoot $Policy.orchestrator.derived_prompt
$SnapshotFile = Join-Path $RepoRoot $Policy.orchestrator.snapshot_file
$OpenCodeConfig = Resolve-UserPath $Policy.orchestrator.opencode_config
$AgentKey = [string]$Policy.orchestrator.agent_key

if (-not (Test-Path -LiteralPath $DerivedPrompt)) {
    throw "Derived prompt not found: $DerivedPrompt"
}

if (-not (Test-Path -LiteralPath $OpenCodeConfig)) {
    throw "OpenCode config not found: $OpenCodeConfig"
}

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

$Config = Get-Content -LiteralPath $OpenCodeConfig -Raw | ConvertFrom-Json
if (-not $Config.agent) {
    throw 'OpenCode config does not contain an agent map'
}

$Agent = $Config.agent.$AgentKey
if (-not $Agent) {
    throw "OpenCode config is missing agent '$AgentKey'"
}

$DesiredPrompt = '{file:' + $DerivedPrompt + '}'
$CurrentPrompt = $Agent.prompt
$SnapshotStatus = 'unchanged'
$ConfigStatus = 'unchanged'

function Get-PromptContentForSnapshot {
    param(
        [AllowNull()]
        [object]$PromptValue,
        [string]$DesiredValue
    )

    if ($PromptValue -isnot [string]) {
        return $null
    }

    if ($PromptValue -eq $DesiredValue) {
        return $null
    }

    if ($PromptValue.StartsWith('{file:') -and $PromptValue.EndsWith('}')) {
        $Candidate = $PromptValue.Substring(6, $PromptValue.Length - 7)
        $Candidate = Resolve-UserPath $Candidate
        if (-not [System.IO.Path]::IsPathRooted($Candidate)) {
            $Candidate = [System.IO.Path]::GetFullPath($Candidate)
        }
        if (Test-Path -LiteralPath $Candidate -PathType Leaf) {
            return Get-Content -LiteralPath $Candidate -Raw
        }
        return $null
    }

    return [string]$PromptValue
}

$SnapshotContent = Get-PromptContentForSnapshot -PromptValue $CurrentPrompt -DesiredValue $DesiredPrompt
if (-not [string]::IsNullOrWhiteSpace($SnapshotContent)) {
    if (-not $SnapshotContent.EndsWith("`n")) {
        $SnapshotContent += "`n"
    }
    Write-Utf8NoBomAtomic -Path $SnapshotFile -Content $SnapshotContent
    $SnapshotStatus = 'updated'
}

$ConfigChanged = $false

foreach ($Override in $Policy.agent_overrides) {
    $Key = [string]$Override.key
    $Model = [string]$Override.model
    $Variant = [string]$Override.variant
    if (-not ($Config.agent.$Key -is [PSCustomObject])) {
        $Config.agent | Add-Member -NotePropertyName $Key -NotePropertyValue ([PSCustomObject]@{}) -Force
    }
    $Config.agent.$Key | Add-Member -NotePropertyName 'model' -NotePropertyValue $Model -Force
    if ($Variant) {
        $Config.agent.$Key | Add-Member -NotePropertyName 'variant' -NotePropertyValue $Variant -Force
    }
    $ConfigChanged = $true
    $OverrideSuffix = if ($Variant) { " ($Variant)" } else { '' }
    Write-Info "  agent override $Key -> $Model$OverrideSuffix"
}

if ($CurrentPrompt -ne $DesiredPrompt) {
    $Agent.prompt = $DesiredPrompt
    $ConfigChanged = $true
}

if ($ConfigChanged) {
    $Json = $Config | ConvertTo-Json -Depth 100
    Write-Utf8NoBomAtomic -Path $OpenCodeConfig -Content ($Json + "`n")
    $ConfigStatus = 'updated'
}

Write-Info "- upstream prompt snapshot: $SnapshotStatus"
Write-Info "- OpenCode prompt redirect: $ConfigStatus"
Write-Info "- desired prompt reference: $DesiredPrompt"
Write-Info 'Done. Restart OpenCode if opencode.json changed.'
