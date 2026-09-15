# BVR Beaver Mascot (v1.2.4)

The Beaver is a terminal-native mascot system designed around fixed-width
frames, so animation never moves the surrounding TUI.

## Render sizes

- **Hero:** 24×12 cells for the homescreen and startup/loading moments.
- **Mini:** 12×7 cells for the persistent chat/sidebar identity.

Both sizes use the same character anchors: rounded ears, X-shaped eyes, nose,
buck teeth, belly, raised paw, textured tail, feet, and ground line.

## Behavior

Mouse motion maps to five gaze buckets: center, left, right, up, and down. A
dead zone around the face prevents twitching. Clicking the hero's fixed hitbox
randomly plays one of six Beaver cues (`WHAT-FLAT.wav` or
`OH-BEAV-01.wav` through `OH-BEAV-05.wav`) and briefly switches to the click
expression before the current gaze returns. The same interaction is available
on the chat sidebar's mini Beaver.

The state vocabulary already includes laugh, shocked, dance, chomp, music, and
boop expressions. Transient expressions are rendered independently from the
passive gaze state so future audio and event triggers do not need to mutate the
frame renderer.

Implementation lives in `internal/ui/anim/beaver.go`, with homescreen wiring in
`internal/ui/model/landing.go` and `internal/ui/model/ui.go`.

## Terminal safety

Every state is normalized to its canonical footprint. Frames are rendered row
by row to keep ANSI styling from collapsing multiline art in screen buffers.
The renderer is ASCII-first and does not depend on emoji or a particular font.
