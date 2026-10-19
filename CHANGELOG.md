# Changelog
## 1.2.7

### New!

- **The same animated BVR now appears on the working screen.** The sidebar used
  to render a separate 12×7 "mini" mascot with a reduced expression set while the
  homescreen used the dense 13×5 hero. Both screens now render the identical
  dense beaver and resolve their pose through one shared
  `(*UI).beaverFrameKey()`, so the two can never drift apart.
  `TestBeaverFrameSharedAcrossScreens` asserts byte-identical output across every
  frame key.
- **GGWave proximity node pairing.** New `ggwave://` transport for audio-based
  device bootstrap: pairing envelope, confirmation dialog, and node
  registration. See `docs/GGWAVE.md` and `docs/NODE_TRANSPORTS.md`.
- **Refreshed screenshots.** New full working-screen, command palette, file
  finder, create-file, image recognition, and header captures. Every screenshot
  now uses one `UPPER_SNAKE_CASE` naming convention.

### Fixed

- **The agent-busy "confused" mascot expression never actually rendered.**
  `confused` and `error` live only in the Beta (x-ray) frame map, but
  `BeaverFrame` searched a single map chosen by the `errored` flag. With
  `errored` false — which is the case while the agent is simply thinking — both
  keys missed and silently fell back to the resting face. Both maps are now
  consulted with an `errored` priority order.
- **Mouse gaze only tracked the homescreen.** Gaze tracking was gated on
  `state == uiLanding`, so the sidebar mascot could not follow the pointer. Gaze
  and click hit-testing now resolve through `visibleBeaverRect()`, which picks
  whichever mascot is actually on screen.
- **Clicks in the short-sidebar layout booped an invisible mascot.** When the
  sidebar is below the height breakpoint the mascot is not drawn, but a hitbox
  was still built. The hitbox is now cleared, and it is sized from the exact
  frame string composed into the logo instead of a re-render that could drift
  from the drawn art.
- **`CGO_ENABLED=0` builds were broken by the GGWave audio work.** The PortAudio
  backend had no build tag, so it imported PortAudio unconditionally and every
  CGO-free build failed. Gating it exposed a second failure — the audio transport
  had no stub backend — which is now added. All four build combinations compile:
  `CGO_ENABLED` 0/1, with and without `-tags=ggwave`.
- **The GGWave stub and native adapters had diverged.** The stub declared
  `Encode`/`Decode` over `[]float32` while the native adapter used `[]int16`, so
  any caller written against the stub would fail to build with the native backend
  compiled in. Signatures are now aligned and the float32 helpers mirror
  `EncodeFloat32`/`DecodeFloat32`.
- **PortAudio was initialized in an `init()` function**, which cannot report an
  error and fired on package import even when no audio was requested.
  Initialization is now lazy and memoized, surfacing failures as a normal error
  from `NewTransport`.
- **Repaired the `internal/catalog` FTS index, which was permanently empty.**
  Two bugs were masked by a discarded error (`_, _ =`): the insert statement
  declared 9 placeholders for 10 columns, and `ON CONFLICT ... DO UPDATE` is not
  implemented for FTS5 virtual tables. A third surfaced once those were fixed:
  the FTS query combined terms with implicit `AND`, collapsing recall to near
  zero, so ordinary multi-signal searches matched nothing. Now an explicit
  transactional delete+insert, `OR` semantics, and quoted terms.

### Testing

- `TestBeaverFrameEveryKeyResolves`, `TestBeaverFrameConfusedRendersWhenNotErrored`,
  `TestBeaverFrameErrorTakesPriorityWhenErrored`, `TestBeaverFrameKeepsDenseFootprint`
  (ui/anim) covering frame resolution across both frame maps and the dense grid.
- `TestBeaverFrameSharedAcrossScreens`, `TestSidebarBeaverFrameUsesDenseMascot`,
  `TestVisibleBeaverRectFollowsState` (ui/model) covering cross-screen parity and
  hitbox behavior.
- `TestInsertProduct_PopulatesFTSIndex`, `TestInsertProduct_ReplacesFTSRow`,
  `TestFTSQuery_CombinesTermsWithOr`, `TestFTSQuery_QuotesTermsWithPunctuation`,
  `TestFTSQuery_EscapesEmbeddedQuotes`, `TestFTSQuery_HandlesBareTerms`
  (catalog) covering the FTS write path and query construction.

### Documentation

- Consolidated every ignored markdown document into `docs/ignore/`, which carries
  a README explaining the convention and how to promote a document back into the
  published set. `CONTEXT.md`, `docs/REMOVAL.md`, `docs/LAUNCH_TASKLIST.md`,
  `docs/ASSET_REQUIREMENTS.md`, `antigravity/rules.md`, the Crush/Cline roadmap,
  the Phase 0 audit, and ADR 0001 moved there.
- `docs/IMG-RESOURCES.md` now matches the committed screenshots, including the new
  captures and the removals.
- `docs/BEAVER_MASCOT.md` rewritten for the single shared mascot, the Alpha/Beta
  frame-map lookup, and the click/gaze contract.
- `docs/GGWAVE.md` documents the full build matrix, why both stubs are required,
  and the stub/native API parity rule.

## 1.2.6

### New!

- **Dense mascot follows the mouse.** The 13×5 homescreen beaver now tracks your cursor: the pointer left of it makes it look left, right of it makes it look right, and directly over it returns to the idle "rest" pose.
- **"What?" reaction on click.** Left-clicking the beaver plays a random Beaver sound cue and flashes its `what` expression for 650 ms; the sidebar mascot simultaneously switches to its click-boop expression.
- **Keyboard direction follows all cursor keys.** The beaver now turns for any key that moves the cursor (typed text, arrows, Home/End, page navigation), not just printable keys.
- **Model switch search no longer swallows keys.** `r`, `p`, `s`, and `c` now work as `ctrl+r` (refresh), `ctrl+p` (pull), `ctrl+s` (start Ollama), and `ctrl+c` (cancel pull), leaving the model-search box free for normal typing.
- **New installers for package testing.** macOS DMGs (darwin_amd64/arm64), Linux
  binaries (linux_amd64/arm64), and Arch Linux packages
  (bvr-cli_1.2.6-1_x86_64/1_arm64.pkg.tar.zst) built locally; Windows EXE/MSI
  and deb/rpm require goreleaser/CI or a capable Windows/Linux machine.

### Fixed

- Fixed mascot frame priority so the mouse direction takes precedence over the rest/idle pose, resolving cases where the beaver ignored cursor movement.
- Rebuilt the dense 13×5 beaver frames with uniform padding so the boot splash
  and homescreen hero no longer wobble between expressions; every row is now
  padded to exactly 13 cells and each pose is internally aligned.
- Extended the boot demo to cycle through the full frame set (directional faces
  plus error, left-dir, right-dir, what, and confused), so every mascot
  expression is shown at startup.

### Testing

- Added `TestBeaverClickBoop`, `TestBeaverMouseDir`, `TestBeaverMouseGazeFromMotion` (ui/model), `TestFrameFromDir` (ui/anim) covering the click reaction, gaze mapping, mouse motion, and frame selection.

## 1.2.5

### Fixes

- **Restored dense beaver logo on homescreen.** Switched the homescreen mascot
  back from `LargeMascotFrame` (8-row "stupid face") to the dense
  `BeaverFrame` (13×5 ASCII beaver from `boot.BeaverFramesDenseAlpha`),
  matching the v1.1.8–v1.2.1 era look. Beaver boop SFX and click detection
  preserved; the dense beaver renders from `beaverFacing`/`beaverErrored`/
  `beaverResting` and has no separate boop frame.
- **Wired IMAGE RECOGNITION button.** The button on the landing page was
  opening the web browser (`openBrowserDialog`) instead of performing image
  recognition. It now opens the file browser with an `imageLookupPending` flag;
  selecting a file runs `bvr lookup-image <path>` via a shell session. Added
  `ctrl+b` keybinding (was only a label on the button before).
- **Added SFX to all command palette items.** Every menu item and dialog
  open/close that was missing audio now has a distinct sound cue: Toggle Beast
  Mode (`chainsaw`), Code Mode, Help, External Editor, Compact Mode, Pills,
  Thinking, Transparent BG, Quit (`exit`), Docker MCP, Initialize, Cline Key,
  Reasoning Effort, Permission Response, Summarize, and all `openDialog`
  cases (Reasoning, Notifications, NODE, Ollama How-To, Themes, Sounds, Other
  Models, Quit).

### Documentation

- Updated `docs/BEAVER_MASCOT.md` with dense beaver dimensions and state table.
- Updated `README.md` shortcuts table (ctrl+b for image recognition, ctrl+o
  for File Finder).
- Added `docs/COMMAND-MENU.md` as the canonical command palette reference.
- Archived `docs/CLINE_FEATURE.md`; canonical provider doc is now
  `docs/CLINE_PROVIDER.md`.
- Trimmed stale beta invite list and personal contact info from
  `docs/BETA.md`.

## 1.2.4

### Image lookup and product catalog

- Added `internal/catalog/` package for local product catalog with SQLite FTS5 search and embedding similarity matching (`embedding.go`, `repository.go`, `repository_test.go`, `schema.sql`).
- Added `internal/cmd/lookup-image` command for identifying handbags, purses, sunglasses, and accessories from photos using local Ollama vision models and a local product catalog.
- Added `internal/imagelookup/` package with vision service, search, rendering, reranking, resizing, validation, and prompt modules (`ollama.go`, `search.go`, `render.go`, `rerank.go`, `resize.go`, `validate.go`, `prompt.go`, `types.go`, `service.go`, and their tests).

### Fixes

- **Restored dense beaver logo on homescreen.** Switched the homescreen mascot
  back from `LargeMascotFrame` (8-row "stupid face") to the original
  `BeaverFrame` (13×5 dense ASCII beaver from `boot.BeaverFramesDenseAlpha`),
  matching the v1.1.8–v1.2.1 era look. Beaver boop SFX and click detection
  preserved; the dense beaver renders from `beaverFacing`/`beaverErrored`/
  `beaverResting` and has no separate boop frame.
- **Wired IMAGE RECOGNITION button.** The button on the landing page was
  opening the web browser (`openBrowserDialog`) instead of performing image
  recognition. It now opens the file browser with an `imageLookupPending` flag;
  selecting a file runs `bvr lookup-image <path>` via a shell session. Added
  `ctrl+b` keybinding (was only a label on the button before).
- **Added SFX to all command palette items.** Every menu item and dialog
  open/close that was missing audio now has a distinct sound cue: Toggle Beast
  Mode (`chainsaw`), Code Mode, Help, External Editor, Compact Mode, Pills,
  Thinking, Transparent BG, Quit (`exit`), Docker MCP, Initialize, Cline Key,
  Reasoning Effort, Permission Response, Summarize, and all `openDialog`
  cases (Reasoning, Notifications, NODE, Ollama How-To, Themes, Sounds, Other
  Models, Quit).

## 1.2.3

## 1.2.3

### Safer audio defaults and branding

- Set new-install audio to a headphone-friendly 25% default.
- Added 50%, 75%, 100%, and Silent choices to the Sounds menu while keeping
  legacy `high`/`low` configuration aliases compatible.
- Added native volume handling for macOS, Linux, and Windows playback paths.
- Replaced the solid-block `--version` mascot with readable ASCII Beaver line
  art closer to the homescreen identity.
- Documented the fixed-size animated hero, persistent mini mascot, gaze states,
  click/BOOP behavior, and audio safety controls.

## 1.2.2

### TUI branding and navigation

- Added a larger fixed-width animated Beaver hero with idle, blink, working,
  and glance poses.
- Moved the hero into the upper landing experience and wide chat sidebar.
- Kept homescreen button hitboxes aligned with the rendered hero height.
- Added extra spacing between the Web Browser icon and its label.

### Audio and GGWave

- Replaced stale sound metadata with the current processed WAV inventory.
- Added distinct quick-notify cues for File Finder, Create File, and Web
  Browser launch actions while preserving existing interaction sounds.
- Added a portable GGWave pairing envelope and animated `GGWAVE NODE LINK`
  waveform surface without making the default CGO-free build depend on native
  audio libraries.

### Release and documentation

- Bumped runtime, packaging, Winget, installer, and task metadata to `1.2.2`.
- Reworked the root README around the actual product workflow and two primary
  screenshots.
- Updated audio, onboarding, features, image inventory, beta, and NODE/GGWave
  documentation to match the current implementation.

## 1.1.9

### Stability & polish
- **Beaver mascot no longer spazzes**: 4s pulse + 400ms hover debounce
  (new `hoverSettleMsg` and `beaverFacing` field) so the mascot tracks the
  cursor/prompt direction smoothly instead of flipping on every mouse
  movement. `anim.BeaverFrame` now takes a direction (-1/0/+1) instead of
  raw coordinates.
- **Duplicate NODE section removed**: NODE/MCP/LSP only appears in the
  left column of the bottom resource bar, giving the beaver its own lane.
- **`--version` beaver in BVR neon green (#3BF66B)**: the version mascot
  is now in the brand green instead of the theme Dolly pink.
- **`docs/TASKLIST.md` no longer broadcast**: added to `.gitignore` and
  `git rm --cached`'d; the file stays on disk for local notes.
- **Project organization**: removed duplicate `Taskfile.yml`, added
  `packaging/icons/` (`bvr-cli-1024.png` + `.icon` set) for app/installer
  integration.

## 1.1.8

### Beaver mascot & UI stability
- **Fixed mascot spazzing**: slowed pulse timer to 4s + added 400ms hover
  debounce so the beaver tracks cursor/prompt direction smoothly without rapid
  flipping on every micro mouse movement. The mascot's facing is now
  debounced through a new `hoverSettleMsg` and a `beaverFacing` field, with
  `anim.BeaverFrame` taking a direction (-1/0/+1) instead of raw coordinates.
- **Duplicate NODE section removed**: fixed the homescreen layout so the
  NODE/MCP/LSP monitors appear only in the left column of the bottom
  resource bar, giving the beaver mascot its own dedicated lane.
- **Version mascot recolored to BVR neon green (#3BF66B)**: the
  `--version` beaver head now renders in the brand green rather than the
  theme Dolly pink, matching the rest of the brand palette.
- **Layered icons added**: high-resolution `bvr-cli-1024.png` (1024×1024
  PNG) and a proper macOS `.icon` set under `packaging/icons/` for app
  bundle and installer integration.

### Project organization
- Removed the duplicate `Taskfile.yml`; consolidated into `Taskfile.yaml`.
- Added `packaging/icons/` to keep brand assets with the rest of the
  packaging bundle, leaving a cleaner project root.

### Brand rename & animated mascot
- **Total rename** `DONK-CLI` → `BVR-CLI` across the module path
  (`github.com/richavery/bvr-cli`), binary, config (`bvr.json`, `.bvr` data dir),
  env vars, all identifiers, docs, assets, packaging, and the iOS companion app
  (`DonkCompanion` → `BvrCompanion`).
- **Animated beaver mascot on the homescreen** (`internal/ui/anim/beaver.go` +
  `landingView`), idling in step with the existing `bannerFrame` ticker — no
  extra tick loop. Reuses the boot splash frames and neon-green (`#39f66b`)
  styling so it matches the intro.
- **Dense character-filled mascot**: the beaver is now textured with fillers
  (`A, V, @@, H, AW, WX`) instead of hollow outlines. An x-ray (`X_X`) **Beta**
  variant is shown when the agent errors
  (`notify.TypeAgentError`), returning to the normal (`0_0`) **Alpha** variant on
  completion (`notify.TypeAgentFinished`) via a new `beaverErrored` flag.

## 1.1.7

### Homescreen UI & animations

- New boot brand banner plays a sequence at startup (each element held ~2s):
  the wordmark **BVR**, the version (`v1.1.7`), **OH BEAV!**, and the
  attribution `created by RICHARD AIZEN AVERY III`. Letters scramble into place
  with a slow fade; the bar uses the theme **ACCENT** color for its background
  and the app **BACKGROUND** color for its text so it pops
  (`internal/ui/model/version_banner.go`).
- Project location on the homescreen is now **bold**, **underlined**, and
  rendered in the accent color for readability against the dark background.
- Added a prominent **MODEL / PROVIDER** line directly on the homescreen
  (accent values, alt labels).
- Added Command palette and File Finder buttons on the homescreen, each prefixed
  with a superfile icon glyph (a terminal icon for Commands, a folder icon for
  File Finder) and themed to the active theme Primary color. The buttons are
  stacked vertically with a blank line between them, left-aligned one above the
  other.
- CPU/RAM resource monitors now use the active **ACCENT** color (e.g. pink bars
  on the green theme).
- Lightened the section heading / help text under LOCAL DEVICE, SKILLS, and
  MCP so it reads against the dark surface.

### Global theme & color mapping

- Introduced per-theme **Accent** and **Alt** colors, exposed as
  `Styles.ThemeColor.Accent` / `.Alt`, following the mapping table:
  Green→(Pink,Purple), Pink→(Purple,Blue), Purple→(Blue,Orange),
  Blue→(Orange,White), Orange→(White,Yellow), White→(Yellow,Red),
  Yellow→(Red,Green), Red→(Green,Pink).
- Command buttons use the accent color for text/highlighting and the alt color
  for metadata (shortcut column).

### File Finder

- File Finder metadata, rule separators, and the close button now use the
  active **Alt** color instead of the default red/muted tones.

### Cline provider integration

- Added Cline's hosted API gateway (`api.cline.bot`) as a first-class known
  provider so Cline-hosted models appear in the model picker and onboarding.
- Auth uses the `X-Api-Key` header; the key is read from `$CLINE_API_KEY` or
  the existing `providers.cline.api_key` config entry (in-app API key dialog
  works as with any other provider).
- Default model catalog lives in `internal/config/cline.go` and can be
  extended/overridden per-user via `providers.cline.models` in config.
- Added an "Other Models" entry to the `/` command palette (aliases: other,
  cline, external) that opens a dedicated dialog listing the Cline catalog;
  selecting a model prompts for the Cline API key if none is stored yet.
- The Other Models dialog now includes an "ADD CLINE API KEY" row when no
  Cline key is configured. Selecting it saves the key (verified via the
  `X-Api-Key` header) and immediately loads the full live Cline catalog,
  including Cline's free models, without a restart.
- Cline key verification (`TestConnection`) now authenticates via the
  `X-Api-Key` header so valid Cline keys are accepted.

### Menu reorganization

- Reorganized the main command-palette menu into the documented order: New
  Session, Sessions, Switch Model, Open File Finder, Change Project, Node
  Connections, Ollama Models, Other Models, Ollama How To, Themes,
  Notification Style, Toggle Beast Mode, Toggle Code Mode, Toggle Help,
  Initialize Project, Disable Background Color, Quit.

### Fixed (post-release)

- Project location renders cleanly on the homescreen (no garbled output) by styling
  the path a single time instead of re-rendering an already-styled string.
- The Command button is clickable (opens the command palette, same as `/` /
  `ctrl+p`) and the File Finder hit box is correctly positioned for the stacked layout.
- Bottom status-bar help text (descriptions and separators) is legible on the dark
  surface instead of near-black.
- The boot banner completes its full sequence and fades away; the reveal cap no longer
  traps the 34-rune attribution line.
- Long File Finder entries are clipped with an ellipsis instead of wrapping.
- Clicking outside the File Finder panel dismisses it (click-away close), in addition to
  the `[X]` close button.
- Stacked NODE, MCP, and LSP resource monitors vertically on the left side of the
  homescreen (NODE → MCP → LSP), with SKILLS on the right, for tighter space use.
- Clicking outside the Commands (command palette) dialog now dismisses it (click-away
  close), matching the File Finder.
- Restored the CLI suffix in the header wordmark (now renders "BVR-CLI"), per the
  wordmark source of truth documented in docs/UI_BRANDING.md.

## 1.1.5

### Mobile companion bridge

- Added mobile companion bridge planning and server design.
- Added `docs/MOBILE-CLI.md` for the iOS/Android companion roadmap.
- Added `docs/BVR-SERVER.md` for the host-side companion server.
- Added iOS companion app scaffold under `mobile-cli/ios/`.
- **Note:** the mobile companion is present for testing, but the mobile version
  is currently in **test mode** and not yet ready for general use.

### Validation

- Documented companion server architecture and protocol requirements.
- Added mobile companion app scaffolding with SwiftUI.

### Packaging & onboarding

- Added `docs/ONBOARDING.md` for first-run setup and dependency guidance.
- Added `docs/installation.md` for first-run setup and dependency guidance.
- Added `docs/IMG-RESOURCES.md` for screenshot, video, and icon asset guidance.
- Added `docs/ABOUT.md` for company and maintainer contact info.
- Added packaging wrappers under `packaging/` for Homebrew, NPM, macOS, and Windows.
- Added `.goreleaser.dist.windows.yaml` for Windows archive packaging.
- Updated `.goreleaser.yml` tap/scoop ownership and metadata for distribution.
- Added macOS DMG with compiled Swift launcher.
- Launcher auto-detects terminal preference: Ghostty first, then Alacritty, Kitty, WezTerm, iTerm2, Terminal.app.
- Added app icon assets: `.icns` for macOS, `.ico` for Windows, iOS icon exports.

## 1.1.3

### Native Ollama coding

- Added `bvr-cli code --native` for small local Ollama coding models.
- Added the `cmd/codetool` sidecar with a native Ollama `/api/chat` client and
  a minimal flat coding tool set.
- Added recovery for Qwen-style text tool calls, fenced JSON, truncated view
  calls, object-form arguments, and `file_path` aliases.
- Added bounded Ollama turns and explicit incomplete-stream errors so stalled
  local models do not leave the terminal loading indefinitely.
- Buffered likely tool-call JSON during streaming so protocol payloads do not
  clutter the user-facing response.
- Added installation, diagnostics, and local-model documentation for the native
  coder path.

### Bottom resource bar

- Added a live CPU and RAM resource monitor bar at the bottom of the TUI.
- The bar updates every second and shows CPU usage percentage and heap RAM
  usage percentage with token-driven theme colors.

### Validation

The native path was live-tested with `qwen2.5-coder:3b-instruct` against a local
Ollama daemon on macOS. Chat/file-read turns completed successfully in roughly
2–10 seconds during validation.
