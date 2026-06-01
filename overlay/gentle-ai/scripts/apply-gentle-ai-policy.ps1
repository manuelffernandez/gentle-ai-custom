#Requires -Version 5.1
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$ScriptDir = Split-Path -Parent $PSCommandPath
$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $ScriptDir '..\..\..'))
$goName = if ($env:GO) { $env:GO } else { 'go' }
$goCommand = Get-Command $goName -ErrorAction SilentlyContinue
if (-not $goCommand) {
    [Console]::Error.WriteLine('ERROR: go is required to apply the Gentle AI overlay policy')
    exit 1
}

Push-Location $RepoRoot
try {
    $env:GENTLE_AI_CUSTOM_ENTRYPOINT = Split-Path -Leaf $PSCommandPath
    & $goCommand.Source run .\cmd\gentle-ai-overlay --repo-root $RepoRoot apply-policy @args
    exit $LASTEXITCODE
}
finally {
    Pop-Location
}
