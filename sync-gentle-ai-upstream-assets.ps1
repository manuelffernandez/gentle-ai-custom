#Requires -Version 5.1
[CmdletBinding()]
param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$Args
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$SourceDir = $PSScriptRoot
$goName = if ($env:GO) { $env:GO } else { 'go' }
$goCommand = Get-Command $goName -ErrorAction SilentlyContinue
if (-not $goCommand) {
    [Console]::Error.WriteLine('ERROR: go is required to sync approved Gentle AI upstream assets')
    exit 1
}

Push-Location $SourceDir
try {
    $env:GENTLE_AI_CUSTOM_ENTRYPOINT = Split-Path -Leaf $PSCommandPath
    & $goCommand.Source run .\cmd\gentle-ai-overlay --repo-root $SourceDir sync-upstream-assets @Args
    exit $LASTEXITCODE
}
finally {
    Pop-Location
}
