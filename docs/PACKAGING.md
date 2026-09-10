# Packaging & Distribution

This document summarizes the current installer/distribution setup for BVR
and where each artifact lives.

## Release automation

- `.goreleaser.yml` — main release config
  - Homebrew tap: `richavery/homebrew-tap`
  - Scoop bucket: `richavery/scoop-bucket`
  - NPM package: `@bvrcli/bvr-cli`
  - Linux packages via `nfpms`
  - AUR source/binary packages
  - Nix package
  - Winget manifest
- `.github/workflows/release.yml` — GitHub Actions release workflow

## Platform packaging scripts

- `scripts/package/macos.sh` — build and zip macOS binary
- `scripts/package/windows.bat` — build and zip Windows binary
- `scripts/package/build-beta-packages.sh` — build local beta archives for
  macOS ARM64/Intel, Windows x64/ARM64, and Linux x64/ARM64
- `.goreleaser.dist.windows.yaml` — Windows archive/zip fallback config

## Install wrappers

- `packaging/homebrew-bvr-cli.rb` — Homebrew formula skeleton
- `packaging/npm-package.json` — NPM package manifest
- `packaging/scripts/install.js` — NPM postinstall downloader
- `packaging/scripts/uninstall.js` — NPM uninstall cleanup

## Installer contents

Each package should include:

- `bvr-cli` binary
- README and license
- Shell completions
- Man page
- Icon assets from `resources/icons/`

## Notes

- Signed/notarized DMG/EXE/MSI packaging is not yet implemented.
- Run `./scripts/package/build-beta-packages.sh 1.2.3` to create six local
  archives under `dist/packages/` for smoke testing on the target machines.
- The GoReleaser configuration produces Linux `deb`, `rpm`, `apk`, and
  ArchLinux packages, plus release archives for the same six targets.
- The current `.goreleaser.yml` targets cross-platform archives and tap/scoop/npm manifests.
- Windows MSI generation is available through
  `scripts/package/windows-msi.ps1` when WiX Toolset is installed.
