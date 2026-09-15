# BVR Beaver Mascot (v1.2.4)

The Beaver is a terminal-native mascot system designed around fixed-width
frames, so animation never moves the surrounding TUI.

## Render sizes

- **Hero (homescreen):** 13×5 cells — the dense character-filled ASCII beaver
  from `boot.BeaverFramesDenseAlpha`, with center/left/right facing poses.
- **Mini (sidebar):** 12×7 cells for the persistent chat/sidebar identity
  (`MiniMascotFrame`).

The homescreen hero uses the **dense** beaver (`anim.BeaverFrame`) — a
compact 13-wide × 5-tall frame textured with `A, V, @@, H, AW, WX` fillers
and a `0_0` Alpha face. An x-ray `X_X` **Beta** variant is shown when the
agent errors (`beaverErrored` flag), returning to the normal Alpha variant
on completion.

The sidebar (chat view) continues to use the larger `MiniMascotFrame`
which includes idle, blink, and working poses.

## Behavior

Mouse motion maps to five gaze buckets: center, left, right, up, and down. A
dead zone around the face prevents twitching. Clicking the hero's fixed hitbox
randomly plays one of six Beaver cues (`WHAT-FLAT.wav` or
`OH-BEAV-01.wav` through `OH-BEAV-05.wav`) and briefly switches to the
click expression before the current gaze returns. The same interaction is
available on the chat sidebar's mini Beaver.

The `beaverFacing` field (-1/0/+1) tracks the last applied direction and is
debounced through a `hoverSettleMsg` so the mascot does not flip on every
micro mouse movement (4s pulse + 400ms hover debounce).

State fields on the `UI` struct:

| Field             | Type                | Purpose                                      |
|-------------------|---------------------|----------------------------------------------|
| `beaverFacing`    | `int`               | 0=center, -1=left, +1=right                   |
| `beaverErrored`   | `bool`              | Show x-ray Beta variant when agent errors     |
| `beaverResting`   | `bool`              | Idle "rest" (center) pose between direction changes |
| `beaverGaze`      | `anim.MascotState`  | Gaze direction for the sidebar mini beaver    |
| `beaverRect`      | `image.Rectangle`   | Click hitbox for the homescreen hero          |
| `beaverBoopUntil` | `time.Time`         | Boop cooldown timer (650ms)                   |

The dense beaver (`BeaverFrame`) ignores `beaverBoopUntil` and `beaverGaze` —
it renders purely from `beaverFacing`, `beaverErrored`, and `beaverResting`.
The boop SFX still plays on click; the dense beaver has no separate boop
frame, so it stays in its current facing pose during the 650ms cooldown.

The state vocabulary already includes laugh, shocked, dance, chomp, music, and
boop expressions. Transient expressions are rendered independently from the
passive gaze state so future audio and event triggers do not need to mutate the
frame renderer.

## Terminal safety

Every state is normalized to its canonical footprint. Frames are rendered row
by row to keep ANSI styling from collapsing multiline art in screen buffers.
The renderer is ASCII-first and does not depend on emoji or a particular font.

## Implementation

- `internal/ui/anim/beaver.go` — `BeaverFrame()` (dense hero) +
  `MiniMascotFrame()` (sidebar) + `LargeMascotFrame()` (legacy, still used
  by `anim_test.go` footprint validation and the `version_banner.go`
  `--version` mascot).
- `internal/ui/boot/boot.go` — `BeaverFramesDenseAlpha` (normal `0_0`
  faces) and `BeaverFramesDenseBeta` (x-ray `X_X` faces), each with
  center/left/right poses.
- `internal/ui/model/landing.go` — homescreen rendering using
  `BeaverFrame(m.beaverFacing, m.beaverErrored, m.beaverResting)`.
- `internal/ui/model/ui.go` — beaver state fields, click detection via
  `beaverRect`, `triggerBeaverBoop()` SFX, and `sidebarBeaverRect`.
