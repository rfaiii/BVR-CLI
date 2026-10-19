# ADR 0001: What BVR borrows from Crush and Cline, and what it reimplements

- **Status:** Accepted
- **Date:** 2026-10-18
- **Context:** [BVR_Crush_Cline_Integration_Roadmap.md](../BVR_Crush_Cline_Integration_Roadmap.md)
- **Audit:** [PHASE0_AUDIT.md](../PHASE0_AUDIT.md)

## Context

BVR-CLI is an existing, shipping Go/Bubble Tea AI development environment. Two
external projects are proposed as references:

- **Crush** (`charmbracelet/crush`) — a Go terminal coding agent with a
  comparable architecture.
- **Cline** (`cline/cline`) — a TypeScript IDE extension plus CLI/SDK work.

The roadmap proposes evolving BVR using patterns from both, and explicitly
warns against a wholesale source merge of Cline into BVR.

The Phase 0 audit changed the shape of the problem significantly:

- BVR already has a single agent runtime, session persistence, permissions,
  hooks, skills, MCP, LSP, and Ollama support.
- BVR **already integrates OpenRouter for inference** via
  `charm.land/fantasy/providers/openrouter`, and already integrates **Cline's
  hosted gateway** as a first-class provider.
- The "file finder" is a single-directory browser with no search and no
  persistence, not the capable search tool the roadmap assumed.
- BVR keeps **one SQLite database per project**, which removes the need for the
  `project_id` column the roadmap's candidate schema specifies.
- The roadmap's proposed `catalog` component name is already taken by an
  unrelated product/image-lookup database.

## Decision

### 1. Borrow behavior and shape, not source

Adopt the **behavior** of Cline (model discovery UX, free/paid labeling,
plan/act workflows, checkpoints, context compaction) and the **architectural
shape** of Crush (provider-neutral agent runtime, pub/sub decoupling, migrated
SQLite persistence, permission-aware tools).

**Do not** copy Cline's TypeScript runtime into BVR. **Do not** import Crush's
`internal/...` packages — Go's `internal` rule makes them non-importable across
module boundaries by design, and vendoring them would create a permanent
maintenance and license obligation.

Where an external implementation is a useful reference, **reimplement from the
public shape and write our own tests**. Examples already identified:

- `catwalk/cmd/openrouter/main.go` is the reference for the live catalog fetch,
  endpoint selection, and pricing normalization. Reimplement the shape; do not
  vendor the file.
- Crush's `internal/agent` tool loop and `internal/db` migration conventions are
  patterns to follow, not packages to link.

### 2. One runtime, extended — not replaced

BVR's existing `internal/agent` runtime is the single agent loop. Crush patterns
are adopted *inside* it. No parallel "Crush agent" is introduced.

### 3. Persistence follows BVR's existing conventions

- Project scoping is **implicit in the per-project `bvr.db`**. New tables do
  **not** carry a `project_id` column.
- Keys are project-relative paths.
- Migrations are goose + sqlc, matching `internal/db/migrations/` and
  `internal/db/sql/`.
- New SQL lives in `internal/db/sql/` and is sqlc-generated, not hand-written.

### 4. Naming must not collide with the existing tree

The roadmap's `catalog` component is renamed. Model discovery will live in a new
package that does not collide with `internal/catalog` (which is a product /
image-lookup database). **Proposed: `internal/modelcatalog`.** This is subject to
maintainer confirmation.

### 5. Safety rules are non-negotiable

These are accepted as constraints on all implementation work, not as future
ideas:

- Never silently fall back from a free model to a paid model. If a selected
  model disappears or errors, stop and ask.
- Pricing that is unknown is **excluded** from free-only selection, not assumed
  free.
- "Free" is a time-stamped observation from provider metadata, never a
  guarantee. The UI shows when the catalog was last refreshed.
- No API keys or authorization headers in SQLite, config files, logs, or error
  messages.
- No filesystem I/O in Bubble Tea `Update`/`Draw`. Long work runs in a
  `tea.Cmd` and returns a typed message, with a sequence guard to discard stale
  results. *(This pattern already exists in `filebrowser.go`; keep it.)*
- Tool arguments are validated before execution. Model output is never trusted
  as shell input.
- Migrations and checkpoint restores must be safe against partial failure.

### 6. Tests never depend on live APIs

Unit and CI tests use fixtures and `httptest`. Live catalog/OpenRouter smoke
tests are opt-in and clearly separated.

### 7. Never discard errors

Errors from persistence writes must be returned, not assigned to `_`. The
`internal/catalog` FTS index was silently empty on `main` for exactly this
reason — see the Phase 0 audit §1. This applies with particular force to the
integration work, because a swallowed error in a catalog fetch or a search write
produces the same silent total feature failure.

## Consequences

**Positive**

- Every change is a small, reviewable diff against a working application.
- No second language runtime, no vendored fork to rebase.
- BVR's identity, existing features, and packaging are preserved by
  construction — nothing has to be re-added later.

**Negative**

- We forgo Cline's mature editor integrations and its battle-tested TypeScript
  agent loop. Any behavior we take from it is a reimplementation and must earn
  its own test coverage.
- Crush's `internal` packages cannot be diffed against ours over time. Pattern
  drift is a maintenance cost we accept deliberately.

**Neutral**

- `projects.json` is guarded by an in-process mutex only, with no cross-process
  lock. The roadmap requires safe concurrent writes.

## Resolved questions

1. P0 scope: **two steps — persist first, then search.**
2. `internal/catalog` FTS failure: **fixed as a separate change** before
   integration work began, so the branch starts from a green baseline.
3. Package name: **`internal/modelcatalog`**.
4. License posture: **reimplement from behavior only; never copy source.** Crush
   and Cline are read as behavioral references, and any borrowed shape is
   rewritten natively in BVR's idiom with BVR's own tests.
