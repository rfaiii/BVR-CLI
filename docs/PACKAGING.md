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

What `build-installers.sh` actually puts in each package today:

- `bvr-cli` binary at `/usr/bin` (Windows: `$INSTDIR\bvr-cli.exe`)
- README and license
- Bash, zsh, and fish completions from `packaging/nfpm/completions/`
- Man page `packaging/nfpm/manpages/bvr-cli.1.gz`

Icon assets in `resources/icons/` and `packaging/icons/` are **not** installed
by any package format here. A terminal TUI has no desktop entry to point at
them, so nothing in the current pipeline consumes them; wiring up an optional
`.desktop`/hicolor-icon install for the deb is unfinished work, not a
present-day behavior.

## Notes

- **Nothing in this repository signs or notarizes anything.** `scripts/package/macos-dmg.sh`
  and `scripts/package/build-installers.sh` produce **unsigned** DMGs; macOS
  Gatekeeper will block the first run until the user right-clicks through the
  warning or clears the quarantine attribute. Developer ID signing and
  `notarytool` stapling are GoReleaser (PRO) steps and need certificates that
  are not in this checkout. The Windows EXE is likewise unsigned, so
  SmartScreen will warn on first run.
- `./scripts/package/build-installers.sh [version]` builds the full installer
  set locally: unsigned macOS DMGs (arm64 + Intel), an Ubuntu `.deb`,
  Omarchy/Arch `.pkg.tar.zst` (x86_64 + aarch64), and a Windows x64 NSIS
  installer. Requires `nfpm`, `makensis`, and `hdiutil`.
- The nfpm configs are tracked templates at `packaging/nfpm/{deb,archlinux}.yaml`
  and read `${BVR_VERSION}` / `${BVR_DEB_ARCH}` / `${BVR_ARCH_ARCH}` from the
  environment. Note that nfpm expands those variables in metadata fields but
  **not** inside `contents.src`, so the build script always stages the
  cross-compiled binary at `dist/pkg-staging/bvr-cli` before invoking nfpm.
- The Windows installer is per-user (`RequestExecutionLevel user`, installs to
  `$LOCALAPPDATA\Programs\bvr-cli`) and writes an Add/Remove Programs entry,
  which is what makes the winget manifest's `InstallerType: exe` and
  `ElevationRequirement: user` accurate. Before 1.2.7 the artifact published
  under that name was the bare console binary, so winget would have "installed"
  it by launching the TUI.
- `./scripts/package/build-beta-packages.sh 1.2.7` creates portable
  tar/zip archives for six OS/arch pairs plus the Arch packages, for smoke
  testing on target machines.
- The GoReleaser configuration produces Linux `deb`, `rpm`, `apk`, and
  ArchLinux packages, plus the Arch User Repository packages `bvr-cli` (source)
  and `bvr-cli-bin` (prebuilt binaries). On Arch and Omarchy install with
  `yay -S bvr-cli` (or `bvr-cli-bin` for the binary variant).
- The current `.goreleaser.yml` targets cross-platform archives and tap/scoop/npm manifests.
- Windows MSI generation is available through
  `scripts/package/windows-msi.ps1` when WiX Toolset is installed. It consumes
  the raw `bvr-cli_<version>_windows_<arch>.exe` **binary** produced by
  `scripts/package/windows.bat`, which is a different artifact from the NSIS
  installer of the same name; keep the two straight when publishing.
- `winget/manifest/richavery.BvrCLI.<version>.manifest` is an installer-schema
  1.6.0 document. The 1.2.7 one previously declared `InstallerType: exe` with a
  `Url` key, an `InstallationMetadata.InstallLocation` that is not a valid
  field, an `InstallerFunction` that does not exist in the schema, and locale
  strings (`Moniker`, `Publisher`, `License`) that belong in the separate
  locale manifest. It now uses `InstallerType: nullsoft` with `InstallerUrl` /
  `InstallerSha256` inside the `Installers` entry, matching what the NSIS build
  produces. `winget/manifest/richavery.BvrCLI.1.2.4.manifest` is left as the
  historical record of the old shape.
- `InstallerSha256` must be the hash of the bytes that actually get uploaded.
  `scripts/package/winget/publish.ps1 -Artifact <path>` computes and writes it
  rather than relying on a hand-typed value.
- A real winget PR needs the three-file layout (version + installer + locale
  manifests) under `manifests/r/richavery/BvrCLI/<version>/`. This repo keeps
  one hand-written installer document per version; `wingetcreate` expands it.
  Nothing here has been through `winget validate` on Windows.

## Image and icon resources

This section lists the image assets for BVR, where they live in the repo,
and how to regenerate or replace them. It was formerly `IMG-RESOURCES.md`.

### Installing a new icon set

One command updates every icon surface from a designer drop:

```sh
python3 scripts/update_icons.py ~/Desktop/BVR-CLI/V127/ICONS/bvr-cli-AppIcons
```

The source directory must contain the iOS asset-catalog names
(`Icon-20.png` … `Icon-1024.png`); the script validates each pixel size
before writing anything. Everything else is derived from `Icon-1024.png`:
the iOS `Assets.xcassets` set, the `packaging/icons/` staging copies, the
`resources/icons/ios/` generator set, `bvr-icon.icns`, `bvr-icon.ico`, and
the 128×128 PNG embedded for OSC 99 notifications.

Icons are composited onto black with the alpha channel dropped: the App
Store rejects a 1024 image with transparency, and the source exports carry
a transparent border outside the rounded square.

### Source logo

- `resources/icons/bvr-logo-1024.png`
  - High-resolution app logo.
  - Source for app icons, installer graphics, and README branding.
  - `resources/icons/bvr-icon-00.png` is a byte-identical copy kept for the
    branding references that predate the logo path.

### macOS icon

- `resources/icons/bvr-icon.icns`
  - Use in macOS app bundles, DMG disk images, and macOS installer metadata.
  - Regenerate with `scripts/update_icons.py`, or by hand from an iconset
    folder holding 16/32/64/128/256/512/1024 px PNGs:
    ```sh
    iconutil -c icns bvr-icon.iconset -o resources/icons/bvr-icon.icns
    ```

### Windows icon

- `resources/icons/bvr-icon.ico`
  - Use in Windows EXE/MSI installers and shortcuts.
  - Regenerate with `scripts/update_icons.py`, or by hand with ImageMagick:
    ```sh
    magick resources/icons/bvr-logo-1024.png \
      -define icon:auto-resize="256,128,64,48,32,16" \
      resources/icons/bvr-icon.ico
    ```

### iOS / Mac app icons

- `resources/icons/ios/`
  - Complete iOS app icon set generated from the source logo.
  - Sizes include 16×16 through 1024×1024, including @2x/@3x variants.
- `mobile-cli/ios/BvrCompanion/BvrCompanion/Assets.xcassets/AppIcon.appiconset/`
  - The set the companion app actually builds against, using the
    `Icon-<point size>[@2x|@3x].png` names its `Contents.json` lists.
  - Prefer editing this catalog directly; `resources/icons/ios/` is the
    generator-format mirror.

### Notification icon

- `internal/ui/notification/bvr-icon.png` and `bvr-icon-solo.png`
  - Embedded into the binary and base64-encoded into every OSC 99
    notification, so they stay at 128×128 on purpose.

### Screenshots

The committed product screenshots are:

| File | Subject |
| --- | --- |
| `resources/screenshots/BVR-CLI-HOME-MENU.png` | Homescreen and launch actions. |
| `resources/screenshots/BVR-CLI.png` | Full working screen: chat, editor, and sidebar. |
| `resources/screenshots/COMMAND-MENU.png` | Command palette and navigation. |
| `resources/screenshots/COMMAND.png` | Command palette detail. |
| `resources/screenshots/HEADER.png` | Header treatment. |
| `resources/screenshots/FILE-FINDER.png` | Project file browser. |
| `resources/screenshots/CREATE_FILE.png` | New-file workflow. |
| `resources/screenshots/IMAGE_RECOGNITION.png` | Image recognition / lookup flow. |
| `resources/screenshots/BACKGROUND.png` | Terminal background treatment. |

The older per-panel screenshots (`BVR-MASCOT`, `MODEL-INFO`, `NODE-INFO`,
`PROJECT-TEXT`, `SKILLS-INFO`, `FOOTER`, `WEB-BROWSER`) were removed in 1.2.7;
their content is now covered by the full working-screen and sidebar captures.
`CREATE-FILE.png` was renamed to `CREATE_FILE.png` and `file-finder.png` to
`FILE-FINDER.png` so every screenshot follows one `UPPER_SNAKE_CASE`
convention. On case-sensitive filesystems (Linux CI, release packaging) the
old lowercase name would not have resolved.

Use the homescreen and command menu images in the root README. The remaining
images are supporting references for feature and design documentation.

### Future media

Animations and walkthrough videos are intentionally not listed as committed
assets until they exist in the repository. Keep future GIFs under 5 MB and use
MP4/WebM for longer external walkthroughs rather than embedding them in the
root README. Demo recorder input lives in `docs/media/` (`.tape` files).

### Usage in documentation

Reference images from this folder in documentation with paths like:

```md
![BVR-CLI home menu](../resources/screenshots/BVR-CLI-HOME-MENU.png)
```

### About / contact assets

- Company/personal contact details belong in `docs/FEATURES.md`.
- Use `resources/icons/bvr-logo-1024.png` as the About pane icon where
  applicable.

### Notes

- Keep all source PNGs under `resources/`.
- Commit `.icns` and `.ico` only when they are part of a release or installer
  packaging change; otherwise keep them in `resources/icons/` as build inputs.
- Regenerate the whole set with `scripts/update_icons.py` rather than editing
  derived files by hand.
