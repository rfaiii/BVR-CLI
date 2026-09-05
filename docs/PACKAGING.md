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
- The current `.goreleaser.yml` targets cross-platform archives and tap/scoop/npm manifests.
- Windows installer scripts under `scripts/windows/` are placeholders for future NSIS/WiX work.
