# BVR-CLI UI branding

## Wordmark source of truth

The display name is `logo.Wordmark` in
`internal/ui/logo/logo.go`. Use `scripts/set-wordmark.sh` to change it:

```bash
./scripts/set-wordmark.sh BVR-CLI
```

The script validates the name and updates the shared constant used by the
compact header, sidebar, and custom wide banner.

## Wide banner font

The wide banner does not use the original BVR letterform list. Its glyphs are
defined in `bvrCLIASCII`, a five-row map of block-character strings. The
renderer assembles those glyphs with `renderASCIIWordmark`, applies the theme's
title gradient, and places the result between the diagonal side fields.

To change a letter, edit that letter's five strings in `bvrCLIASCII`. Keep all
five rows the same display width so the banner remains aligned. Add a new map
entry before using a character in `Wordmark`.

## Persistent prompt-style animation

The strip below the banner reuses `internal/ui/anim.Anim`, the same component
used by the `Generating` prompt spinner and assistant thinking indicators. The
banner creates one animation with ID `bvr-banner`, cycling characters, and the
theme logo gradient. `UI.Init` starts its 20 FPS `StepMsg` chain; the main UI
`Update` loop forwards matching ticks to the animation, and `renderLogo` passes
`bannerAnim.Render()` to the logo package.

The strip is intentionally rendered below the banner with no blank row. Its
width is truncated to the current terminal width by the logo renderer. Adjust
the animation's `Size`, colors, or `CycleColors` settings in `UI.New` if a
different density or palette is desired.