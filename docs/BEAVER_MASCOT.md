# BVR Beaver Mascot (v1.2.7)

The Beaver is a terminal-native mascot system designed around fixed-width
frames, so animation never moves the surrounding TUI.

## One mascot, two screens

As of **1.2.7 the homescreen and the working screen show the same animated
BVR**. There is no longer a separate "hero" beaver and "sidebar" beaver.

| Screen | Renderer | Footprint |
| --- | --- | --- |
| Homescreen / landing | `anim.BeaverFrame(key, errored)` | 13×5 |
| Working screen / sidebar | `anim.BeaverFrame(key, errored)` | 13×5 |

Both call `(*UI).beaverFrameKey()` in `internal/ui/model/ui.go`, which is the
single source of truth for pose selection. This guarantees the two screens can
never drift apart: same priority order, same reactions, same mouse tracking.
`TestBeaverFrameSharedAcrossScreens` asserts byte-identical output across every
frame key.

### What changed in 1.2.7

Before this release the sidebar used `MiniMascotFrame` — a 12×7 mascot with a
*different, reduced* expression set (center, look-left/right/up/down, laugh,
shocked, dance, chomp, click-boop, music). That path is gone from the TUI. The
helpers `MiniMascotFrame` and `LargeMascotFrame` remain in `internal/ui/anim`
and are still covered by `anim_test.go`, but nothing in the UI renders them.

`LargeMascotFrame` is still used by `version_banner.go` for `--version`.

## Render sizes

- **Dense (homescreen and sidebar):** 13×5 cells — the plain-ASCII beaver from
  `boot.BeaverFramesDenseAlpha` / `boot.BeaverFramesDenseBeta`, with 6 frame
  keys: `center` (resting), `left-dir`, `right-dir` (driven by mouse gaze OR
  keys/selected text left/right of center), `what` (beaver clicked), `error`
  (agent error), and `confused` (agent loading/thinking).
- **Large (legacy):** 24×8 via `LargeMascotFrame`, used only by the version
  banner and the anim footprint tests.

The Alpha map holds the normal-expression frames (center, left-dir, right-dir,
what); the Beta map holds the x-ray expressions (error, confused). The boot demo
(`boot.Run`) cycles through the full frame set, so every expression is shown
during startup.

### Frame lookup across the Alpha and Beta maps

The two maps hold **disjoint** poses, so `BeaverFrame` consults both:

- When `errored` is true the Beta (x-ray) map wins, then Alpha.
- When `errored` is false the Alpha map wins, then Beta.
- Only if neither map has the key does it fall back to Alpha `center`.

This matters. An earlier version searched a single map based on `errored`,
which made `confused` and `error` unreachable whenever `errored` was false — the
agent-busy expression silently rendered as the resting face. `TestBeaverFrameEveryKeyResolves`
and `TestBeaverFrameConfusedRendersWhenNotErrored` guard this.

## Behavior

Mouse motion maps to five gaze buckets: center, left, right, up, and down. A
dead zone around the face prevents twitching. When the pointer is clearly left
or right of the mascot, the beaver looks left/right to follow it; over the
mascot it returns to the idle "rest" pose.

Gaze resolves against **whichever mascot is on screen**: `beaverRect` on the
homescreen, `sidebarBeaverRect` on the working screen. Both go through
`(*UI).visibleBeaverRect()`.

Clicking the mascot's hitbox randomly plays one of six Beaver cues
(`WHAT-FLAT.wav` or `OH-BEAV-01.wav` through `OH-BEAV-05.wav`) and briefly shows
the `what` reaction for 650 ms. This works on both screens. The hitbox is sized
from the exact frame string that was composed into the logo
(`sidebarBeaverFrame`), so it always matches the art on screen, and it is
**cleared when the sidebar is too short to draw the mascot** so clicks in empty
space cannot boop an invisible mascot.

Keys and selected text still drive direction too: text typed left/right of the
prompt center makes the beaver look left/right, and any cursor-moving key
(arrow keys, Home/End, typing) updates the direction.

## The six frame functions

| Frame key  | Shown when...                                                    |
|------------|------------------------------------------------------------------|
| `center`   | Resting — the neutral idle pose.                                   |
| `left-dir` | Mouse or text is to the LEFT of center (mouse follows; typing, arrows, Home/End). |
| `right-dir`| Mouse or text is to the RIGHT of center (mouse follows; typing, arrows, Home/End). |
| `what`     | The user clicked the beaver mascot (650 ms reaction + SFX).      |
| `error`    | The agent errored (`notify.TypeAgentError`).                     |
| `confused` | The agent is loading or thinking (`isAgentBusy()`).              |

Priority order, identical on both screens:

```
error > what (recent click) > confused (thinking)
  > mouse-dir (pointer clearly left/right of mascot)
  > resting (idle pulse)
  > keyboard-dir (text/cursor left/right of center)
  > center
```

## Implementation

- `internal/ui/anim/beaver.go` — `BeaverFrame(key, errored)` plus
  `lookupBeaverFrame`, `normalizeDenseFrame` (keeps every pose at 13×5),
  `DenseFrameWidth` / `DenseFrameHeight`, and `FrameFromDir(dir int)`
  for mouse/keyboard direction frames.
- `internal/ui/boot/boot.go` — `BeaverFramesDenseAlpha` (center, left-dir,
  right-dir, what), `BeaverFramesDenseBeta` (error, confused), and `BootSequence`.
- `internal/ui/model/ui.go` — `beaverFrameKey()` (shared pose resolver),
  `beaverFrame()`, `visibleBeaverRect()`, beaver state fields,
  `triggerBeaverBoop()` SFX, `beaverMouseDir()`, `handleAgentNotification`,
  and `cacheSidebarLogo()` which renders the sidebar mascot into
  `sidebarBeaverFrame`.
- `internal/ui/model/landing.go` — homescreen hero, now just
  `m.beaverFrame()` plus hitbox layout.
- `internal/ui/model/sidebar.go` — sizes `sidebarBeaverRect` from the cached
  frame and clears it when the mascot is not drawn.

### Tests

`internal/ui/model/ui_test.go`:
`TestBeaverFrameSharedAcrossScreens`, `TestSidebarBeaverFrameUsesDenseMascot`,
`TestVisibleBeaverRectFollowsState`, `TestBeaverClickBoop`,
`TestBeaverMouseDir`, `TestBeaverMouseGazeFromMotion`.

`internal/ui/anim/anim_test.go`:
`TestBeaverFrameEveryKeyResolves`, `TestBeaverFrameConfusedRendersWhenNotErrored`,
`TestBeaverFrameErrorTakesPriorityWhenErrored`, `TestBeaverFrameKeepsDenseFootprint`,
`TestFrameFromDir`.
