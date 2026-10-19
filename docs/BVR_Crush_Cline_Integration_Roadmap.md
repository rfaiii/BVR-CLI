# BVR-CLI × Crush × Cline: Integration and Roadmap

> **Goal:** evolve BVR-CLI into a personal, keyboard-first Go AI development environment that preserves BVR's identity and existing features, uses Crush's Go/Charm architecture where it helps, and selectively adopts Cline's agent workflows and model-discovery experience.
>
> **Recommended strategy:** do **not** attempt a wholesale source-code merge of Cline into BVR. Cline is a large TypeScript/JavaScript codebase, while BVR and Crush are Go applications. Treat Cline as a source of product behavior, protocols, UX ideas, and selectively portable algorithms; implement those ideas natively in BVR's existing Go architecture.

Repositories:
- BVR-CLI (destination): https://github.com/rfaiii/BVR-CLI
- Cline: https://github.com/cline/cline
- Crush: https://github.com/charmbracelet/crush

This plan is based on the public repository pages and documentation available when written (2026-10-18). Re-check the current default branches, dependency versions, licenses, and provider API docs before implementing. Repository internals and APIs evolve.

---

## 1. Target architecture: BVR remains the product

### Product principles

1. **Preserve BVR features first.** Keep its branded Bubble Tea TUI, file finder/previews, attachments, project switching, persistent project tracking, Ollama support, permissions, skills, MCP/LSP integrations, NODE transports, and packaging.
2. **Borrow architecture, not branding or code indiscriminately.** Study Crush for Go-native patterns and Cline for agent behavior and model-provider UX.
3. **One agent runtime.** Avoid maintaining separate “BVR agent”, “Crush agent”, and “Cline agent” loops. Build one BVR-owned runtime with pluggable providers and tools.
4. **Persistence is a first-class subsystem.** Search history, recent files, pinned files, provider/model choices, sessions, and preferences should survive restarts and have explicit schema migrations.
5. **Provider-neutral core.** OpenRouter is a priority, but the core should not assume one vendor's model catalog or request format.
6. **Safe by default.** No automatic paid-model fallback, no silent destructive tool calls, and no secrets in logs or database records.

### Proposed component map

```text
cmd/ / main.go
  └── internal/app (startup, dependency wiring, lifecycle)
       ├── config       settings, migrations, environment overrides
       ├── persistence  SQLite repositories + schema migrations
       ├── providers    provider interface, OpenRouter, Ollama, others
       ├── catalog      model discovery, normalization, capability metadata
       ├── agent        planning/acting loop, context management, retries
       ├── tools        file/search/shell/edit/MCP/LSP tools
       ├── permissions  approval policy and audit trail
       ├── workspace    projects, file index, ignore rules, recent files
       ├── sessions     conversations, checkpoints, resume/compaction
       ├── events       typed events/pub-sub for UI and runtime
       └── TUI           Bubble Tea views, commands, keybindings
```

These are logical boundaries, not a mandate to create every package immediately. Fit them to the existing `internal/` tree after auditing it. Prefer extending existing packages over duplicating functionality.

---

## 2. Phase 0 — Establish a safe baseline

### 2.1 Make a clean integration branch

```bash
git clone https://github.com/rfaiii/BVR-CLI.git
cd BVR-CLI
git switch -c feat/cline-inspired-agent
git status --short
git log --oneline -10
```

Capture the current baseline before changes:

```bash
go version
go env
GOWORK=off go test ./...
GOWORK=off go vet ./...
GOWORK=off go build ./...
```

Run the project’s documented Taskfile checks as well. BVR’s README documents `go test ./...`, `go vet ./...`, and `go build ./...`; it also notes using `GOWORK=off` for specific build workflows. Record failures that already exist so they are not misattributed to integration work.

Create a baseline tag or backup branch if that fits your release workflow. Do not start by replacing `go.mod`, the app entry point, or the TUI.

### 2.2 Audit BVR before designing changes

Read these first:

- `README.md`, `AGENTS.md`, `go.mod`, `Taskfile.yaml`
- `docs/SYSTEM_ARCHITECTURE.md`
- `docs/FEATURES-OVERVIEW.md`
- `docs/FILE_FINDER.md`
- `docs/OLLAMA_HOW_TO.md`
- `docs/SKILLS.md`
- `sqlc.yaml`, SQL schema/migration directories, and database repository code
- `internal/`, `cmd/codetool/`, `mcp-server/`
- Current model/provider picker, session persistence, permissions, agent loop, and search/file-finder implementations

Build a feature inventory with columns: **feature**, **current owner/package**, **persistent today?**, **tests**, **desired change**, **migration risk**. Mark which data already lives in SQLite or config files before adding any new storage.

Useful reconnaissance:

```bash
find internal -maxdepth 3 -type f | sort
grep -RInE 'OpenRouter|openrouter|model|provider|session|sqlite|search|file.?finder|permission' internal docs
grep -RInE 'CREATE TABLE|ALTER TABLE|sqlc' .
```

Use `rg` instead of `grep` if installed. Avoid printing `.env`, tokens, or other secret files into logs or prompts.

### 2.3 Audit the source projects and licenses

Clone reference repositories separately; do not copy files into BVR yet:

```bash
git clone https://github.com/charmbracelet/crush.git ../crush-reference
git clone https://github.com/cline/cline.git ../cline-reference
```

Inspect the current licenses and relevant package boundaries:

- Crush: `AGENTS.md`, `README.md`, `go.mod`, `internal/app`, `internal/config`, `internal/agent`, `internal/cmd`, `internal/db`, `internal/pubsub`, `internal/agent/tools`
- Cline: `LICENSE`, `README.md`, `apps/cli`, `sdk/`, provider/model catalog code, tool execution, task/session persistence, context compaction, permissions, and free-model documentation

**License hygiene:** Crush and Cline have their own license terms; BVR has its own license and contribution terms. Before copying any implementation, verify the exact license in the checked-out revision, preserve required notices, track provenance, and check dependency licenses. Reimplementing a behavior from public documentation is often cleaner than importing a file. This is engineering guidance, not legal advice.

---

## 3. What to learn from each project

### BVR-CLI — keep as the foundation

BVR already describes itself as a Go/Bubble Tea workspace with project files, conversations, local models, tools, permissions, skills, LSP, automation, file finder, attachments, persistent project tracking, and Ollama workflows.

**Keep and strengthen:**
- Current app startup and TUI shell
- File finder and its previews, metadata, hidden-file handling, paging, clipboard, attachments, and project switching
- Existing database and project/session persistence
- Ollama discovery/pull/start/warm-up behavior
- Current permissions, MCP, LSP, skills, NODE integrations
- Current packaging, release, and cross-platform build paths

### Crush — preferred Go/Charm reference

Crush is a Go terminal coding agent. Its documented architecture is useful for patterns around:
- App-level dependency wiring and service boundaries
- Provider/model abstraction via `charm.land/fantasy`
- Session-oriented agents and coordination
- Tool descriptions and modular built-in tools
- SQLite persistence with migrations and `sqlc`
- Pub/sub events to decouple agent work from the TUI
- MCP, LSP, hooks, project context files, and permission-aware tools

**Integration rule:** reuse a library or a small, well-isolated Go pattern when it fits BVR's versions and abstractions. Do not import Crush's `internal/...` packages as dependencies: Go `internal` packages are intentionally not public APIs, and importing them directly is generally not possible across module boundaries. Copying selected code also creates an ongoing maintenance and licensing obligation. Prefer BVR-owned interfaces with a thin adapter to public libraries.

### Cline — product and workflow reference

Cline's repository includes a TypeScript/JavaScript IDE extension and CLI/SDK work. Useful behaviors to study:
- Model catalog discovery and normalized model metadata
- Provider-specific request/stream adapters
- Tool-call parsing and resilient agent execution
- Plan/act-style workflows, task/session lifecycle, checkpoints, and context compaction
- Clear approval prompts and tool-result handling
- Distinguishing free models, subscription models, and usage-billed models in the picker
- Graceful handling of provider limits, missing models, and catalog failures

**Do not port the VS Code extension wholesale.** Editor-specific APIs and the TypeScript runtime do not map directly to a Go TUI. For BVR, implement the equivalent UX using Bubble Tea and the existing Go tool/runtime interfaces.

---

## 4. Priority feature plan

### P0 — Persistent search and file finder

Make this the first feature because it strengthens BVR's existing workflow without requiring a new agent architecture.

#### User-facing behavior

- Search files by path/name and, optionally, file contents.
- Search the current project by default; allow an explicit project/workspace scope.
- Persist recent queries, recent files, pinned files, and last-used search filters.
- Restore the finder state after restart (at least last query/filter and recents; don't necessarily reopen a modal automatically).
- Support hidden files and ignore rules with visible indicators.
- Preview text files safely; cap file size and avoid blocking the UI on large/binary files.
- Provide keyboard navigation, paging, copy path, open/attach, and project switching.
- Offer a “clear history” command and a configurable retention limit.

#### Implementation outline

1. Inventory the current finder and reuse its existing indexing/scanning code.
2. Define a small `SearchService` interface; keep filesystem traversal out of Bubble Tea update/render methods.
3. Run long searches asynchronously and emit typed result events to the TUI.
4. Respect `.gitignore` and BVR's own ignore rules; make hidden-file inclusion explicit.
5. Use a persistent store for *history and user state*. Do not store every file's contents in SQLite just to make search persistent.
6. If content search becomes slow, add an optional index only after measuring. Keep the index rebuildable from disk and version its schema.
7. Add cancellation, result limits, path normalization, symlink-loop protection, and tests for large trees.

#### Candidate tables (adapt to the existing schema)

```sql
CREATE TABLE IF NOT EXISTS search_history (
  id INTEGER PRIMARY KEY,
  project_id TEXT NOT NULL,
  query TEXT NOT NULL,
  scope TEXT NOT NULL DEFAULT 'path',
  filters_json TEXT NOT NULL DEFAULT '{}',
  created_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_search_history_project_time
  ON search_history(project_id, created_at DESC);

CREATE TABLE IF NOT EXISTS recent_files (
  project_id TEXT NOT NULL,
  path TEXT NOT NULL,
  last_opened_at INTEGER NOT NULL,
  open_count INTEGER NOT NULL DEFAULT 1,
  PRIMARY KEY (project_id, path)
);

CREATE TABLE IF NOT EXISTS pinned_files (
  project_id TEXT NOT NULL,
  path TEXT NOT NULL,
  pinned_at INTEGER NOT NULL,
  PRIMARY KEY (project_id, path)
);
```

This is a **starting point**, not a drop-in migration. Match BVR's ID types, timestamp convention, migration style, and `sqlc` workflow. Store paths relative to the project root when possible; normalize them and validate that actions remain inside the workspace unless the user explicitly authorizes outside-workspace access.

### P1 — OpenRouter provider and dynamic model catalog

Make OpenRouter a first-class provider, but retain BVR's current provider and Ollama implementations.

#### Desired UX

- `/models` or the existing model picker can refresh the OpenRouter catalog.
- Search/filter by model ID, provider, modality, context window, tool calling, and price where metadata is available.
- Show a clear **free / paid / unknown pricing** label, last refresh time, and source of metadata.
- Allow model details and a harmless connection test.
- Persist the selected provider/model and picker preferences.
- If a saved model disappears, warn and ask the user to choose; don't silently switch to a paid model.
- Cache catalog responses with a TTL and support offline use from the last known catalog.
- Add a strict **free-only mode** that excludes models with nonzero or unknown pricing from automatic selection.

#### OpenRouter endpoints to verify against current docs

- Model catalog: `GET https://openrouter.ai/api/v1/models`
- Chat completions: `POST https://openrouter.ai/api/v1/chat/completions`
- Authentication: `Authorization: Bearer $OPENROUTER_API_KEY`

Verify current endpoint schemas, rate limits, response fields, tool-call support, and pricing metadata against the official OpenRouter docs before coding. Do not assume every catalog model supports streaming, tools, vision, JSON mode, or the same context window.

#### Suggested implementation sequence

1. Add an OpenRouter provider adapter behind BVR's provider interface.
2. Define a normalized internal `ModelInfo` type, e.g. provider/model ID, display name, context window, input/output price, modalities, tool-call capability, streaming support, free status, metadata timestamp.
3. Fetch the live catalog with timeouts, bounded response sizes, cancellation, and useful error messages.
4. Parse unknown fields defensively; providers can change their catalogs.
5. Cache the last successful response in SQLite or a small cache file, with schema/version and fetched-at time.
6. Add filters and sorting to the model picker without blocking rendering.
7. Persist explicit user choice. If the chosen model becomes unavailable, surface a warning rather than silently choosing a different one.
8. Add a provider smoke test and mock-server unit tests.

### P2 — “Find a free model and try it” workflow

This is the Cline-inspired feature you specifically want. Build it around OpenRouter first, with optional future adapters for other catalogs.

**Important distinction:** “free” is not a permanent property of a model across all providers and time. Catalog labels can change, quotas/rate limits apply, and some providers may have temporary promotions. Treat provider metadata as time-stamped evidence, not a guarantee.

#### Suggested commands / picker actions

```text
bvr models
bvr models --provider openrouter
bvr models --free
bvr models --refresh
bvr models --details <provider/model>
bvr models --test <provider/model>
```

Adapt names to BVR's existing command conventions; don't add duplicate command trees if `/models` already covers the workflow.

#### Free-model selection policy

- Identify explicitly free models using current catalog metadata and/or the provider's documented free-model identifier.
- For OpenRouter, investigate its `openrouter/free` router as a convenient option that chooses from available free models. Confirm current behavior and capabilities in the official docs.
- Offer both **specific free models** and **automatic free router** as separate choices.
- Before a request, show the provider/model that will actually receive it.
- Never automatically fall back from free to paid.
- If pricing is unknown, exclude the model from free-only selection.
- Display rate-limit/context/tool-call caveats and report provider errors without retrying indefinitely.
- Do not claim “guaranteed free”; show when catalog data was last refreshed.

#### Example UI details

```text
OpenRouter models  •  catalog refreshed 2 minutes ago

> deepseek/...:free      FREE*   tools: yes   context: ...
  qwen/...:free          FREE*   tools: unknown
  openrouter/free        FREE ROUTER   chooses a compatible free model

*Catalog indicates zero price; provider limits and availability still apply.
```

The model IDs above are placeholders for the UI example, not recommendations or a claim of current availability.

### P3 — Agent loop and tool-call reliability

Use Crush as the Go implementation reference and Cline as a behavior reference.

- One BVR-owned agent loop with provider-neutral messages and tool-call types.
- Tool registry with stable names, JSON schemas/argument validation, descriptions, and structured results.
- Explicit turn limits, cancellation, request timeouts, retry/backoff rules, and error classification.
- Streaming events for text, tool calls, tool results, usage, errors, and completion.
- Persist user/assistant/tool messages with ordered IDs and session association.
- Make interrupted runs resumable where safe.
- Validate tool arguments before execution; never treat model output as trusted shell input.
- Separate read-only tools from mutating tools and require approval based on policy.
- Test providers that return malformed tool arguments, partial streams, multiple tool calls, or unsupported tool calls.

Potential initial tools: `read_file`, `search_files`, `list_files`, `apply_patch`, `run_command`, `git_status`, and existing BVR/MCP/LSP tools. Reuse existing implementations where possible.

### P4 — Plan/act mode, checkpoints, and context management

Adopt the workflow concepts rather than Cline's editor implementation.

- **Plan mode:** agent can inspect and propose a plan but cannot mutate files or execute arbitrary commands.
- **Act mode:** agent can use tools subject to the normal permission policy.
- Persist mode per session and optionally a global default.
- Show a compact plan checklist in the TUI.
- Add checkpoints before multi-file edits; let the user review diffs and restore a checkpoint.
- Track context budget using model metadata, but treat estimates as estimates.
- Compact old conversation history with a summary while preserving key decisions, file changes, and unresolved tasks.
- Never compact away pending tool results or the current user's instruction.
- Make compaction transparent and recoverable; retain the original transcript in persistence.

Start with a simple, tested summarization policy before implementing complex autonomous compaction.

### P5 — Provider flexibility and local models

- Keep Ollama's current discovery, pull, startup, and warm-up behavior intact.
- Standardize provider capabilities and model metadata behind a common interface.
- Support OpenAI-compatible endpoints with configurable base URL, auth, and optional headers.
- Handle provider differences in tool calls, streaming, images, token usage, and context limits explicitly.
- Store secrets in the OS keychain where practical or environment variables; avoid storing raw API keys in SQLite or project config.
- Support a connection test that makes a minimal request only after the user initiates it.

### P6 — MCP, skills, hooks, and integrations

BVR already has these capabilities, so focus on compatibility and consistency:
- Normalize tool names and error reporting across built-in and MCP tools.
- Keep tool permissions visible and per-tool configurable.
- Add hooks only if there is a concrete use case; define timeouts, environment, output schema, and failure behavior.
- Keep skills discoverable and document which files are trusted.
- Preserve NODE transports and avoid coupling them to the agent loop.

---

## 5. Persistence design

### Choose a single source of truth

Before creating tables, determine what BVR already stores in SQLite, config files, or local files. Avoid duplicating session/provider state in two stores.

A reasonable division:

- **SQLite:** sessions, messages, search history, recent/pinned files, model catalog cache, usage summaries, checkpoints metadata, schema version.
- **Config:** non-secret user preferences, provider definitions, default model, free-only preference, UI settings.
- **Secret store/environment:** API keys and tokens.
- **Filesystem:** user project files, optional rebuildable search index, patch/checkpoint artifacts if needed.

### Migration rules

1. Use numbered/versioned migrations consistent with the current project.
2. Make migrations transactional where the database library supports it.
3. Keep old data readable; do not reset the database during upgrades.
4. Add indexes for real query patterns, not every column.
5. Include retention/cleanup policies for search history and catalog cache.
6. Test migration from a copy of a realistic older database.
7. Use `sqlc` generation if that is BVR's current database workflow.
8. Ensure concurrent app instances cannot corrupt preference/history writes.

### Persistence acceptance tests

- Restart BVR and verify recents, pins, provider/model selection, and preferences persist.
- Migrate a prior-version database with existing sessions intact.
- Verify clear-history removes only the requested history, not sessions or pinned files.
- Verify a corrupt or stale catalog cache fails gracefully and can be refreshed.
- Verify no API key or auth header is stored in query history, logs, or model metadata.

---

## 6. Provider interface sketch

Use this as a design discussion aid; adapt types to BVR's current agent/provider implementation instead of introducing a parallel abstraction unnecessarily.

```go
type ModelInfo struct {
    ID             string
    Name           string
    Provider       string
    ContextWindow  int
    InputPrice     *float64
    OutputPrice    *float64
    Free           bool
    PricingKnown   bool
    SupportsTools  bool
    SupportsImages bool
    SupportsStream bool
    UpdatedAt      time.Time
}

type Catalog interface {
    ListModels(ctx context.Context, refresh bool) ([]ModelInfo, error)
    GetModel(ctx context.Context, id string) (ModelInfo, error)
}

type Provider interface {
    ID() string
    ListModels(ctx context.Context, refresh bool) ([]ModelInfo, error)
    Stream(ctx context.Context, req Request) (<-chan Event, error)
}
```

Consider separate interfaces for chat completion, embeddings, and catalog access if providers vary significantly. Avoid returning a channel that can leak goroutines: define cancellation, channel closure, and error delivery clearly. Do not assume every provider implements every capability. A `Capabilities` struct or optional interfaces may be more accurate than booleans that imply support without testing.

---

## 7. Suggested implementation order and milestones

### Milestone A — Baseline and architecture map
- [x] Run tests/build/vet and record baseline. → `docs/PHASE0_AUDIT.md` §1. Build/vet
      pass; `go test ./...` has 2 pre-existing `internal/catalog` FTS failures.
- [x] Document current package ownership and data stores. → `docs/PHASE0_AUDIT.md` §2–3.
- [x] Identify existing persistence schema and migration mechanism. → goose + sqlc,
      one `bvr.db` per project data dir. Audit §2.
- [ ] Verify licenses and dependencies of all three projects. (Needs maintainer
      confirmation — see ADR 0001 open question 4.)
- [x] Write a one-page architecture decision record (ADR) explaining what is
      borrowed vs. reimplemented. → `docs/adr/0001-borrowed-vs-reimplemented.md`.

**Done when:** the existing app still builds, and the integration scope is a set of small, reviewable changes.

> **Scope corrections from the audit — read before implementing P0–P4:**
>
> 1. **Drop `project_id` from the candidate tables in §4/P0.** Persistence is
>    already per-project (one `bvr.db` per data dir). Project scoping is implicit.
>    Key new tables on project-relative paths instead.
> 2. **The `catalog` component name is taken.** `internal/catalog` is an unrelated
>    product/image-lookup database. Use a different name (proposed
>    `internal/modelcatalog`).
> 3. **P0 is a build, not an enhancement.** The existing "file finder"
>    (`internal/ui/dialog/filebrowser.go`) is a single-directory `os.ReadDir`
>    lister with no query, no fuzzy match, no content search, no gitignore
>    awareness, and no persistence. Re-read §4/P0 with that scope.
> 4. **P1's provider work is already done.** OpenRouter inference works via
>    `charm.land/fantasy/providers/openrouter`, including per-step cost tracking,
>    and catwalk already ships an OpenRouter catalog with pricing. The real gaps
>    are: live catalog fetch, TTL cache/offline fallback, free/paid/unknown
>    labeling, free-only mode, and a connection test. See audit §5.
> 5. **Cline the gateway is already integrated** (`internal/config/cline.go`).
>    What is missing is only Cline's product behavior, which is what P1/P2/P4 target.
> 6. **P3's "consolidate the agent loop" is lower value than it looks** — a single
>    runtime already exists. Defer unless P4 requires it. Likewise, check whether
>    the versioned `files` table is already a checkpoint mechanism before building one.

### Milestone B — Persistent finder/search
- [ ] Add migrations for search history/recent files/pins only where missing.
- [ ] Add a repository/service layer and unit tests.
- [ ] Run search asynchronously with cancellation.
- [ ] Wire finder interactions to persistence.
- [ ] Add clear-history and retention behavior.
- [ ] Verify file preview and attachment flows still work.

**Done when:** search state survives restarts and large searches don't freeze the TUI.

### Milestone C — OpenRouter catalog and free-only discovery
- [ ] Add OpenRouter provider adapter.
- [ ] Fetch and normalize current model catalog.
- [ ] Add timeouts, cache, refresh, and offline fallback.
- [ ] Display price/free status with last-refresh timestamp.
- [ ] Add free-only filtering and explicit no-paid-fallback policy.
- [ ] Add model detail and connection-test actions.
- [ ] Add tests using fixture responses and an HTTP test server.

**Done when:** a user can discover, inspect, select, and test a free model without accidentally switching to a paid model.

### Milestone D — Reliable agent loop
- [ ] Consolidate existing agent execution behind one BVR runtime.
- [ ] Normalize provider events and tool calls.
- [ ] Add argument validation, permission checks, cancellation, and bounded retries.
- [ ] Persist all session/tool events.
- [ ] Add provider contract tests.

**Done when:** provider differences do not leak into the TUI and failed/partial calls produce understandable errors.

### Milestone E — Plan/act, checkpoints, compaction
- [ ] Add session-level plan/act state.
- [ ] Add diff review and checkpoints before mutating multi-file tasks.
- [ ] Add context-budget estimates and transparent compaction.
- [ ] Test interruption/resume and rollback paths.

**Done when:** the agent can perform longer tasks with visible user control and recoverable changes.

### Milestone F — Polish and release
- [ ] Run full tests, vet, build, formatting, and cross-platform checks.
- [ ] Test fresh install and upgrade from a prior BVR version.
- [ ] Verify config migration and secret handling.
- [ ] Update README/docs and add a migration note.
- [ ] Add logs that are useful but redact secrets and sensitive prompts by default.
- [ ] Validate packaging/release artifacts.

---

## 8. Test strategy

### Unit tests
- Catalog JSON parsing, missing fields, price parsing, unknown capabilities.
- Free-only filtering when price is zero, nonzero, missing, or stale.
- Search-history persistence, project scoping, path normalization, retention.
- Permission decisions and plan-mode restrictions.
- Agent event ordering, cancellation, retry limits, and tool argument validation.
- Database migrations and repository methods.

### Integration tests
- Mock OpenRouter endpoints with `httptest`; never depend on live catalog data in ordinary CI.
- Simulate timeout, 401, 429, 5xx, malformed JSON, missing model, and provider-side context-limit errors.
- Test Ollama behavior without requiring a live daemon; keep separate optional smoke tests for a real local server.
- Test app startup with empty config, old config, missing credentials, and offline catalog.
- Exercise keyboard flows in the TUI and verify state is restored after restart.

### Manual acceptance checklist
- [ ] Existing BVR features still work.
- [ ] Search/finder remains responsive on a large repository.
- [ ] Recents and pins survive restart.
- [ ] OpenRouter catalog refresh and cached/offline view both work.
- [ ] Free-only mode excludes paid and unknown-price models.
- [ ] A missing model never triggers silent paid fallback.
- [ ] No credentials appear in logs, database dumps, or diagnostics.
- [ ] Model errors show a useful message and do not create infinite retry loops.
- [ ] Build and packaging work on the supported OS/architecture matrix.

---

## 9. Security and reliability guardrails

- Never silently change from a free model to a paid model.
- Treat model names and catalog metadata as untrusted remote input.
- Apply HTTP timeouts, response-size limits, and context cancellation.
- Redact API keys, authorization headers, and secrets from errors and logs.
- Keep shell execution and file writes behind explicit permission checks.
- Use path containment checks to prevent traversal outside the selected workspace.
- Respect ignore files but show when a file is excluded.
- Do not run hooks or project config from untrusted repositories without clear trust boundaries.
- Validate MCP tool results and all model-supplied arguments.
- Keep an audit trail for mutating actions, but make sensitive-content logging opt-in.
- Make database migrations and checkpoint restore operations safe against partial failure.

---

## 10. Common traps to avoid

1. **Copying Cline's TypeScript runtime into the Go application.** It won't become native Go; it adds a second runtime, dependency surface, build pipeline, and integration boundary. Only consider a separate process if you deliberately want a long-term Go/Node architecture.
2. **Importing Crush's `internal` packages.** They are not public reusable APIs. Port the pattern or depend on public libraries.
3. **Rebuilding the finder before auditing it.** BVR already documents a capable file finder; persist and refine it instead of replacing it.
4. **Duplicating provider abstractions.** Adapt the current BVR/Crush-compatible provider layer rather than maintaining two model registries and two agent loops.
5. **Assuming `:free` means permanently free.** Re-check live catalog metadata, enforce free-only selection, and make pricing uncertainty visible.
6. **Using live APIs in unit tests.** Use fixtures and mock servers; make live smoke tests opt-in.
7. **Adding an index before measuring.** Persist recents and queries first; index file contents only if performance measurements justify the complexity.
8. **Migrating all storage at once.** Use small schema migrations and verify upgrade paths.
9. **Allowing automatic paid fallback.** Missing model, exhausted quota, or API error should stop and ask the user.
10. **Making the TUI perform I/O synchronously.** Use Bubble Tea commands/background work and typed events.

---

## 11. Research links

### BVR
- Repository: https://github.com/rfaiii/BVR-CLI
- System architecture: https://github.com/rfaiii/BVR-CLI/blob/main/docs/SYSTEM_ARCHITECTURE.md
- File finder: https://github.com/rfaiii/BVR-CLI/blob/main/docs/FILE_FINDER.md
- Ollama guide: https://github.com/rfaiii/BVR-CLI/blob/main/docs/OLLAMA_HOW_TO.md
- Skills: https://github.com/rfaiii/BVR-CLI/blob/main/docs/SKILLS.md

### Crush
- Repository: https://github.com/charmbracelet/crush
- Development/architecture guide: https://github.com/charmbracelet/crush/blob/main/AGENTS.md
- README/provider configuration: https://github.com/charmbracelet/crush/blob/main/README.md

### Cline
- Repository: https://github.com/cline/cline
- Free models: https://github.com/cline/cline/blob/main/docs/getting-started/free-models.mdx
- OpenRouter provider guide: https://github.com/cline/cline/blob/main/docs/provider-config/openrouter.mdx
- Models API: https://github.com/cline/cline/blob/main/docs/api/models.mdx
- CLI changelog: https://github.com/cline/cline/blob/main/apps/cli/CHANGELOG.md

### OpenRouter
- Models API documentation: https://openrouter.ai/docs/api-reference/list-available-models
- Free Models Router: https://openrouter.ai/docs/guides/routing/routers/free-router

---

## 12. The first five concrete tasks

If you want the shortest path to useful progress, do these in order:

1. **Map the existing BVR database and provider/model picker.** Identify what already persists and what is missing.
2. **Add persistent search history, recent files, and pins** to the existing finder, with migrations and tests.
3. **Implement an OpenRouter catalog adapter** with cache, refresh, and model capability metadata.
4. **Add a free-only model picker** with explicit pricing status and no paid fallback.
5. **Only then** refactor the agent loop toward the strongest applicable Crush patterns and add Cline-inspired plan/act, checkpoints, and context compaction.

The outcome should be **BVR first**: a cohesive Go/Charm application with Cline-inspired workflows, rather than three projects awkwardly glued together.
