# BVR-CLI UI Layout Analysis

## Overall Structure (Vertical, Top to Bottom)

```
┌─────────────────────────────────────────────────────┐
│ HEADER (7 rows)                                      │
│ Logo, Version, Banner Animation                      │
├─────────────────────────────────────────────────────┤
│ MAIN AREA (fills remaining space)                    │
│   CWD line (1 row)                                   │
│   [blank] (1 row)                                    │
│   Hero/Beaver mascot image (8 rows)                  │
│   [blank] (1 row)                                    │
│   BUTTONS row (1 row) ← FIX: needs more space     │
│   [blank] (1 row)                                    │
│   MODEL/PROVIDER line (1 row) ← FIX: down a space │
│   [blank] (1 row)                                    │
│   GGWAVE waveform (2-3 rows)                         │
├─────────────────────────────────────────────────────┤
│ LEFT COLUMN (~30 cols) │ RIGHT COLUMN (~full)      │
│ NODE, MCP, LSP status    │ SKILLS list              │
├─────────────────────────────────────────────────────┤
│ EDITOR AREA (textarea + attachments)                 │
├─────────────────────────────────────────────────────┤
│ STATUS / HELP (1-2 rows)                             │
│ Help keybindings text                                │
├─────────────────────────────────────────────────────┤
│ RESOURCE MONITOR (1 row) ← MOVE TO BOTTOM          │
│ CPU bar | RAM bar                                    │
└─────────────────────────────────────────────────────┘
```

## Layout Rectangles (from `generateLayout`)

| Rectangle | Description | Rows |
|-----------|-------------|------|
| `headerRect` | Top 7 rows, logo/version | ~7 |
| `mainRect` | Everything between header and editor | varies |
| `editorRect` | Textarea + attachments | ~5-10 |
| `statusRect` (`helpRect`) | Bottom 2 rows, help text | ~2 |
| `resourceRect` | 1 row below status, CPU/RAM bars | ~1 |
| `appRect` | Main content area (margins applied) | varies |

## Landing Page Components (`landingView`)

The `infoSection` is a vertical join of these parts:

```
Part 0: cwdStyled          → 1 row (styled working directory)
Part 1: "" (blank)         → 1 row
Part 2: hero               → 8 rows (beaver mascot image)
Part 3: "" (blank)         → 1 row
Part 4: buttons            → 1 row (4 buttons side by side)
Part 5: "" (blank)         → 1 row
Part 6: modelLine          → 1 row (MODEL + PROVIDER)
Part 7: "" (blank)         → 1 row
Part 8: waveLabel + wave   → 2-3 rows (ggwave waveform)
```

## Footer Area (Bottom of Screen)

### Current Layout (top to bottom):
```
┌─────────────────────────────────────┐
│ STATUS / HELP ← helpRect (2 rows)  │
│ Help keybindings                    │
├─────────────────────────────────────┤
│ RESOURCE ← resourceRect (1 row)    │
│ CPU ██████ 12% | RAM ████████ 25% │
└─────────────────────────────────────┘
```

### Requested Layout (after changes):
```
┌─────────────────────────────────────┐
│ STATUS / HELP ← helpRect           │
│ Help keybindings                    │
├─────────────────────────────────────┤
│                                     │
│ ← 1 blank row gap                   │
│                                     │
├─────────────────────────────────────┤
│ RESOURCE ← resourceRect (bottom)   │
│ CPU ██████ 12% | RAM ████████ 25% │
└─────────────────────────────────────┘
```

## Horizontal Space Breakdown

For a typical terminal of ~120 cols width:

| Element | Cols | % of Width | Notes |
|---------|------|------------|-------|
| Left sidebar | 32 | 26.7% | NODE, MCP, LSP (chat mode) |
| Main content | ~86 | 71.7% | Messages/landing content |
| Margins | 2 | 1.7% | Left/right padding |

## Key Files

| File | Purpose |
|------|---------|
| `internal/ui/model/landing.go` | Landing page layout, buttons, hero, ggwave |
| `internal/ui/model/ui.go` | Layout generation, draw dispatch |
| `internal/ui/model/resource_monitor.go` | CPU/RAM bars rendering |
| `internal/ui/model/status.go` | Help/status bar rendering |
| `internal/ui/model/header.go` | Header/logo rendering |
| `internal/ggwave/waveform.go` | ggwave visual waveform |
| `internal/ui/model/keys.go` | Key bindings |
| `internal/ui/dialog/commands.go` | Command palette items |

## Fix List

1. **Button Spacing**: Add more vertical space around buttons row
2. **Model/Provider**: Move MODEL/PROVIDER line down a space
3. **CPU/RAM to Bottom**: Move resource monitor to very bottom of UI
4. **Footer Text**: Add more text to the footer area

## ggwave Animation Speed
- `ggwaveSpeed` field (default `0.5`) slows the waveform animation by half
- Applied in `landingView()` via `ggwave.Waveform(waveWidth, m.ggwaveFrame, m.ggwaveMode, m.ggwaveSpeed)`
