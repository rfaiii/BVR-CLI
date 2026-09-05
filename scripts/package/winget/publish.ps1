param(
    [string]$Version = "1.2.1"
)

$ErrorActionPreference = 'Stop'

if (-not (Get-Command wingetcreate -ErrorAction SilentlyContinue)) {
    Write-Error "wingetcreate is not installed. Install it from https://github.com/microsoft/winget-cli/releases"
}

$manifestPath = Join-Path $PSScriptRoot "richavery.BvrCLI.$Version.manifest"
if (-not (Test-Path $manifestPath)) {
    Write-Error "Manifest not found at $manifestPath"
}

Write-Host "Validating manifest..."
winget validate --manifest $PSScriptRoot

Write-Host "Publishing manifest for richavery.BvrCLI $Version ..."
winget create --owner richavery --url https://github.com/microsoft/winget-pkgs --fork $PSScriptRoot

Write-Host "Done. Complete PR submission at https://github.com/microsoft/winget-pkgs"
