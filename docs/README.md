# BVR-CLI Documentation Index

The `docs/` tree is organized by topic. Each top-level document is the single
reference for its area; working notes live in `ignore/` (never published) and
superseded specs live in `archive/`.

## Topics

| Document | Covers |
| --- | --- |
| [SETUP.md](SETUP.md) | Installation (macOS, Linux, Arch/AUR, Windows), build from source, dependencies, troubleshooting |
| [PROVIDERS.md](PROVIDERS.md) | Ollama local models, native small-model coder, Cline hosted gateway |
| [UI.md](UI.md) | Branding/wordmark, Beaver mascot, File Finder, command palette, SFX, dialogs |
| [BVR_COLOR_SCHEME.md](BVR_COLOR_SCHEME.md) | Source of truth for theming: color roles, markdown tokens, adding a theme |
| [UI_LAYOUT.md](UI_LAYOUT.md) | Landing layout geometry, mascot direction behavior, footer and resource-monitor placement |
| [NODES.md](NODES.md) | Multi-device NODE layer: HTTP, WebSocket, SSH transports and GGWave audio pairing |
| [MOBILE.md](MOBILE.md) | iOS/Android companion plan and host-side companion bridge |
| [ARCHITECTURE.md](ARCHITECTURE.md) | System boundaries, rendering rules, roadmap phases, decision log |
| [FEATURES.md](FEATURES.md) | Canonical feature inventory, contact, and support info |
| [SKILLS.md](SKILLS.md) | Agent Skills discovery and builtin skills |
| [PACKAGING.md](PACKAGING.md) | Release automation, installers, and image/icon assets |
| [BETA.md](BETA.md) | Beta program, testing checklist, release process |
| `TASKLIST.md` | Active performance and optimization work plan — local only (gitignored, not published) |

## Subdirectories

| Directory | Purpose |
| --- | --- |
| [config/](config/README.md) | `bvrrc` Bash configuration reference and config future work |
| [hooks/](hooks/README.md) | Hook protocol reference (PreToolUse), examples, future work |
| [media/](media) | asciinema `.tape` demo recordings |
| [archive/](archive) | Superseded/historical specs — not current contract |
| [ignore/](ignore/README.md) | Local working notes, intentionally gitignored |

## Naming conventions

- One topic, one file. New documentation merges into the existing topic file
  instead of adding a sibling.
- Root-level files are `UPPER_SNAKE_CASE.md` or `UPPER-Case.md` matching the
  existing set; avoid introducing new casing styles.
- Reference other docs with relative links (`[\`NODES.md\`](NODES.md)`) and
  section anchors (`UI.md#file-finder`) so moves stay traceable.
- Screenshots live in `../resources/screenshots/` and follow
  `UPPER_SNAKE_CASE.png`; see the asset section in
  [PACKAGING.md](PACKAGING.md).
