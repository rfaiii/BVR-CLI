# Changelog

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
