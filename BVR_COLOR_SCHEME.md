# BVR Color Scheme

This document is the source of truth for BVR's theming system. It describes
the current color roles, markdown rendering tokens, layout primitives, and the
steps for adding a complete theme.

## Current theme structure

A BVR theme is built from two parts:

1. `quickStyleOpts` palette in `internal/ui/styles/quickstyle.go`.
2. Theme overrides in a theme function under `internal/ui/styles/themes.go`.

`quickStyle(o)` builds the base theme from semantic color roles. The theme
function may override specific fields after that. `ThemeForProvider(providerID)`
returns the active theme, currently `DarkBvrTheme()` for all providers.

The main theme styles are collected in the `Styles` struct in
`internal/ui/styles/styles.go`. Key groups include:

- `Background`
- `Dialog`, `Dialog.FileBrowser`
- `TextInput`, `Editor`
- `Markdown`
- `Section`
- `Initialize`
- `LSP`
- `Sidebar`
- `ModelInfo`
- `Resource`
- `Files`
- `Messages`, `Messages.Shell*`
- `Tool`, `Tool.Diff*`, `Tool.Todo*`, `Tool.MCP*`, `Tool.Job*`
- `Status`, `Status.Resource*`
- `Completions`
- `Attachments`
- `Pills`

## Color roles

These are the semantic roles exposed by `quickStyleOpts`:

- `primary`
- `secondary`
- `accent`
- `keyword`
- `fgBase`
- `bgBase`
- `separator`
- `fgSubtle`
- `fgMoreSubtle`
- `fgMostSubtle`
- `onPrimary`
- `bgMostVisible`
- `bgLessVisible`
- `bgLeastVisible`
- `destructive`
- `error`
- `warning`
- `warningSubtle`
- `denied`
- `busy`
- `info`
- `infoMoreSubtle`
- `infoMostSubtle`
- `success`
- `successMoreSubtle`
- `successMostSubtle`
- `ansiBlack`, `ansiRed`, `ansiGreen`, `ansiYellow`, `ansiBlue`, `ansiMagenta`, `ansiCyan`, `ansiWhite`
- `ansiBrightBlack`, `ansiBrightRed`, `ansiBrightGreen`, `ansiBrightYellow`, `ansiBrightBlue`, `ansiBrightMagenta`, `ansiBrightCyan`, `ansiBrightWhite`

Use this palette to reason about contrast, hierarchy, and status semantics.
`quickStyle` maps each role into the concrete `Styles` fields used across the UI.

## Theme palette reference

Use this section when you need exact theme colors for docs or marketing.
Each theme is listed with its role names and hex values.

### Rich Aizen Green

- `Primary`: `#3BF66B`
- `Secondary`: `#6BFF91`
- `Gradient`: `#3BF66B`
- `Surface`: `#0C0E0D`
- `SurfaceSubtle`: `#141716`
- `SurfaceMuted`: `#222825`
- `OnSurface`: `#FFFFFF`
- `Muted`: `#8C9691`
- `Subtle`: `#59645E`
- `Border`: `#222825`
- `StatusSuccess`: `#3BF66B`
- `StatusError`: `#FF5F56`
- `StatusWarning`: `#F2C14E`
- `StatusInfo`: `#8CDED0`
- `CodeBackground`: `#141716`

### Crazy Jeff Pink

- `Primary`: `#FF4FA3`
- `Secondary`: `#FF86C8`
- `Gradient`: `#FF4FA3`
- `Surface`: `#14060C`
- `SurfaceSubtle`: `#1E0814`
- `SurfaceMuted`: `#2E1020`
- `OnSurface`: `#FFF0F5`
- `Muted`: `#C27A8C`
- `Subtle`: `#7A4050`
- `Border`: `#2E1020`
- `StatusSuccess`: `#FF4FA3`
- `StatusError`: `#FF5F56`
- `StatusWarning`: `#FFB86C`
- `StatusInfo`: `#FF86C8`
- `CodeBackground`: `#1E0814`

### Kobe Yang Purple

- `Primary`: `#B56CFF`
- `Secondary`: `#E0B3FF`
- `Gradient`: `#B56CFF`
- `Surface`: `#0E0816`
- `SurfaceSubtle`: `#18101F`
- `SurfaceMuted`: `#261633`
- `OnSurface`: `#F5F3FF`
- `Muted`: `#A090B8`
- `Subtle`: `#5E4F73`
- `Border`: `#261633`
- `StatusSuccess`: `#B56CFF`
- `StatusError`: `#FF5F56`
- `StatusWarning`: `#FFD866`
- `StatusInfo`: `#C4A8FF`
- `CodeBackground`: `#18101F`

### Steve DaBeav Blue

- `Primary`: `#5CC8FF`
- `Secondary`: `#A9E7FF`
- `Gradient`: `#5CC8FF`
- `Surface`: `#081218`
- `SurfaceSubtle`: `#0F1C24`
- `SurfaceMuted`: `#162C38`
- `OnSurface`: `#F0F9FF`
- `Muted`: `#82A8BF`
- `Subtle`: `#4C6878`
- `Border`: `#162C38`
- `StatusSuccess`: `#5CC8FF`
- `StatusError`: `#FF5F56`
- `StatusWarning`: `#FFD866`
- `StatusInfo`: `#A9E7FF`
- `CodeBackground`: `#0F1C24`

### Jenny Ann Orange

- `Primary`: `#FF8A3D`
- `Secondary`: `#FFC078`
- `Gradient`: `#FF8A3D`
- `Surface`: `#16100A`
- `SurfaceSubtle`: `#241C12`
- `SurfaceMuted`: `#3A2A18`
- `OnSurface`: `#FFF7ED`
- `Muted`: `#B08C6E`
- `Subtle`: `#7C5E40`
- `Border`: `#3A2A18`
- `StatusSuccess`: `#FF8A3D`
- `StatusError`: `#FF5F56`
- `StatusWarning`: `#FFD866`
- `StatusInfo`: `#FFC078`
- `CodeBackground`: `#241C12`

### Felix Tornado White

- `Primary`: `#FFFFFF`
- `Secondary`: `#E8EEF2`
- `Gradient`: `#FFFFFF`
- `Surface`: `#111316`
- `SurfaceSubtle`: `#1C1F24`
- `SurfaceMuted`: `#2C3038`
- `OnSurface`: `#F8FAFC`
- `Muted`: `#C7CDD4`
- `Subtle`: `#8A929C`
- `Border`: `#2C3038`
- `StatusSuccess`: `#E8EEF2`
- `StatusError`: `#FF5F56`
- `StatusWarning`: `#F2C14E`
- `StatusInfo`: `#A9E7FF`
- `CodeBackground`: `#1C1F24`

### Luis Mellow Yellow

- `Primary`: `#D6C84A`
- `Secondary`: `#F2E98B`
- `Gradient`: `#D6C84A`
- `Surface`: `#14120A`
- `SurfaceSubtle`: `#1F1D12`
- `SurfaceMuted`: `#302B18`
- `OnSurface`: `#FFFDF0`
- `Muted`: `#B5AD70`
- `Subtle`: `#6E6940`
- `Border`: `#302B18`
- `StatusSuccess`: `#D6C84A`
- `StatusError`: `#FF5F56`
- `StatusWarning`: `#FFD866`
- `StatusInfo`: `#F2E98B`
- `CodeBackground`: `#1F1D12`

### Bobur Blood Red

- `Primary`: `#FF1F1F`
- `Secondary`: `#FF6B6B`
- `Gradient`: `#FF1F1F`
- `Surface`: `#140606`
- `SurfaceSubtle`: `#1E0A0A`
- `SurfaceMuted`: `#2E1414`
- `OnSurface`: `#FFF0F0`
- `Muted`: `#C28A8A`
- `Subtle`: `#7A4A4A`
- `Border`: `#2E1414`
- `StatusSuccess`: `#FF1F1F`
- `StatusError`: `#FF5F56`
- `StatusWarning`: `#FFB86C`
- `StatusInfo`: `#FF8A8A`
- `CodeBackground`: `#1E0A0A`

## Bottom resource bar

The bottom resource bar is a persistent status surface. Its styles are:

- `Status.ResourceLabel`
- `Status.ResourceFilled`
- `Status.ResourceEmpty`
- `Status.ResourceValue`

Reserve these rows in theme designs so resource indicators remain visible
across resize and compact/non-compact modes.

## Markdown rendering

BVR renders markdown through Glamour v2. The active theme configures
`Styles.Markdown`, an `ansi.StyleConfig` from `charm.land/glamour/v2/ansi`.

Supported Glamour styling points include:

- `Document`
- `BlockQuote`
- `List`
- `Heading`
- `H1` through `H6`
- `Strikethrough`
- `Emph`
- `Strong`
- `HorizontalRule`
- `Item`
- `Enumeration`
- `Task`
- `Link`
- `LinkText`
- `Image`
- `ImageText`
- `Code`
- `CodeBlock`
- `Table`
- `DefinitionDescription`

`CodeBlock.Chroma` maps code token colors through `Styles.ChromaTheme()` in
`internal/ui/styles/styles.go`. The chroma theme follows the Glamour style
guide's token set.

BVR also defines a quieter markdown variant:

- `Styles.QuietMarkdown`

Use it for muted backgrounds such as thinking content.

## Layout primitives from Lip Gloss

BVR's theme code uses `charm.land/lipgloss/v2` directly. Useful primitives
for theme authors:

- color utilities: `lipgloss.Color`, `lipgloss.Darken`, `lipgloss.Lighten`,
  `lipgloss.Complementary`, `lipgloss.Alpha`
- block formatting: `Padding*`, `Margin*`, `Width`, `Height`, `Align`
- borders: `BorderStyle`, `BorderForeground`, `BorderForegroundBlend`,
  `RoundedBorder`, `NormalBorder`, `ThickBorder`, `DoubleBorder`
- joining: `JoinHorizontal`, `JoinVertical`
- placement: `Place`, `PlaceHorizontal`, `PlaceVertical`
- measurement: `lipgloss.Width`, `lipgloss.Height`, `lipgloss.Size`
- rendering: `Render`, `Println`, `Sprint`
- compositing: `lipgloss.NewLayer(...).X(...).Y(...).Z(...)` and
  `compositor.Compose(...)`

## CRUSH by Charm tm

BVR's UI stack is built on Charm libraries. The relevant pieces are already
available through Go module dependencies:

- `charm.land/lipgloss/v2`
- `charm.land/glamour/v2`
- `charm.land/bubbletea/v2`
- `charm.land/bubbles/v2`

Reference copies are present under `./CRUSH/` in this workspace:

- `./CRUSH/lipgloss`
- `./CRUSH/glamour`
- `./CRUSH/glamour/styles`

Use these copies when you need exact examples for:

- `glamour` style schema and JSON style generation
- `glamour` built-in styles such as `dracula` and `tokyo-night`
- `lipgloss` v2 API, color helpers, and upgrade patterns

## Current color combos

These are the effective combinations in the current default theme:

- `fgBase` on `bgBase`: primary readable text
- `fgSubtle` / `fgMoreSubtle` / `fgMostSubtle` on `bgBase`: secondary text,
  hints, and inactive labels
- `onPrimary` on `primary`: short labels or indicators that sit on brand color
- `fgBase` on `bgLessVisible`: panels and floating content surfaces
- `fgBase` on `bgLeastVisible`: thinking boxes and deepest surfaces
- `accent` for prompts, focused cursors, and key interactive glyphs
- `info` for headings and top-level structure
- `warningSubtle` for H1 badges
- `success`, `warning`, `error`, `destructive` for terminal status and tool
  outcomes
- ANSI palette for shell and tool output fallbacks

## How to make a complete theme

1. Pick a `quickStyleOpts` palette.
2. Call `quickStyle(o)` to build the base `Styles`.
3. Override only the fields that genuinely differ from token defaults.
4. Map markdown roles to your palette in `Markdown`.
5. Add a matching `CodeBlock.Chroma` mapping if the theme introduces code
  highlighting colors.
6. If you want a lighter variant, duplicate the theme function and swap base
  roles; keep field names stable so the rest of the UI does not need changes.
7. Add a theme selector hook if the user should choose at runtime.

## Theme authoring checklist

- [ ] All `Styles.*` groups that appear in `styles.go` are covered or intentionally inherited.
- [ ] Status colors are distinguishable in the target terminal profile.
- [ ] Markdown `Document`, `Heading`, `CodeBlock`, `Link`, and `Table` are defined.
- [ ] Bottom resource bar colors remain visible on the chosen backgrounds.
- [ ] ANSI 16 and ANSI 256 fallback colors remain readable if truecolor is unavailable.
- [ ] `ChromaTheme()` still produces valid chroma entries after theme changes.

## Reference

- `internal/ui/styles/quickstyle.go`
- `internal/ui/styles/styles.go`
- `internal/ui/styles/themes.go`
- `./CRUSH/lipgloss/README.md`
- `./CRUSH/lipgloss/UPGRADE_GUIDE_V2.md`
- `./CRUSH/glamour/README.md`
- `./CRUSH/glamour/UPGRADE_GUIDE_V2.md`
- `./CRUSH/glamour/styles/README.md`
- `./CRUSH/glamour/styles/dracula.go`
- `./CRUSH/glamour/styles/tokyo-night.go`

## Future

- [ ] Revisit `charmbracelet/mods` for modal/prompt flows.
