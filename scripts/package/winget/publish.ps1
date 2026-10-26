param(
    [string]$Version = "1.2.7",
    # Path to the NSIS exe that will actually be uploaded to the v$Version
    # release. The hash has to come from those exact bytes: Go embeds build
    # info, so two runs of build-installers.sh do not produce the same file and
    # a hash copied from an earlier build silently fails winget's check.
    [Parameter(Mandatory = $true)]
    [string]$Artifact
)

$ErrorActionPreference = 'Stop'

if (-not (Test-Path $Artifact)) {
    Write-Error "No installer at $Artifact. Build it first: ./scripts/package/build-installers.sh $Version"
}

if (-not (Get-Command wingetcreate -ErrorAction SilentlyContinue)) {
    Write-Error "wingetcreate is not installed. Install it from https://github.com/microsoft/winget-cli/releases"
}

# The manifests live in winget/manifest at the repo root, not beside this
# script, which is what the original Join-Path $PSScriptRoot looked for.
$repoRoot = Join-Path $PSScriptRoot '..\..\..'
$manifestDir = Join-Path $repoRoot 'winget\manifest'
$manifestPath = Join-Path $manifestDir "richavery.BvrCLI.$Version.manifest"
if (-not (Test-Path $manifestPath)) {
    Write-Error "Manifest not found at $manifestPath"
}

$sha = (Get-FileHash -Algorithm SHA256 -Path (Resolve-Path $Artifact)).Hash.ToLower()

$manifest = Get-Content -Raw -Path $manifestPath | ConvertFrom-Json
if ($manifest.Installers[0].InstallerSha256 -eq $sha) {
    Write-Host "Manifest already carries the hash for $Artifact."
} else {
    $manifest.Installers[0].InstallerSha256 = $sha
    $manifest | ConvertTo-Json -Depth 10 | Set-Content -Path $manifestPath -Encoding UTF8
    Write-Host "Wrote InstallerSha256 = $sha"
}

Write-Host 'Validating manifest...'
winget validate --manifest $manifestPath

Write-Host "Publishing manifest for richavery.BvrCLI $Version ..."
winget create --owner richavery --url https://github.com/microsoft/winget-pkgs --fork $repoRoot

Write-Host 'Done. Complete PR submission at https://github.com/microsoft/winget-pkgs'
