# CLINE_FEATURE — Integrating Cline as a First-Class Coding Assistant

This document tracks the work to make Cline a permanent, model-independent
coding assistant inside BVR-CLI. It is the companion spec to
`internal/skills/builtin/cline/SKILL.md`.

## Goal

Let a user reach and run Cline from anywhere in the app — including before,
during, and after onboarding — without first selecting or signing up for a remote
model. When invoked, Cline runs as a full autonomous agent (Claude/OpenAI/Google/
OpenRouter/Bedrock via its own API keys and config), while BVR stays the
terminal/workspace host and the local-model planner for smaller jobs.

## Status by capability

| # | Capability | Status |
| - | - | - |
| 1 | Load local Ollama models at startup | Done |
| 2 | Skip model selection and enter the homescreen | Done |
| 3 | `Cline` / `Hermes` available in the command palette regardless of model | Done |
| 4 | Cline plan/act mode selector (Tab) persisted | Planned |
| 5 | Direct "run Cline on this task" entry (bypass the local model) | Planned |
| 6 | Cline tool permission passthrough / session hand-off | Planned |

## 1. Load local Ollama models at startup (done)

`internal/ui/dialog/models.go` `Models.LocalModelsCmd()` now starts an
installed-but-offline Ollama daemon before listing models, so local models
appear on first open (and on `r` refresh) without a manual `ollama serve`.

- Behavior: `runtime.Status()` -> if `StatusOnline`, list; otherwise if
  `status.Installed`, call `runtime.Start()` (idempotent; no-ops when healthy)
  then re-check, then list. Not installed -> returns the unavailable status.
- Onboarding help includes `r` (refresh) and `s` (start Ollama) so the user can
  surface models manually if needed.
- Limitation: when Ollama is not installed, no local group is shown; a follow-up
  should render the `RuntimeStatus` in-dialog (offline / not installed) with a
  one-click "install / start" call-to-action. Tracked separately.

## 2. Skip model selection and enter the homescreen (done)

- The startup model picker opens in onboarding mode (`openModelsDialog` ->
  `dialog.NewModels(com, isOnboarding=true)` at `internal/ui/model/ui.go`).
- Pressing `esc` (Close) during onboarding now emits `ActionSkipOnboarding`
  instead of being a no-op (`internal/ui/dialog/models.go`, `HandleMsg`).
- `handleDialogAction` (`internal/ui/model/ui.go`) handles
  `ActionSkipOnboarding` by closing the models dialog, transitioning to
  `uiLanding` (homescreen), calling `SetupAgents()` and `InitCoderAgent()` —
  mirroring the post-select onboarding path (`handleSelectModel`, ~line 2326).
- An on-screen hint ("or press esc to skip and explore later.") makes the action
  discoverable.

Design note: skipping does not require a configured provider. The coder agent
coordinator is initialized with the default config agent; if no model is
selected yet, agent runs that need a model will surface a clear "select a model"
error, while skill-driven tasks (Cline/Hermes) only need the skill attached.

## 3. Cline / Hermes in the command palette, regardless of model (done)

- Both skills are built in (`internal/skills/builtin/{cline,hermes}/SKILL.md`),
  `user-invocable: true`, and load at app start via `loadCustomCommands` ->
  `commands.FromSkillCatalog` -> `ActionAttachSkill` regardless of model state.
- `internal/ui/dialog/commands.go` now renders user-invocable skill commands
  title-cased (e.g. `Cline`, `Hermes`) with aliases `<name>` and `/<name>`, so
  they filter by `/cline` / `/hermes` in the palette — matching the existing
  `/ollama`, `/themes`, `/node` convention.
- The sidebar **Skills** status list also shows them under system skills.
- Invoke: open palette (`ctrl+p`), type `/cline` (or `cline`), select -> Cline's
  `SKILL.md` is attached as a message attachment; send a task -> the coding
  agent loads the skill and delegates to the Cline CLI.

Filtering is **palette-wide** (`internal/ui/dialog/commands.go`): typing a query
searches the System, User, and MCP command tabs at once, so `/cline` is found
from the default System view without first switching tabs. Skill commands also
render title-cased with `<name>`/`/<name>` aliases.

## Known small-model failure modes (escalate to Cline)

When the active coder model is a small local model (e.g. `qwen2.5-coder:3b`),
it can emit **malformed or hallucinated tool calls** instead of answering a
straightforward task. Example observed for
"scan README.md and tell me what this app is about":

- `{"name": "describe", "arguments": {}}` — `describe` is **not** a registered
  BVR tool (pure hallucination).
- `{ "function": { "sourcegraph_search_code": { "query": ..., "context_window": 10,
  "count": 20, "timeout": 120 } } }` — the real BVR tool is `sourcegraph`
  (`internal/agent/tools/sourcegraph.go`, `SourcegraphToolName = "sourcegraph"`).
  The model used the wrong function name and chose a web code-search tool for a
  local README read, even though it reproduced `sourcegraph`'s real param names
  (`context_window`/`count`/`timeout`).

Root cause: model-fit. A 3B model given the full BVR tool surface confuses tool
selection and naming. This is the exact case Proposal B (model-fit tagging +
warning) and Proposal C (coder prompt for small models / fallback retry) in
`BVR_CODER.md` target. The mitigation today is to hand the task to Cline via
`/cline`, which routes to a capable model. Add this query to the regression
checklist when testing small-model + Cline handoff.

## Notes / follow-ups

- **Reset onboarding in-app (future):** there is currently no way to re-trigger
  the startup model picker from inside the app — you have to quit and restart.
  A good follow-up is a command (e.g. `Reset Setup`) that re-opens the models
  dialog in onboarding mode, so users can re-pick a provider/model without
  relaunching. Tracked as a future task.
- **Palette-wide filter caveat:** while typing a non-empty query the radio
  header still highlights the currently selected tab while the list shows matches
  across all tabs. A future polish could dim the radio or group matches by tab.
- **Model dependency:** `/cline` attaches the skill and has the *local* BVR
  model drive the delegation. Using Cline without any configured model requires
  item 5 (direct "run Cline on this task") or a local Ollama model to act as the
  orchestrator brain.

## 4. Plan / Act mode selector (Tab) + persistence (planned)

Cline can operate in **plan** (propose, ask, then act) and **act** (execute)
modes. The intended UX:

- Add a `cline.mode` setting (`plan` | `act`, default `act`) persisted in
  `config.Options` (`internal/config/config.go`).
- Add a Tab-bound toggle in the message composer that is only active when the
  Cline skill attachment is present (a future "skill args" mechanism). Tab is
  otherwise globally bound to focus-change (`keys.go`), so the binding must be
  gated on `skill attached + coder/assistant context` to avoid conflicts.
- When Cline is invoked, pass the persisted mode into the skill invocation
  (e.g. append `--mode <plan|act>` to the `cline` command in `SKILL.md`, or set
  `.clinerules`/env). This requires parameterizing skill invocations — a
  follow-up to the current full-instructions `<loaded_skill>` flow.

Open design questions: (a) exact Cline CLI flag for mode (verify against the
installed `cline` version before hardcoding `cline.md`); (b) whether the mode is
per-task (composer toggle) or global (settings). Mark global default in settings,
per-task override via the Tab toggle.

## 5. Direct "run Cline on this task" entry (planned)

The current `/cline` path attaches the skill and lets the *local* model decide
to delegate. For a true model-independent shortcut (no local model required),
add an action that shells out `cline --yolo --cwd <project> "<pending input>"`
using the user's queued message. This needs:
- A pending-input capture (the message composer text).
- A non-agent "external assistant" entry point so the local model is not
  invoked at all.
- Routing Cline's output into the chat transcript and the file tracker.

## 6. Settings-menu conflict avoidance

Cline keeps its own settings (`~/.config/Cline/`, `.clinerules`). BVR should
NOT duplicate or mirror those settings in the in-app settings menu, to avoid
stale/conflicting configuration. Instead:

- Keep Cline configuration in the Cline-native locations only.
- The BVR `cline` skill surfaces the relevant `cline` commands/flags in
  `SKILL.md` (install, auth, `.clinerules`, MCP) so users operate Cline through
  Cline's own tooling.
- The settings menu should only expose BVR-side concerns: whether the Cline
  skill is enabled/disabled, and the optional `cline.mode` default
  (`agents.coder.cline.mode` or a dedicated `cline:` option namespace).

## Testing with local models

- Start Ollama: `ollama serve` (or rely on the auto-start added in item 1).
- Run `./bvr-cli` and on the model picker, confirm an `Ollama (Local)` group
  lists your local models; pick one (e.g. `qwen2.5-coder`) as the large/coder
  model. No remote signup required.
- Open the palette (`ctrl+p`), type `/cline`, and select it. Send a coding task.
  The agent should load the Cline skill and delegate to `cline`.
- To exercise the no-model path: on the startup picker, press `esc` to skip,
  then open the palette and select `/cline`. (Requires item 5, or a local
  Ollama model, for the agent to drive the delegation.)
- Regression: `go build ./...`, `go test ./internal/ui/... ./internal/skills/...
  ./internal/commands/... ./internal/localmodel/...`.

## Related

- `internal/skills/builtin/cline/SKILL.md` — Cline skill content.
- `internal/skills/builtin/hermes/SKILL.md` — Hermes skill content.
- `internal/skills/embed.go` — `BuiltinPrefix` / `DiscoverBuiltin`.
- `internal/ui/dialog/commands.go` — palette rendering of skill commands.
- `internal/ui/model/ui.go` — `loadCustomCommands`, `openModelsDialog`,
  `handleSelectModel`, `handleDialogAction` (onboarding/skip).
- `docs/OLLAMA_HOW_TO.md` — native coder installation and diagnostics.
