# Image & Icon Resources

This document lists the image assets for BVR, where they live in the repo,
and how to regenerate or replace them.

## Source logo

- `resources/icons/bvr-logo-1024.png`
  - High-resolution app logo.
  - Source for app icons, installer graphics, and README branding.

## macOS icon

- `resources/icons/bvr-icon.icns`
  - Use in macOS app bundles, DMG disk images, and macOS installer metadata.
  - Regenerate from the source logo with:
    ```sh
    iconutil -c icns resources/icons/bvr-icon.iconset -o resources/icons/bvr-icon.icns
    ```

## Windows icon

- `resources/icons/bvr-icon.ico`
  - Use in Windows EXE/MSI installers and shortcuts.
  - Regenerate from the source logo with ImageMagick:
    ```sh
    magick convert resources/icons/bvr-logo-1024.png \
      -define icon:auto-resize="256,128,64,48,32,16" \
      resources/icons/bvr-icon.ico
    ```

## iOS / Mac app icons

- `resources/icons/ios/`
  - Complete iOS app icon set exported from the source logo.
  - Use these directly in `Assets.xcassets` for the companion app.
  - Sizes include 16×16 through 1024×1024, including @2x/@3x variants.

## General app icons

- `resources/icons/bvr-icon-00.png`
  - App icon for notifications and general UI branding.
- `resources/icons/bvr-icon-info.png`
  - Info/about-style icon variant.

## Screenshots

- `resources/screenshots/`
  - Place terminal/app screenshots here for README and marketing.
  - Suggested filenames:
    - `bvr-cli-macos.png`
    - `bvr-cli-windows.png`
    - `bvr-cli-linux.png`
    - `bvr-cli-onboarding.png`

## Animations

Use GIF for short terminal animations only when the motion adds value.
Keep these under 5 MB for GitHub rendering.

| Filename | Purpose | Duration |
| --- | --- | --- |
| `resources/screenshots/bvr-cli-onboarding.gif` | Onboarding flow demo | 5–10s |
| `resources/screenshots/bvr-cli-session.gif` | Live typing/response demo | 10–15s |

## Video walkthroughs

Use MP4 or WebM for external documentation and release pages.
Do not rely on GitHub README for video playback.

| Filename | Purpose | Notes |
| --- | --- | --- |
| `resources/videos/bvr-cli-onboarding.mp4` | Full onboarding walkthrough | 1080p, <2 min |
| `resources/videos/bvr-cli-install-macos.mp4` | macOS install walkthrough | 1080p, <1 min |
| `resources/videos/bvr-cli-install-windows.mp4` | Windows install walkthrough | 1080p, <1 min |

## Usage in README

Reference images from this folder in the README with paths like:

```md
![BVR on macOS](resources/screenshots/bvr-cli-macos.png)
```

## About / contact assets

- Company/personal contact details belong in `docs/ABOUT.md`.
- Use `resources/icons/bvr-icon-info.png` as the About pane icon where
  applicable.

## Notes

- Keep all source PNGs under `resources/`.
- Commit `.icns` and `.ico` only when they are part of a release or installer
  packaging change; otherwise keep them in `resources/icons/` as build inputs.
- If you regenerate the icon set, replace the files in place and update
  this document.
