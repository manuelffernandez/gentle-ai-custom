#Requires -Version 5.1
[CmdletBinding()]
param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$Targets
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$SourceDir   = $PSScriptRoot
$SharedDir   = Join-Path $SourceDir 'shared'
$SharedSkill = Join-Path $SharedDir 'skills\commit-planner\SKILL.md'
$PlanBody    = Join-Path $SharedDir 'commands\commit-plan-body.md'
$ApplyBody   = Join-Path $SharedDir 'commands\commit-apply-body.md'

$SupportedTargets = @('opencode', 'claude', 'codex')

# UTF-8 without BOM — consistent with the bash script output
$Utf8NoBom = New-Object System.Text.UTF8Encoding $false

function Show-Usage {
    $name = [System.IO.Path]::GetFileName($PSCommandPath)
    Write-Host "Usage: $name all | [opencode|claude|codex ...]"
    Write-Host 'Examples:'
    Write-Host "  $name opencode"
    Write-Host "  $name claude codex"
    Write-Host "  $name all"
}

function Exit-WithError([string]$message) {
    [Console]::Error.WriteLine($message)
    exit 1
}

function Assert-SourceFile([string]$path) {
    if (-not (Test-Path $path -PathType Leaf)) {
        Exit-WithError "Missing source: $path"
    }
}

function Assert-Sources {
    Assert-SourceFile $SharedSkill
    Assert-SourceFile $PlanBody
    Assert-SourceFile $ApplyBody
}

function Resolve-Targets([string[]]$input) {
    if ($input.Count -eq 0) {
        Show-Usage
        exit 1
    }

    if ($input.Count -eq 1 -and $input[0] -eq 'all') {
        return $SupportedTargets
    }

    $result = [System.Collections.Generic.List[string]]::new()
    foreach ($t in $input) {
        if ($t -eq 'all') {
            Exit-WithError "Use 'all' by itself, or pass explicit targets only."
        }
        if ($SupportedTargets -notcontains $t) {
            Exit-WithError "Unknown target: $t"
        }
        $result.Add($t)
    }
    return , $result.ToArray()
}

function Install-Skill([string]$targetDir) {
    $dest = Join-Path $targetDir 'skills\commit-planner'
    New-Item -ItemType Directory -Force -Path $dest | Out-Null
    Copy-Item $SharedSkill (Join-Path $dest 'SKILL.md') -Force
}

function Write-RenderedFile([string]$path, [string]$content) {
    $dir = [System.IO.Path]::GetDirectoryName($path)
    New-Item -ItemType Directory -Force -Path $dir | Out-Null
    [System.IO.File]::WriteAllText($path, $content, $Utf8NoBom)
}

function Render-OpenCodeCommand([string]$targetFile, [string]$mode, [string]$cmdType, [string]$desc, [string]$bodyFile) {
    $body = [System.IO.File]::ReadAllText($bodyFile, $Utf8NoBom)
    $lines = @(
        '---',
        "description: $desc",
        'agent: gentleman',
        '---',
        '',
        'Read the skill file at `~/.config/opencode/skills/commit-planner/SKILL.md` FIRST, then follow it exactly.',
        '',
        'CONTEXT:',
        '- Working directory: !`echo -n "$(pwd)"`',
        '- Current project: !`echo -n "$(basename "$(pwd)")"`',
        "- Mode: $mode",
        "- Command type: $cmdType",
        ''
    )
    Write-RenderedFile $targetFile (($lines -join "`n") + $body)
}

function Render-ClaudeCommand([string]$targetFile, [string]$mode, [string]$cmdType, [string]$desc, [string]$bodyFile) {
    $body = [System.IO.File]::ReadAllText($bodyFile, $Utf8NoBom)
    $lines = [System.Collections.Generic.List[string]]@(
        '---',
        "description: $desc",
        'argument-hint: [optional-context]',
        'allowed-tools:',
        '  - Read',
        '  - Glob',
        '  - Bash(git:*)',
        '  - Bash(pwd:*)',
        '  - Bash(basename:*)'
    )
    if ($mode -eq 'apply') {
        $lines.Add('disable-model-invocation: true')
    }
    $lines.AddRange([string[]]@(
        '---',
        '',
        'Read the skill file at `~/.claude/skills/commit-planner/SKILL.md` FIRST, then follow it exactly.',
        '',
        'CONTEXT:',
        '- Working directory: !`pwd`',
        '- Current project: !`basename "$PWD"`',
        "- Mode: $mode",
        "- Command type: $cmdType",
        ''
    ))
    Write-RenderedFile $targetFile (($lines -join "`n") + $body)
}

function Render-CodexPrompt([string]$targetFile, [string]$mode, [string]$cmdType, [string]$desc, [string]$bodyFile) {
    $body = [System.IO.File]::ReadAllText($bodyFile, $Utf8NoBom)
    $lines = @(
        '---',
        "description: $desc",
        'argument-hint: [optional-context]',
        'allowed-tools:',
        '  - Read',
        '  - Glob',
        '  - Bash(git:*)',
        '  - Bash(pwd:*)',
        '  - Bash(basename:*)',
        '---',
        '',
        'Read the skill file at `~/.codex/skills/commit-planner/SKILL.md` FIRST, then follow it exactly.',
        '',
        'CONTEXT:',
        '- Working directory: !`pwd`',
        '- Current project: !`basename "$PWD"`',
        "- Mode: $mode",
        "- Command type: $cmdType",
        ''
    )
    Write-RenderedFile $targetFile (($lines -join "`n") + $body)
}

function Apply-OpenCode {
    # On Windows, OpenCode may store config under $env:APPDATA\opencode instead.
    # Adjust $targetDir below if needed.
    $targetDir = Join-Path $HOME '.config\opencode'
    Install-Skill $targetDir
    Render-OpenCodeCommand `
        (Join-Path $targetDir 'commands\commit-plan.md') `
        'plan' 'read-only' `
        'Propose a post-SDD commit plan without changing git state' `
        $PlanBody
    Render-OpenCodeCommand `
        (Join-Path $targetDir 'commands\commit-apply.md') `
        'apply' 'state-changing' `
        'Execute an approved post-SDD commit plan, or generate one first if missing' `
        $ApplyBody
    Write-Host "Applied OpenCode overlays -> $targetDir"
}

function Apply-Claude {
    $targetDir = Join-Path $HOME '.claude'
    Install-Skill $targetDir
    Render-ClaudeCommand `
        (Join-Path $targetDir 'commands\commit-plan.md') `
        'plan' 'read-only' `
        'Propose a post-SDD commit plan without changing git state' `
        $PlanBody
    Render-ClaudeCommand `
        (Join-Path $targetDir 'commands\commit-apply.md') `
        'apply' 'state-changing' `
        'Execute an approved post-SDD commit plan, or generate one first if missing' `
        $ApplyBody
    Write-Host "Applied Claude overlays -> $targetDir"
}

function Apply-Codex {
    $targetDir = Join-Path $HOME '.codex'
    Install-Skill $targetDir
    Render-CodexPrompt `
        (Join-Path $targetDir 'prompts\commit-plan.md') `
        'plan' 'read-only' `
        'Propose a post-SDD commit plan without changing git state' `
        $PlanBody
    Render-CodexPrompt `
        (Join-Path $targetDir 'prompts\commit-apply.md') `
        'apply' 'state-changing' `
        'Execute an approved post-SDD commit plan, or generate one first if missing' `
        $ApplyBody
    Write-Host "Applied Codex overlays -> $targetDir"
}

# --- Main ---

Assert-Sources
$resolvedTargets = Resolve-Targets $Targets

foreach ($target in $resolvedTargets) {
    switch ($target) {
        'opencode' { Apply-OpenCode }
        'claude'   { Apply-Claude }
        'codex'    { Apply-Codex }
    }
}

Write-Host 'Reminder: re-run this script after syncs, upgrades, or managed config refreshes.'
