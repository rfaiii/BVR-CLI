param(
    [string]$Version = "1.2.1",
    [string]$Arch = "amd64",
    [string]$Root = "."
)

$ErrorActionPreference = 'Stop'

$root = Resolve-Path $Root
$dist = Join-Path $root "dist\windows"
$msiDir = Join-Path $dist "msi"
$srcExe = Join-Path $root "dist\release\bvr-cli_${Version}_windows_${Arch}.exe"

if (-not (Test-Path $srcExe)) {
    Write-Error "Missing built exe at $srcExe. Run scripts\package\windows.bat first."
}

if (-not (Get-Command candle -ErrorAction SilentlyContinue)) {
    Write-Error "WiX Toolset is not installed. Install it from https://wixtoolset.org/releases/"
}

$wxs = Join-Path $msiDir "BvrCli.wxs"
if (-not (Test-Path $wxs)) {
    New-Item -ItemType Directory -Path $msiDir -Force | Out-Null
    $content = @"
<?xml version="1.0" encoding="UTF-8"?>
<Wix xmlns="http://schemas.microsoft.com/wix/2006/wi">
  <Product Id="*" Name="BVR CLI" Language="1033" Version="$Version" Manufacturer="richavery" UpgradeCode="A1B2C3D4-E5F6-7890-1234-56789ABCDEF0">
    <Package InstallerVersion="500" Compressed="yes" InstallScope="perMachine" />
    <Media Id="1" Cabinet="cab1.cab" EmbedCab="yes" />
    <Directory Id="TARGETDIR" Name="SourceDir">
      <Directory Id="ProgramFilesFolder">
        <Directory Id="INSTALLFOLDER" Name="BVR CLI">
          <Component Id="BvrCliExe" Guid="*">
            <File Id="BvrCliExe" Source="$srcExe" KeyPath="yes" />
            <Environment Id="PATH" Name="PATH" Value="[INSTALLFOLDER]" Action="set" System="yes" Part="last" />
          </Component>
        </Directory>
      </Directory>
    </Directory>
    <Feature Id="MainFeature" Title="BVR CLI" Level="1">
      <ComponentRef Id="BvrCliExe" />
    </Feature>
  </Product>
</Wix>
"@
    Set-Content -Path $wxs -Value $content -Encoding UTF8
}

Push-Location $msiDir
try {
    candle BvrCli.wxs
    light -ext WixUIExtension BvrCli.wixobj -o "$dist\bvr-cli_${Version}_windows_${Arch}.msi"
}
finally {
    Pop-Location
}
