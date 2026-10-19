# Phase 0 Audit: BVR-CLI × Crush × Cline

Companion to [BVR_Crush_Cline_Integration_Roadmap.md](./BVR_Crush_Cline_Integration_Roadmap.md).
This document records what actually exists in the tree today, so the roadmap's
recommendations can be adapted rather than applied literally.

Audit date: 2026-10-18. Branch: `main` @ `969b4b8`.

---

## 1. Baseline

Recorded before any integration changes.

| Command | Result |
| --- | --- |
| `go version` | `go1.27.2 darwin/arm64` |
| `go env GOOS GOARCH` | `darwin` / `arm64` |
| `GOWORK=off go build ./...` | **pass** |
| `GOWORK=off go vet ./...` | **pass** (no output) |
| `GOWORK=off go test ./...` | **pass** (after the `internal/catalog` fix below) |

### Pre-existing failures — FIXED

`internal/catalog` — the product/image-lookup catalog, unrelated to the agent,
was already red on `main`:

```
--- FAIL: TestSearch_FindsProductsByFTS      repository_test.go:66  "[]" should have 2 items, but has 0
--- FAIL: TestSearch_WithEmbeddingBoost     repository_test.go:107 "[]" should have 1 item,  but has 0
```

Reproducible in isolation. **Verified not an environment problem:** the active
driver on darwin/arm64 is `modernc.org/sqlite`
(`internal/db/connect_modernc.go`), and `sqlite_compileoption_used('ENABLE_FTS5')`
returns `1`. Products inserted without error, so the `products_fts` row insert
"worked" — it did not.

**Root cause: two independent bugs, both masked by a discarded error.**

In `InsertProduct` the FTS upsert was written as `_, _ = r.db.ExecContext(...)`.
The error was thrown away, so `InsertProduct` returned `nil` while the index
stayed permanently empty and every search returned zero rows.

1. **Placeholder mismatch.** The statement listed 10 columns but only 9 `?`:
   `SQL logic error: 9 values for 10 columns`.
2. **UPSERT is unimplemented on FTS5.** Even with the count corrected,
   `ON CONFLICT(product_id) DO UPDATE` fails with
   `UPSERT not implemented for virtual table "products_fts"`. `product_id` is
   `UNINDEXED`, so there is no unique constraint to conflict on either — the row
   must be explicitly deleted before reinsertion.

A third bug surfaced once the first two were fixed: `ftsQuery` built a
space-separated term list, which FTS5 interprets as **implicit AND**. Since that
query only retrieves candidates — relevance is decided afterwards by
`ComputeScore` over the parsed signals — AND collapsed recall to near zero, so a
normal multi-signal query matched almost nothing.

**Fix:** explicit `DELETE` + `INSERT` on the FTS row, wrapped in a transaction
together with the alias inserts so the index is never left empty; all previously
discarded errors are now returned. `ftsQuery` now joins terms with `OR` and
quotes each term, which both restores recall and stops untrusted input from
injecting FTS5 operators into the `MATCH` expression. Covered by new regression
tests for the FTS write path, replace-not-duplicate semantics, OR semantics,
punctuation handling, and quote escaping.

This is a genuine lesson for the integration work: **BVR already contains at
least one instance of a swallowed error hiding a total feature failure.** Audit
new code for `_, _ =` and empty error branches.

### Environment notes

- Module path is `github.com/richavery/bvr-cli`; `go.mod` declares `go 1.26.5`.
- `go.work` exists but contains only `./`, so `GOWORK=off` is equivalent today.
  Keep using it anyway — the roadmap mandates it and it future-proofs against
  sibling workspaces.
- SQLite driver is build-tag switched: `modernc.org/sqlite` on common platforms,
  `github.com/ncruces/go-sqlite3` elsewhere. Any new FTS/index work must be
  validated on **both** drivers.

---

## 2. Storage map (this is the load-bearing finding)

### 2.1 One SQLite database per project

`internal/db/connect.go` opens `<dataDir>/bvr.db`. `dataDir` is **per project**
(`internal/projects/projects.go` maps project absolute path → data dir).
There is no global multi-tenant database.

Consequences for the roadmap:

- **The roadmap's `project_id TEXT` column is unnecessary and should be
  dropped.** Project scoping is implicit in which `bvr.db` you are talking to.
  This removes a column, a foreign key, and an entire class of scoping bugs from
  the P0 schema.
- Search history / recents / pins should key on **project-relative paths**, not
  absolute paths, so they stay portable across checkouts.
- `internal/cmd/stats` already relies on this by opening many project DBs
  read-only via `ConnectReadOnly` — a good precedent for any cross-project view.

### 2.2 Current schema (goose + sqlc)

Migrations in `internal/db/migrations/`, embedded via `//go:embed`, applied by
`goose.Up` on every `Connect`.

| Table | Purpose |
| --- | --- |
| `sessions` | id, parent_session_id, title, token/cost counters, `todos` JSON |
| `messages` | id, session_id, role, `parts` JSON, model, provider, summary flags |
| `files` | versioned file snapshots per session (path, content, version) |
| `read_files` | (path, session_id, read_at) — files the agent has read |

Conventions to match when adding tables:

- Timestamps are **Unix milliseconds** in `INTEGER` (except `read_files.read_at`,
  which is seconds — note the inconsistency and pick one deliberately).
- `updated_at` is maintained by `AFTER UPDATE` triggers, not by the app.
- Text `id` primary keys, not autoincrement integers.
- `FOREIGN KEY ... ON DELETE CASCADE` from children to `sessions`.

### 2.3 Other stores

| Store | Location | Owner |
| --- | --- | --- |
| Provider/model/user preferences | `bvr.json` / `bvrrc` (scoped) | `internal/config` |
| Recent models (max 5 per type) | same config JSON | `internal/config/config.go` |
| Tracked projects | global `projects.json` | `internal/projects` |
| Product/image catalog | **separate** SQLite, `./data/catalog.db` | `internal/catalog` |

> **`internal/catalog` is a name collision.** It is *not* the model catalog the
> roadmap proposes. It is a handbag/product lookup database for the image
> recognition feature (FTS5 + product embeddings). Any new model-discovery
> package must take a different name — `internal/modelcatalog` or
> `internal/providers` — or the roadmap's component map needs amending.

### 2.4 Concurrency

- `internal/db`: `SetMaxOpenConns(1)` deliberately, to avoid WAL/header desync
  under concurrent sub-agents. Keep this in mind for any new repository.
- `internal/projects`: in-process `sync.Mutex` only. **No cross-process lock.**
  The roadmap's "concurrent app instances must not corrupt preference/history
  writes" requirement is therefore unmet for `projects.json` today.
- `internal/config`: has both an in-process mutex and a cross-process file lock
  (`internal/lock`), with an explicit "no I/O while holding the lock" rule.

---

## 3. Feature inventory

Legend: **Persist** = survives restart. Store column names where.

| Feature | Owner | Persist | Notes |
| --- | --- | --- | --- |
| TUI shell, layout, themes | `internal/ui` | config | Bubble Tea v2 + Ultraviolet screen pipeline |
| Agent loop | `internal/agent/agent.go`, `coordinator.go` | SQLite | Already a single runtime with named agents |
| Sessions & messages | `internal/session`, `internal/message` | SQLite `sessions`/`messages` | |
| File versioning / checkpoints | `internal/db` `files` table | SQLite | Versioned snapshots already exist — reuse for P4 |
| Permissions | `internal/permission` | config | Allow-list driven |
| Hooks | `internal/hooks` | config | PreToolUse, runs before permission check |
| Skills | `internal/skills` | filesystem | |
| MCP | `internal/agent/tools/mcp` | config | |
| LSP | `internal/lsp` | config | Auto-discovery |
| Ollama / local models | `internal/localmodel` | config | Discovery/pull/start/warm-up |
| Projects | `internal/projects` | `projects.json` | By abs path, mutex only |
| Recent models | `internal/config` | config JSON | Max 5 |
| Model picker UI | `internal/ui/dialog/models*.go` | — | **No price/free display** |
| `bvr models` CLI | `internal/cmd/models.go` | — | Static list only, no flags for free/refresh |
| **File finder** | `internal/ui/dialog/filebrowser.go` | **nothing** | See below |
| OpenRouter inference | `internal/agent/*` via `fantasy/providers/openrouter` | — | **Already works**, incl. cost tracking |
| OpenRouter catalog | `charm.land/catwalk` **static JSON** | n/a | **No live fetch** |
| Cline gateway provider | `internal/config/cline.go` | config | Already a first-class provider |
| Pubsub | `internal/pubsub` | — | Decouples agent from UI |
| Telemetry | `internal/event` (PostHog) | — | |

---

## 4. The "file finder" is not a search tool

This is the most important correction to the roadmap. `docs/FILE_FINDER.md` and
the roadmap both describe BVR as having "a capable file finder" to persist and
refine (roadmap trap #3: *"Rebuilding the finder before auditing it"*).

What actually exists, `internal/ui/dialog/filebrowser.go` (581 lines,
`FileBrowserID = "file-browser"`, bound to `ctrl+o`):

- A **single-directory `os.ReadDir` lister**, not a recursive fuzzy finder.
- Per-directory entries, two-pane list + preview, clipboard pane.
- Preview truncates at 12 000 bytes and detects binary via NUL byte.
- Keys: `enter/l` descend, `left/h` go up, `.` toggle hidden, `r` refresh,
  `y` copy path, `a` attach, `o` open editor, `s` switch project,
  `b` web browser, `ctrl+n` new file.
- I/O correctly runs in `tea.Cmd`s (`requestReload`, `requestPreview`) with a
  `loadSeq` guard — **this pattern is already correct**, keep it.
- **No query input, no fuzzy matching, no content search, no gitignore
  awareness, no recents, no pins, no persistence, no retention.**

Therefore P0 as written is **not** "add persistence to an existing search". It
is "build a real search/finder, and persist it". That is a larger scope than
Milestone B implies, and it is still the correct first feature because it needs
no new agent architecture. Budget it accordingly.

---

## 5. OpenRouter status: the roadmap overstates the work

Roadmap P1 says *"Make OpenRouter a first-class provider"* and Milestone C
starts from *"Add OpenRouter provider adapter"*. Both are already done:

1. **Inference works.** `internal/agent/agent.go` and `coordinator.go` import
   `charm.land/fantasy/providers/openrouter`, build the client from
   `$OPENROUTER_API_KEY`, and already parse `openrouter.ProviderMetadata` to
   record **per-step cost** on sessions.
2. **The model catalog exists.** `charm.land/catwalk` ships OpenRouter as a
   known provider with `catwalk.Model` entries carrying `CostPer1MIn`,
   `CostPer1MOut`, `ContextWindow`, `CanReason`, and `SupportsImages`.
3. **`TypeOpenRouter` is already a distinct provider type** in catwalk.

So the *actual* gap is narrower and mostly about freshness and labeling:

| Roadmap P1 ask | Reality | Real work |
| --- | --- | --- |
| OpenRouter provider adapter | Done | — |
| Normalized `ModelInfo` | `catwalk.Model` is close; add `Free`/`PricingKnown`/`FetchedAt` on the BVR side | Small |
| Fetch live catalog | **Missing** — the catalog is a static JSON snapshot baked into the catwalk module | Fetch + TTL cache |
| Free/paid/unknown label | **Missing** in CLI and TUI | Derive from `CostPer1M*` |
| Cache with TTL + offline | **Missing** | New |
| Free-only mode, no paid fallback | **Missing** | New |
| Connection test | **Missing** | New |

Free detection is cheap: `catwalk.Model` has non-pointer `float64` costs, so a
live fetch can set an explicit `PricingKnown` flag while the static catalog
cannot distinguish "free" from "unknown" — another reason the live fetch matters.

Note the reference implementation already exists and is worth reading before
writing one: `catwalk/cmd/openrouter/main.go` shows the full model + per-model
`/endpoints` fetch, endpoint selection (prefers `tools` support, then context
length, then uptime), and pricing normalization to cost-per-1M. Reimplement the
shape; do not vendor the file.

---

## 6. Cline status

`internal/config/cline.go` integrates **Cline's hosted API gateway**
(`https://api.cline.bot/v1`, `X-Api-Key`) as a first-class provider with six
curated models. `docs/CLINE_PROVIDER.md` and `internal/skills/builtin/cline/` exist.

So BVR already consumes Cline *as a provider*. What is missing is Cline's
*product behavior*: catalog discovery, free-model distinction, plan/act
workflows, checkpoints, and context compaction. That matches the roadmap's
intent and requires no clarification.

---

## 7. Adjusted plan

Reordering, with reasons:

| Roadmap | Verdict | Change |
| --- | --- | --- |
| P0 finder/search | **Keep first** | Drop `project_id`; it is a build, not an enhancement. Reuse `loadSeq`-guarded `tea.Cmd` pattern. Reuse the `files` table approach for anything snapshot-shaped. |
| P1 OpenRouter | **Keep, rescope** | Provider is done. Build: live catalog fetch, TTL cache, free/paid labeling. |
| P2 Free model workflow | **Keep** | Depends on P1. Cheap once pricing metadata is surfaced. |
| P3 Agent loop | **Defer** | A single runtime already exists. Consolidating it is a refactor with no user-visible gain; do it only if P4 needs it. |
| P4 Plan/act + checkpoints | **Re-scoped** | `files` table already versions file content — check whether that is already a checkpoint mechanism before building one. |
| P5/P6 | **Keep as-is** | |

Add two items the roadmap does not mention but the audit surfaced:

- ~~**Fix or triage the `internal/catalog` FTS failure**~~ — **done**, see §1.
- **Add a cross-process lock to `projects.json`**, or fold projects into SQLite,
  to satisfy the roadmap's concurrency requirement.

---

## 8. Open questions for the maintainer

1. **P0 scope.** Build a real recursive fuzzy finder with content search, or
   ship P0 in two steps — (a) recents/pins/last-query persistence on the existing
   directory browser, (b) the search itself? *(Resolved: two steps, persist
   first.)*
2. **`internal/catalog` FTS failure.** *(Resolved: fixed in a separate change
   before integration work begins.)*
3. **Package naming.** The roadmap's `catalog` component name is taken. *(Resolved:
   `internal/modelcatalog`.)*
4. **Licensing.** *(Resolved: reimplement from behavior only, never copy source.)*
