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

The current committed product screenshots are:

- `resources/screenshots/BVR-CLI-HOME-MENU.png` - homescreen and launch actions.
- `resources/screenshots/COMMAND-MENU.png` - command palette and navigation.
- `resources/screenshots/HEADER.png` - header treatment.
- `resources/screenshots/BVR-MASCOT.png` - Beaver branding reference.
- `resources/screenshots/file-finder.png` - project file browser.
- `resources/screenshots/CREATE-FILE.png` - new-file workflow.
- `resources/screenshots/WEB-BROWSER.png` - embedded browser.
- `resources/screenshots/MODEL-INFO.png` - model/provider information.
- `resources/screenshots/NODE-INFO.png` - NODE status.
- `resources/screenshots/SKILLS-INFO.png` - skills status.
- `resources/screenshots/FOOTER.png` - footer/status treatment.

Use the homescreen and command menu images in the root README. The remaining
images are supporting references for feature and design documentation.

## Future media

Animations and walkthrough videos are intentionally not listed as committed
assets until they exist in the repository. Keep future GIFs under 5 MB and use
MP4/WebM for longer external walkthroughs rather than embedding them in the
root README.

## Usage in README

Reference images from this folder in documentation with paths like:

```md
![BVR-CLI home menu](../resources/screenshots/BVR-CLI-HOME-MENU.png)
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
