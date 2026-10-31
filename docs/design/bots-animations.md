# Bot mascot animations

Planning note for the `/bots` picker. Pillo's six large frames and the
compact 11×10 grids are on screen. The large extracted grids below remain
the source for a later swap of Dingus, BroBro, Luvah, and Gnarf. `z` in
the picker still shows the older 16×16 sprites for those four.

Small preview: [`docs/media/bots/small-preview.png`](../media/bots/small-preview.png).
Small Pillo loop: [`docs/media/bots/pillo-small.gif`](../media/bots/pillo-small.gif).
Chat still: [`docs/media/bots/chat-preview.png`](../media/bots/chat-preview.png).

## Source

Richard's combined painting is
[`docs/media/bots/mascots-source.png`](../media/bots/mascots-source.png)
(1536×1024). Individual files such as `resources/img/triangle.png` are
**not on `origin/main` or this branch** (fetched 2026-10-31; `resources/img/`
only has `beaver.png` and `venmo.jpg`). When those files land, they replace
the combined image as the canonical source and these grids get re-extracted.
Until then, the combined painting is the source.

A color preview of the extracted pixels, scaled with nearest-neighbor so
each art pixel stays a square, is
[`docs/media/bots/preview.png`](../media/bots/preview.png).

## Look

NES-flat. One body color, one near-black eye color, no gradients, no
outline. Sprites are drawn with Unicode half-blocks (`▀▄█`) so one
terminal cell holds two vertical pixels. The terminal row count is
`ceil(height / 2)` after padding an odd height with one transparent row.

Color:

- Truecolor when the terminal profile is truecolor and `NO_COLOR` is unset.
- Otherwise the nearest xterm 256-color index.
- `NO_COLOR` or `TERM=dumb`: no SGR. Body cells use a medium shade (`▒`),
  eyes use a heavy block (`█`), so the ovals stay readable on white Dingus.

The selected bot keeps today's dark rounded tile behind the sprite and the
blinking `▶`. That tile is chrome, not part of the mascot grid.

## How the grids were measured

Outline steps on the white circle cluster near 12 source pixels, so each
art pixel is one 12×12 block. A cell is body when at least 42% of that
block is brighter than channel 48. Dark cells that do not touch the crop
border are eyes (holes in the silhouette). A one-cell transparent speck
with six or more body neighbors and no eye neighbor is filled, which
drops sampling nicks without closing the eyes. The hex values are the
median of the inner 6×6 of each cell. The full bitmap's height is not
divisible by 12, so each sprite is aligned from its own crop, not from
pixel (0,0) of the PNG.

Transparent cells are `.` in the text grids. They are not the canvas
color. The painting's empty background samples near `#040404`, which is
also where the eye medians land. Eyes are stored as their own swatch so
they still read when the terminal background is not that black.

White glints are **not** a separate flat pixel in this extraction. The
black ovals are solid. The bright rim is antialiasing where the eye meets
the body. See open questions.

Dingus's orange star badge is **not** in the painting. The grid leaves
it off. The badge stays an overlay, same as today.

## Per-mascot specs

Coordinates are the top-left of the trimmed grid. `eye1` is the left
oval, `eye2` the right. Both eyes on a bot are the same 3×4 oval, top
shifted one column right. The right oval also sits one row higher.

| Bot | File | Size | Body | Eye | Left eye | Right eye |
| --- | --- | --- | --- | --- | --- | --- |
| Dingus | `internal/bots/sprites/dingus.txt` | 16×15 | `#FEFDFE` | `#070606` | 6,3 3×4 | 10,2 3×4 |
| BroBro | `internal/bots/sprites/brobro.txt` | 17×15 | `#2BE26E` | `#040404` | 6,5 3×4 | 10,4 3×4 |
| Pillo | `internal/bots/sprites/pillo.txt` | 17×12 | `#9A9898` | `#050404` | 8,3 3×4 | 12,2 3×4 |
| Luvah | `internal/bots/sprites/luvah.txt` | 16×15 | `#FD3BAC` | `#040404` | 7,4 3×4 | 11,3 3×4 |
| Gnarf | `internal/bots/sprites/gnarf.txt` | 17×14 | `#9159FD` | `#050505` | 6,5 3×4 | 10,4 3×4 |

Small PNGs of the same pixels, one art pixel per file pixel, with a
transparent background:

- `docs/media/bots/dingus.png`
- `docs/media/bots/brobro.png`
- `docs/media/bots/pillo.png`
- `docs/media/bots/luvah.png`
- `docs/media/bots/gnarf.png`

BroBro's triangle points right. Pillo's eyes sit in the right half of
the pill. Luvah's sit in the right lobe. Gnarf's sit low in the rounded
square. The downsample had stepped on the ovals, so each eye was redrawn
by hand to this mask, both eyes identical:

```
.EE
EEE
EEE
EE.
```

Half-block terminal size after the odd-height pad: Dingus and BroBro and
Luvah are 16 or 17 columns by 8 rows. Pillo is 17×6. Gnarf is 17×7.

## Animation plan

The clock stays the dialog's existing 100ms `tea.Tick` (10 fps). Do not
schedule a faster tick. BroBro feels fast because each tick moves more
and the typewriter reveals more runes, not because the terminal redraws
more often. Sounds fire on the first frame of an action, never on a blink
frame, so idle does not click.

Blink frames keep the oval: drop the bottom row of each eye and one
outer column. Do not replace the eye with a horizontal dash.

```
open          squint
.EE           .E.
EEE           EEE
EEE           EE.
EE.
```

Offsets are in art pixels. Positive `dy` moves the body down inside the
tile. The tile does not move.

### Dingus — chill

| Anim | Frames | Timing | What moves | Sound |
| --- | --- | --- | --- | --- |
| Idle blink | 2 | open 500ms, squint 100ms, repeat | Eyes only | none |
| Selected | 1 | tile on immediately; idle continues | Tile + cursor | `quick-notify-01` once, on enter |
| Talk | 2 | squint held for the whole typewriter; 34ms × 1 rune | Eyes stay squinted | `chat-01` or `chat-02` on the first reveal frame |

Error line: hold the squint one extra tick and play `error-01`. No bob.

### BroBro — wild, fast

| Anim | Frames | Timing | What moves | Sound |
| --- | --- | --- | --- | --- |
| Idle | 4 | 100ms each | `dy` 0, 1, 0, 1; eyes squint on frame 3 only | none |
| Selected pop | 3 | 100ms each, then idle | `dy` -1, +1, 0 | `quick-notify-02` on frame 0 |
| Talk | 3 | 100ms each, loop while revealing; typewriter 12ms × 3 runes | `dx` -1, 0, +1 | rotate `chat-01` … `chat-04` on talk start |

Error: `error-02` on the error line. Sprite returns to idle. No extra pose.

### Pillo — shipped

Richard's six rough frames are coded in `internal/bots/pillo_anim.go`,
not shipped as PNGs. Every frame is 17×13, padded on top, so the bottom
row stays put. Eyes are the same tilted oval (a smaller one on the hop).

| Frame | Pose |
| --- | --- |
| 01 | Rest. Eyes up and to the right. |
| 02 | Skeptical. Right eye raised. |
| 03 | Glance up-right. Right eye high, left eye lower. |
| 04 | Glance left and up. Both eyes shifted left. |
| 05 | Squish. Narrower body, eyes lower. |
| 06 | Hop. Taller and narrower, small squint eyes low. |

The clock is the dialog's 100ms tick. Durations are multiples of that.

| Anim | Sequence | Timing | Sound |
| --- | --- | --- | --- |
| `idle` | 01, 03, 01, 04, loop | 800ms, 200ms, 600ms, 200ms | none on the frames. `chat-03` still plays when a reply starts |
| `skeptical` | 02, 01, loop | 300ms, 200ms | none extra. Used while a reply has not drawn yet |
| `talk` | 03, 01, 04, 01, loop | 200ms each | none extra |
| `select` | 05, 06, 01, once | 200ms each | `quick-notify-01` on enter, already |

`BVR_REDUCE_MOTION` or `PREFERS_REDUCED_MOTION` holds frame 01.
Dingus, BroBro, Luvah, and Gnarf stay on their static sprites.

Preview: `docs/media/bots/pillo-anim.gif` and
`docs/media/bots/pillo-frames.png`.

### Luvah — sweet

| Anim | Frames | Timing | What moves | Sound |
| --- | --- | --- | --- | --- |
| Idle blink | 2 | open 500ms, squint 100ms | Eyes only | none |
| Selected beat | 2 | 200ms each, play twice, then idle | `dy` 0 then -1 | `quick-notify-03` on the first upbeat |
| Talk | 2 | squint held while revealing; 42ms × 1 rune | Eyes | `chat-04`, or `chat-01` when the line index is odd |

Error: `error-01`. Hearts stay in the chat emote, not on the sprite.

### Gnarf — loud

| Anim | Frames | Timing | What moves | Sound |
| --- | --- | --- | --- | --- |
| Idle | 2 | 200ms each (two ticks) | `dy` 0, 1; blink every fourth cycle | none |
| Selected | 2 | 100ms each, then idle | `dy` -1, 0 | `quick-notify-01` on frame 0 |
| Talk | 3 | 100ms each, loop; typewriter 18ms × 2 runes | `dy` -1, +1, 0 | `chat-02` or `chat-04`, plus `quick-notify-03` (the boop) on frame 0 |

The boop stays on by default. `m` / `ctrl+m`, audio volume `silent`, and
`NO_SOUND` still suppress every cue, including the boop. Error: `error-02`
or `error-03` on the error line, and no boop.

## Frame data

Keep one base grid per bot, the `.txt` files above. Do not store a full
copy of the body per frame. An animation is a list of ops applied at
draw time:

```
frame:
  dy: 0
  dx: 0
  eyes: open | squint
  hold_ms: 100
  cue: ""          # event name, empty if this frame is silent
```

`eyes: squint` rewrites pixels inside each eye box using the pattern
above, clipped to the box. `dy` / `dx` shift only body and eye pixels.
The selection tile is composited first and does not shift.

Silhouette changes (BroBro's pop is only a shift, so it still fits this
model). If a later pass needs a real squash, add a full frame file
`internal/bots/sprites/<name>-<anim>-<n>.txt` and point the op at it.
Do not add that until a shift is not enough.

Palette stays on the bot: `body` and `eye` hex from the grid header.
Dingus's badge, when it is drawn, is a second overlay using the existing
orange `#F4A23C` and white `#FEFDFE`, placed at the bottom-right of the
16×15 grid, not stored in `dingus.txt`.

## Renderer and ticks

Today:

- `internal/bots/sprite.go` `Render` composites the tile, applies a bob,
  and paints half-blocks for truecolor, 256, and `NO_COLOR`.
- `internal/ui/dialog/bots.go` runs one `tea.Tick` at 100ms while the
  dialog is open and maps `IdleMillis` / `Bob` / `Blink` onto that clock.
- `internal/ui/model/ui.go` plays `ActionPlayCue` through the existing
  audio backend. `chat-01` … `chat-04`, `error-01` … `error-03`, and
  `quick-notify-01` … `quick-notify-03` already map onto
  `audio/sounds/*-FLAT.wav`.

Next pass, still inside that tick:

1. Build the `Frame` from the `.txt` grid instead of `roster.go`.
2. Replace `pose()` with the op list above. One tick is 100ms. A 500ms
   hold is five ticks.
3. Attach the cue to the frame that starts the action. The dialog already
   drops cues when muted, when the volume is `silent`, or when `NO_SOUND`
   is set. Do not call the audio stack from the sprite package.
4. Stop the tick when the dialog closes. Do not tick the main chat.

Odd heights: pad one transparent row at the bottom before pairing
half-blocks.

## Performance

Five sprites at about 17×16 pixels are a few hundred cells. The cost is
the full Bubble Tea view, which already redraws on every tick. Cap the
bots dialog at 10 fps (100ms). Do not go below 50ms even for BroBro.
No per-bot timer. No animation when the dialog is not in front.

## Accessibility

If `BVR_REDUCE_MOTION` or `PREFERS_REDUCED_MOTION` is non-empty:

- Sprites stay on frame 0 (eyes open, `dy` 0, `dx` 0).
- The selection tile and `▶` still show. The cursor does not blink.
- Typewriter reveals the whole line at once.
- Sounds still play, and still respect mute / `silent` / `NO_SOUND`.

`NO_COLOR` changes the glyph, not the motion.

## Tests

When the grids are wired in:

- Golden half-block dump per bot, `ColorNone`, open eyes, no bob.
  Replace today's goldens in `internal/bots/testdata/`; those match the
  hand-authored frames.
- A second golden for the squint frame. Assert every squint eye pixel
  sits inside the open eye box.
- Palette test: parsed `body` / `eye` hex equals the table above.
- Two eye components per bot, each at least 4 pixels, right eye origin
  `x` greater than the left.
- Reduce-motion: after several ticks, the rendered frame equals frame 0.
- Cue test: talk start emits the chat event; blink emits nothing; Gnarf
  talk start also emits `quick-notify-03`; mute emits nothing.

Do not golden the whole dialog until the runtime sprite swap lands.
`docs/media/bots/preview.png` is the color reference for this plan.

## Open questions

1. `resources/img/*.png` is not pushed. Confirm the filenames (the example
   given was `triangle.png`) and whether each file is already one art
   pixel per file pixel or another scaled painting.
2. Add a 1px white glint in the upper-inner corner of each eye, or keep
   the solid black ovals the extraction produced?
3. Dingus badge: keep the current orange star as an overlay, and at what
   size? It is not in the painting.
4. BroBro points right. Is that the facing we should ship?
5. Is 100ms the right ceiling, or do you want BroBro's talk at 50ms?
6. Should reduce-motion also silence the typewriter only, or the cues too?
   This plan keeps the cues.
7. Idle sound: this plan plays nothing on blinks and bobs. Confirm.

## Checklist

- [x] Commit the combined painting and the extracted grids.
- [x] Color preview at `docs/media/bots/preview.png`.
- [x] Write this plan. Do not animate yet.
- [ ] Re-extract from `resources/img/*.png` when those files are on main.
- [ ] Point `internal/bots` at the grids. Leave animation ops as a no-op
      so the picker matches the preview before any motion lands.
- [ ] Idle blink, then bob for BroBro and Gnarf.
- [ ] Selected poses and the enter cues.
- [ ] Talk poses, typewriter speeds, chat cues, Gnarf boop on frame 0.
- [ ] Reduce-motion.
- [ ] Swap goldens and add the eye-box tests.
- [ ] Dingus badge overlay, after question 3 is answered.
