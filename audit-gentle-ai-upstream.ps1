#Requires -Version 5.1
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$python = if ($env:PYTHON) { $env:PYTHON } else { 'python' }
$pythonCommand = Get-Command $python -ErrorAction SilentlyContinue
if (-not $pythonCommand) {
    [Console]::Error.WriteLine('ERROR: python is required to audit the Gentle AI upstream baseline')
    exit 1
}

$scriptPath = Join-Path $PSScriptRoot 'overlay\gentle-ai\scripts\audit-gentle-ai-upstream.py'
& $pythonCommand.Source $scriptPath
exit $LASTEXITCODE
