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
$CommitSkill = Join-Path $SharedDir 'skills\commit-planner\SKILL.md'
$PrSkill     = Join-Path $SharedDir 'skills\pr-finalizer\SKILL.md'
$PlanBody    = Join-Path $SharedDir 'commands\commit-plan-body.md'
$ApplyBody   = Join-Path $SharedDir 'commands\commit-apply-body.md'
$FastBody    = Join-Path $SharedDir 'commands\commit-fast-body.md'
$PrCreateBody = Join-Path $SharedDir 'commands\pr-create-body.md'
$PrRegenerateBody = Join-Path $SharedDir 'commands\pr-regenerate-body.md'

$SupportedTargets = @('opencode', 'claude', 'codex', 'gemini', 'antigravity')

# UTF-8 without BOM — consistent with the bash script output
$Utf8NoBom = New-Object System.Text.UTF8Encoding $false

function Show-Usage {
    $name = [System.IO.Path]::GetFileName($PSCommandPath)
    Write-Host "Usage: $name all | [opencode|claude|codex|gemini|antigravity ...]"
    Write-Host 'Examples:'
    Write-Host "  $name opencode"
    Write-Host "  $name claude codex"
    Write-Host "  $name gemini"
    Write-Host "  $name antigravity"
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
    Assert-SourceFile $CommitSkill
    Assert-SourceFile $PrSkill
    Assert-SourceFile $PlanBody
    Assert-SourceFile $ApplyBody
    Assert-SourceFile $FastBody
    Assert-SourceFile $PrCreateBody
    Assert-SourceFile $PrRegenerateBody
}

function Resolve-Targets([string[]]$rawTargets) {
    if ($null -eq $rawTargets -or $rawTargets.Count -eq 0) {
        Show-Usage
        exit 1
    }

    if ($rawTargets.Count -eq 1 -and $rawTargets[0] -eq 'all') {
        return , $SupportedTargets
    }

    $result = [System.Collections.Generic.List[string]]::new()
    foreach ($t in $rawTargets) {
        if ($t -eq 'all') {
            Exit-WithError "Use 'all' by itself, or pass explicit targets only."
        }
        if ($SupportedTargets -notcontains $t) {
            Exit-WithError "Unknown target: $t"
        }
        if ($result -notcontains $t) {
            $result.Add($t)
        }
    }
    return , $result.ToArray()
}

function Install-Skill([string]$targetDir, [string]$skillName, [string]$skillSource) {
    $dest = Join-Path $targetDir ("skills\$skillName")
    New-Item -ItemType Directory -Force -Path $dest | Out-Null
    Copy-Item $skillSource (Join-Path $dest 'SKILL.md') -Force
}

function Write-RenderedFile([string]$path, [string]$content) {
    $dir = [System.IO.Path]::GetDirectoryName($path)
    New-Item -ItemType Directory -Force -Path $dir | Out-Null
    [System.IO.File]::WriteAllText($path, $content, $Utf8NoBom)
}

function Render-OpenCodeCommand([string]$targetFile, [string]$skillName, [string]$mode, [string]$cmdType, [string]$desc, [string]$bodyFile) {
    $body = [System.IO.File]::ReadAllText($bodyFile, $Utf8NoBom)
    $lines = @(
        '---',
        "description: $desc",
        'agent: gentleman',
        '---',
        '',
        "Read the skill file at ``~/.config/opencode/skills/$skillName/SKILL.md`` FIRST, then follow it exactly.",
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

function Render-ClaudeCommand([string]$targetFile, [string]$skillName, [string]$mode, [string]$cmdType, [string]$desc, [string]$bodyFile) {
    $body = [System.IO.File]::ReadAllText($bodyFile, $Utf8NoBom)
    $lines = [System.Collections.Generic.List[string]]@(
        '---',
        "description: $desc",
        'argument-hint: [optional-context]',
        'allowed-tools:',
        '  - Read',
        '  - Glob',
        '  - Bash(git:*)',
        '  - Bash(gh:*)',
        '  - Bash(pwd:*)',
        '  - Bash(basename:*)'
    )
    if ($mode -eq 'apply' -or $mode -eq 'auto') {
        $lines.Add('disable-model-invocation: true')
    }
    $lines.AddRange([string[]]@(
        '---',
        '',
        "Read the skill file at ``~/.claude/skills/$skillName/SKILL.md`` FIRST, then follow it exactly.",
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

function Render-CodexPrompt([string]$targetFile, [string]$skillName, [string]$mode, [string]$cmdType, [string]$desc, [string]$bodyFile) {
    $body = [System.IO.File]::ReadAllText($bodyFile, $Utf8NoBom)
    $lines = @(
        '---',
        "description: $desc",
        'argument-hint: [optional-context]',
        'allowed-tools:',
        '  - Read',
        '  - Glob',
        '  - Bash(git:*)',
        '  - Bash(gh:*)',
        '  - Bash(pwd:*)',
        '  - Bash(basename:*)',
        '---',
        '',
        "Read the skill file at ``~/.codex/skills/$skillName/SKILL.md`` FIRST, then follow it exactly.",
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

function Render-GeminiCommand([string]$targetFile, [string]$skillName, [string]$mode, [string]$cmdType, [string]$desc, [string]$bodyFile) {
    $body = [System.IO.File]::ReadAllText($bodyFile, $Utf8NoBom)
    $lines = @(
        "description = `"$desc`"",
        'prompt = """',
        "Read the skill file at ``~/.gemini/skills/$skillName/SKILL.md`` FIRST, then follow it exactly.",
        '',
        'CONTEXT:',
        '- Working directory: !{pwd}',
        '- Current project: !{basename "$PWD"}',
        "- Mode: $mode",
        "- Command type: $cmdType",
        ''
    )
    Write-RenderedFile $targetFile ((($lines -join "`n") + "`n") + $body + "`n\"\"\"`n")
}

function Render-AntigravityWorkflow([string]$targetFile, [string]$skillName, [string]$mode, [string]$cmdType, [string]$desc, [string]$bodyFile) {
    $body = [System.IO.File]::ReadAllText($bodyFile, $Utf8NoBom)
    $lines = @(
        '---',
        "description: $desc",
        'type: workflow',
        'agent: gentleman',
        'allowed-tools:',
        '  - Read',
        '  - Glob',
        '  - Bash',
        '---',
        '',
        "Read the skill file at ``~/.antigravity/skills/$skillName/SKILL.md`` FIRST, then follow it exactly.",
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
    Install-Skill $targetDir 'commit-planner' $CommitSkill
    Install-Skill $targetDir 'pr-finalizer' $PrSkill
    Render-OpenCodeCommand `
        (Join-Path $targetDir 'commands\commit-plan.md') `
        'commit-planner' `
        'plan' 'read-only' `
        'Propose a post-SDD commit plan without changing git state' `
        $PlanBody
    Render-OpenCodeCommand `
        (Join-Path $targetDir 'commands\commit-apply.md') `
        'commit-planner' `
        'apply' 'state-changing' `
        'Execute an approved post-SDD commit plan, or generate one first if missing' `
        $ApplyBody
    Render-OpenCodeCommand `
        (Join-Path $targetDir 'commands\commit-fast.md') `
        'commit-planner' `
        'auto' 'state-changing' `
        'Generate and execute a commit plan in one shot without approval pause' `
        $FastBody
    Render-OpenCodeCommand `
        (Join-Path $targetDir 'commands\pr-create.md') `
        'pr-finalizer' `
        'create' 'state-changing' `
        'Draft a PR from committed changes and optionally create it after approval' `
        $PrCreateBody
    Render-OpenCodeCommand `
        (Join-Path $targetDir 'commands\pr-regenerate.md') `
        'pr-finalizer' `
        'regenerate' 'state-changing' `
        'Regenerate or update an existing PR from the current committed diff after approval' `
        $PrRegenerateBody
    Write-Host "Applied OpenCode overlays -> $targetDir"
}

function Apply-Claude {
    $targetDir = Join-Path $HOME '.claude'
    Install-Skill $targetDir 'commit-planner' $CommitSkill
    Install-Skill $targetDir 'pr-finalizer' $PrSkill
    Render-ClaudeCommand `
        (Join-Path $targetDir 'commands\commit-plan.md') `
        'commit-planner' `
        'plan' 'read-only' `
        'Propose a post-SDD commit plan without changing git state' `
        $PlanBody
    Render-ClaudeCommand `
        (Join-Path $targetDir 'commands\commit-apply.md') `
        'commit-planner' `
        'apply' 'state-changing' `
        'Execute an approved post-SDD commit plan, or generate one first if missing' `
        $ApplyBody
    Render-ClaudeCommand `
        (Join-Path $targetDir 'commands\commit-fast.md') `
        'commit-planner' `
        'auto' 'state-changing' `
        'Generate and execute a commit plan in one shot without approval pause' `
        $FastBody
    Render-ClaudeCommand `
        (Join-Path $targetDir 'commands\pr-create.md') `
        'pr-finalizer' `
        'create' 'state-changing' `
        'Draft a PR from committed changes and optionally create it after approval' `
        $PrCreateBody
    Render-ClaudeCommand `
        (Join-Path $targetDir 'commands\pr-regenerate.md') `
        'pr-finalizer' `
        'regenerate' 'state-changing' `
        'Regenerate or update an existing PR from the current committed diff after approval' `
        $PrRegenerateBody
    Write-Host "Applied Claude overlays -> $targetDir"
}

function Apply-Codex {
    $targetDir = Join-Path $HOME '.codex'
    Install-Skill $targetDir 'commit-planner' $CommitSkill
    Install-Skill $targetDir 'pr-finalizer' $PrSkill
    Render-CodexPrompt `
        (Join-Path $targetDir 'prompts\commit-plan.md') `
        'commit-planner' `
        'plan' 'read-only' `
        'Propose a post-SDD commit plan without changing git state' `
        $PlanBody
    Render-CodexPrompt `
        (Join-Path $targetDir 'prompts\commit-apply.md') `
        'commit-planner' `
        'apply' 'state-changing' `
        'Execute an approved post-SDD commit plan, or generate one first if missing' `
        $ApplyBody
    Render-CodexPrompt `
        (Join-Path $targetDir 'prompts\commit-fast.md') `
        'commit-planner' `
        'auto' 'state-changing' `
        'Generate and execute a commit plan in one shot without approval pause' `
        $FastBody
    Render-CodexPrompt `
        (Join-Path $targetDir 'prompts\pr-create.md') `
        'pr-finalizer' `
        'create' 'state-changing' `
        'Draft a PR from committed changes and optionally create it after approval' `
        $PrCreateBody
    Render-CodexPrompt `
        (Join-Path $targetDir 'prompts\pr-regenerate.md') `
        'pr-finalizer' `
        'regenerate' 'state-changing' `
        'Regenerate or update an existing PR from the current committed diff after approval' `
        $PrRegenerateBody
    Write-Host "Applied Codex overlays -> $targetDir"
}

function Apply-Gemini {
    $targetDir = Join-Path $HOME '.gemini'
    Install-Skill $targetDir 'commit-planner' $CommitSkill
    Install-Skill $targetDir 'pr-finalizer' $PrSkill
    Render-GeminiCommand `
        (Join-Path $targetDir 'commands\commit-plan.toml') `
        'commit-planner' `
        'plan' 'read-only' `
        'Propose a post-SDD commit plan without changing git state' `
        $PlanBody
    Render-GeminiCommand `
        (Join-Path $targetDir 'commands\commit-apply.toml') `
        'commit-planner' `
        'apply' 'state-changing' `
        'Execute an approved post-SDD commit plan, or generate one first if missing' `
        $ApplyBody
    Render-GeminiCommand `
        (Join-Path $targetDir 'commands\commit-fast.toml') `
        'commit-planner' `
        'auto' 'state-changing' `
        'Generate and execute a commit plan in one shot without approval pause' `
        $FastBody
    Render-GeminiCommand `
        (Join-Path $targetDir 'commands\pr-create.toml') `
        'pr-finalizer' `
        'create' 'state-changing' `
        'Draft a PR from committed changes and optionally create it after approval' `
        $PrCreateBody
    Render-GeminiCommand `
        (Join-Path $targetDir 'commands\pr-regenerate.toml') `
        'pr-finalizer' `
        'regenerate' 'state-changing' `
        'Regenerate or update an existing PR from the current committed diff after approval' `
        $PrRegenerateBody
    Write-Host "Applied Gemini overlays -> $targetDir"
}

function Apply-Antigravity {
    $targetDir = Join-Path $HOME '.antigravity'
    Install-Skill $targetDir 'commit-planner' $CommitSkill
    Install-Skill $targetDir 'pr-finalizer' $PrSkill
    Render-AntigravityWorkflow `
        (Join-Path $targetDir 'commands\commit-plan.md') `
        'commit-planner' `
        'plan' 'read-only' `
        'Propose a post-SDD commit plan without changing git state' `
        $PlanBody
    Render-AntigravityWorkflow `
        (Join-Path $targetDir 'commands\commit-apply.md') `
        'commit-planner' `
        'apply' 'state-changing' `
        'Execute an approved post-SDD commit plan, or generate one first if missing' `
        $ApplyBody
    Render-AntigravityWorkflow `
        (Join-Path $targetDir 'commands\commit-fast.md') `
        'commit-planner' `
        'auto' 'state-changing' `
        'Generate and execute a commit plan in one shot without approval pause' `
        $FastBody
    Render-AntigravityWorkflow `
        (Join-Path $targetDir 'commands\pr-create.md') `
        'pr-finalizer' `
        'create' 'state-changing' `
        'Draft a PR from committed changes and optionally create it after approval' `
        $PrCreateBody
    Render-AntigravityWorkflow `
        (Join-Path $targetDir 'commands\pr-regenerate.md') `
        'pr-finalizer' `
        'regenerate' 'state-changing' `
        'Regenerate or update an existing PR from the current committed diff after approval' `
        $PrRegenerateBody
    Write-Host "Applied Antigravity overlays -> $targetDir"
}

# --- Main ---

$resolvedTargets = Resolve-Targets $Targets
Assert-Sources

foreach ($target in $resolvedTargets) {
    switch ($target) {
        'opencode' { Apply-OpenCode }
        'claude'   { Apply-Claude }
        'codex'    { Apply-Codex }
        'gemini'   { Apply-Gemini }
        'antigravity' { Apply-Antigravity }
    }
}

Write-Host 'Reminder: re-run this script after syncs, upgrades, or managed config refreshes.'
