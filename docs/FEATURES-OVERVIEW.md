# BVR Features Overview

This document is the canonical feature inventory for BVR-CLI. It is intended
for beta marketing, release notes, onboarding copy, and internal triage.

## Core experience

- Terminal-native interactive assistant powered by Bubble Tea.
- Keyboard-first command palette and editor workflow.
- Session persistence with SQLite-backed history.
- Streaming assistant responses with markdown rendering.
- Onboarding wizard with terminal-aware ASCII previews.
- Explicit opt-out onboarding; never forces setup without user choice.

## AI and providers

- Multi-provider LLM support through `fantasy`:
  - Anthropic, OpenAI, Gemini, Bedrock, Copilot, Hyper, MiniMax, Vercel,
    OpenRouter, local models, and more.
- Provider-specific authentication flows:
  - API keys, OAuth, and Copilot device-code login.
- Model switching and discovery inside the TUI.
- Reasoning-effort and coder-agent handoff for small/local models.
- Hypercredit balance display in the model info area.

## Tools

- `bash` — run shell commands with streaming output.
- `view`, `write`, `edit`, `multiedit` — read and modify files.
- `glob`, `grep`, `rg`, `ls`, `search` — fast codebase discovery.
- `fetch`, `web_fetch`, `web_search` — web research from the assistant.
- `download` — pull remote assets into the workspace.
- `lsp_definition`, `lsp_symbols`, `lsp_references`, `lsp_rename`,
  `lsp_replace_symbol`, `lsp_call_hierarchy`, `lsp_restart` — code intelligence.
- `todos` — create and manage todo items in the session.
- `question` — interactive multi-step user prompts.
- `node` — device/routing integration for laptop↔iPhone sync.
- `mcp-tools`, `read_mcp_resource`, `list_mcp_resources` — MCP server access.
- `bvr_info`, `bvr_logs` — self-inspection and diagnostics.
- `job_output`, `job_kill` — background shell job control.

## Skills

- Agent Skills open-standard discovery (`SKILL.md` with YAML frontmatter).
- Builtin skills embedded into the binary.
- Auto-discovery from default directories:
  - `~/.config/bvr/skills`
  - `~/.config/agents/skills`
  - `~/.agents/skills`
  - `~/.claude/skills`
  - `~/Documents/AI-SKILLS`
- Project skill directories:
  - `.agents/skills`
  - `.bvr/skills`
- User skills override builtins with the same name.
- Skills are surfaced to the coding agent as `<available_skills>`.
- Skills can be user-invocable from the command palette.

## Sessions and history

- SQLite-backed session CRUD.
- Resume previous sessions by ID.
- Continue the most recent session on startup.
- Session naming and metadata.
- Prompt history with recall.

## LSP and code intelligence

- LSP client manager with auto-discovery.
- On-demand LSP server startup.
- Definition, references, symbols, rename, symbol replacement, call hierarchy.
- LSP restart tooling.
- Workspace-aware context loading.

## MCP integration

- MCP client integration.
- MCP tools, prompts, and resources exposed to the assistant.
- MCP lifecycle management and process handling.
- MCP auth dialogs.
- Channel support for webhook-style MCP servers.

## Node / device sync

- NODE connection testing and device routing.
- Laptop↔iPhone sync breadcrumbs in docs.
- Local transport, HTTP, and WebSocket protocol docs.

## Mobile companion

- iOS companion app scaffold under `mobile-cli/ios`.
- Test-mode mobile companion.
- Web preview assets for companion flows.

## Packaging and distribution

- macOS Apple Silicon and Intel builds.
- Windows x64/arm64 EXE builds.
- Linux amd64/arm64 builds.
- macOS DMG packaging script.
- Windows MSI packaging script.
- Homebrew formula.
- NPM package.
- Winget manifest support.
- `dist/` build outputs are gitignored and ready for release upload.

## Themes and styling

- Theme system with token-driven base builder.
- Multiple concrete themes with per-theme overrides.
- Gradient resource bars for CPU and RAM usage.
- Brightened secondary/help text for terminal readability.
- In-TUI cursor animation.
- Status, pills, dialog, markdown, diff, shell output, and thinking styles.

## Editor and attachments

- File attachment support (`[+]` button mapped to FILEFINDER for attach-only flow)
- Inline editors inside dialogs
- Bang-mode shell execution from the editor
- Shift+Enter for multiline input
- **CREATE FILE** button on the landing screen and `ctrl+n` in the File Finder to create new files
- **Web Browser** (`ctrl+b` or the browser icon button) to fetch and read articles from URLs

## Web browsing

- **Web Browser** dialog (`ctrl+b` or landing screen browser icon) to fetch and read articles from any URL
- Fetches HTML, extracts readable content via go-readability, renders as Markdown in a terminal viewport
- URL input field with `[X]` close button and click-outside-to-close
- Browser history and navigation (`alt+left` for back, `alt+right` for forward, `ctrl+h` for home)
- Seamless viewport scrolling (`up`/`down` or `pgup`/`pgdn`) separated from URL input field focus
- Loading spinner during fetch, error display on failure
- Help text at the bottom showing available keys (`Enter` to fetch, `Esc` to close)

## File creation

- **CREATE FILE** button on the landing screen (paper icon) and `ctrl+n` in the File Finder to create new files
- Opens the file picker dialog (`New File`) for browsing and creating files
- Click **[X]** in the upper-right corner, press `esc`, or click outside the panel to close
- Click-away-to-dismiss works on all dialogs (File Finder, Commands, Browser, and New File)

## Hooks and permissions

- User-defined shell hooks on tool events.
- Permission checking and allow-lists.
- Beastmode flag to auto-accept permissions.

## Config and onboarding

- JSON config via `bvr.json`.
- Bash-style `bvrrc` with builtins:
  - `provider`, `model`, `mcp`, `lsp`, `permissions`, `hook`, `options`.
- Context files: `AGENTS.md`, `BVR.md`, `CLAUDE.md`, `GEMINI.md`.
- Default context paths and project-level overrides.
- Data directory override with `--data-dir`.
- Workspace-aware path resolution.

## Developer and power-user features

- Client/server mode.
- Debug logging with `--debug`.
- Progress bar support for Ghostty, iTerm2, and Rio.
- Telemetry opt-out via PostHog config.
- Shader installation helper for Ghostty cursor shaders.
- Embedded GLSL cursor shaders with in-TUI animation.
- Cursor warp/sweep/tail/ripple shader support.
- Stats dashboard with HTML/CSS/JS export.
- Schema command for config inspection.
- Provider auto-update.

## Contact and beta support

- Maintainer: Richard Aizen Avery III
- Email: averydevz@outlook.com
- GitHub: https://github.com/richavery/bvr-cli-main
- Website: https://bvr-cli.com

## Beta feedback and bug collection

- Beta contact info is listed above.
- Known testing priorities:
  - NODE connection testing and iPhone↔laptop sync flow.
  - Windows MSI and winget manifest validation.
  - Public Homebrew tap and NPM publication.
  - TUI readiness in CI/headless environments.
  - Theme polish across terminal profiles.
  - Mobile companion runtime validation.

## Feature triage template

Use this section to review the list before marketing or beta release.

### Highlight / sell
-

### Needs work
-

### User help / onboarding gaps
-