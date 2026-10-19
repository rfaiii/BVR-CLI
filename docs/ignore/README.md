# Ignored documents

Internal and working notes that are **deliberately not published** to GitHub.

Everything in this folder is matched by the `docs/ignore/` rule in the root
`.gitignore`, so these files stay on disk for local reference and are never
committed or pushed.

## Why this folder exists

BVR-CLI ships a large documentation set, and a few documents are working notes
rather than product documentation. Previously each one needed its own
`.gitignore` line scattered across the file, and they were scattered between the
repository root and `docs/`. Anything ignored that is a document now lives here,
so there is a single place to review them and a single ignore rule to maintain.

## Contents

| File | What it is |
| --- | --- |
| `CONTEXT.md` | Working context and intent snapshots for ongoing development. |
| `REMOVAL.md` | Draft removal/deprecation plan. Not user-facing yet. |
| `LAUNCH_TASKLIST.md` | Release readiness and beta work in progress. |
| `ASSET_REQUIREMENTS.md` | Draft asset spec, largely superseded by `docs/IMG-RESOURCES.md`. |
| `BVR_Crush_Cline_Integration_Roadmap.md` | Proposed Crush/Cline integration roadmap. Not accepted work. |
| `PHASE0_AUDIT.md` | Audit backing that roadmap. Moves with it. |
| `adr/0001-borrowed-vs-reimplemented.md` | ADR for the same initiative. Moved with it so no published file links into this folder. |
| `antigravity/rules.md` | Local editor-tool rules for Antigravity. Machine-specific. |

Everything above except the working notes was previously tracked in the
repository. These were moved here in 1.2.7 to shrink the published docs surface;
nothing was lost. Any of them can be restored with
`git mv docs/ignore/<name>.md docs/<name>.md`.

## Rules for this folder

- **Keep it small.** If a document is stable, user-facing, and useful to
  contributors, promote it out of here into `docs/` proper and delete the
  `.gitignore` entry mentality — promotion is a `git mv` away.
- **Do not link to these from published docs.** A tracked file that points into
  this folder will render a broken link on GitHub, because the target is not
  published. When a document moves here, fix its inbound links in the same
  change.
- **Review before deleting.** These are working notes, not litter. Weed them out
  deliberately; do not clear the folder by reflex.
- **Secrets stay out.** These files are still committed-adjacent working files.
  Never paste API keys or tokens into anything in this folder, even though the
  folder is gitignored.

## Auditing

See which ignored files exist (including non-markdown ones):

```sh
git status --ignored --short
```

To move a document out of hiding and start tracking it:

```sh
git add -f docs/ignore/<name>.md
```
