# BVR Beaver Mascot (v1.2.6)

The Beaver is a terminal-native mascot system designed around fixed-width
frames, so animation never moves the surrounding TUI.

## Render sizes

- **Hero (homescreen):** 13×5 cells — the dense plain-ASCII beaver from
  `boot.BeaverFramesDenseAlpha` / `boot.BeaverFramesDenseBeta`, with 6 frame
  keys: `center` (resting), `left-dir`, `right-dir` (driven by mouse gaze OR
  keys/selected text left/right of center), `what` (beaver clicked), `error`
  (agent error), and `confused` (agent loading/thinking).
- **Mini (sidebar):** 12×7 cells for the persistent chat/sidebar identity
  (`MiniMascotFrame`).

The homescreen hero uses the **dense** beaver (`anim.BeaverFrame`) — a compact
13-wide × 5-tall plain-ASCII frame in the v1.1.8–v1.2.1 "donk" style. The
Alpha map holds the normal-expression frames (center, left-dir, right-dir,
what); the Beta map holds the x-ray expressions (error, confused).

The sidebar (chat view) continues to use the larger `MiniMascotFrame`, whose
12 expression variants (center, look-left/right/up/down, laugh, shocked,
dance, chomp, click-boop, music) ride on the same `beaverGaze` state and also
show the X_X x-ray eyes when `beaverErrored` is true.

## Behavior

Mouse motion maps to five gaze buckets: center, left, right, up, and down. A
dead zone around the face prevents twitching. When the pointer is clearly
left or right of the mascot, the hero looks left/right to follow it; over the
mascot it returns to the idle "rest" pose. Clicking the hero's fixed hitbox
randomly plays one of six Beaver cues (`WHAT-FLAT.wav` or
`OH-BEAV-01.wav` through `OH-BEAV-05.wav`), briefly shows the `what` reaction
on the hero, and switches the mini mascot to its `StateClickBoop` expression.

Keys and selected text still drive direction too: text typed left/right of the
prompt center makes the beaver look left/right, and any cursor-moving key
(arrow keys, Home/End, typing) updates the direction.

## The eight frame functions

| Frame key  | Shown when...                                                    |
|------------|------------------------------------------------------------------|
| `center`   | Resting — the neutral idle pose.                                   |
| `left-dir` | Mouse or text is to the LEFT of center (mouse follows; typing, arrows, Home/End). |
| `right-dir`| Mouse or text is to the RIGHT of center (mouse follows; typing, arrows, Home/End). |
| `what`     | The user clicked the beaver mascot (650 ms reaction + SFX).      |
| `error`    | The agent errored (`notify.TypeAgentError`).                     |
| `confused` | The agent is loading or thinking (`isAgentBusy()`).              |

Priority order on the hero: error > what (recent click) > confused (thinking)
> mouse-dir (pointer clearly left/right of mascot) > resting (idle pulse) >
keyboard-dir (text/cursor left/right of center) > center.

## Implementation

- `internal/ui/anim/beaver.go` — `BeaverFrame(key, errored)` (dense hero,
  `normalizeDenseFrame` keeps every pose at 13×5),
  `MiniMascotFrame()` (sidebar) + `LargeMascotFrame()` (legacy, still used by
  the `anim_test.go` footprint validation and the `version_banner.go`
  `--version` mascot), plus `FrameFromDir(dir int) BeaverFrameKey` for
  mouse/keyboard direction frames.
- `internal/ui/boot/boot.go` — `BeaverFramesDenseAlpha` (normal expressions:
  center, left-dir, right-dir, what) and `BeaverFramesDenseBeta` (x-ray
  expressions: error, confused), plus the intro `BootSequence`.
- `internal/ui/model/landing.go` — homescreen rendering: maps `beaverErrored`,
  `beaverBoopUntil`, `isAgentBusy()`, `beaverMouseDir()`, `beaverResting`, and
  `beaverFacing` to a frame key, then calls `anim.BeaverFrame(key,
  m.beaverErrored)`.
- `internal/ui/model/ui.go` — beaver state fields, click detection via
  `beaverRect`, `triggerBeaverBoop()` SFX, `beaverMouseDir()` mapping mouse
  gaze to direction, `handleAgentNotification` (sets `beaverErrored` on
  `TypeAgentError`, clears it on `TypeAgentFinished`), and `sidebarBeaverRect`.
  `sidebarBeaverRect`.
