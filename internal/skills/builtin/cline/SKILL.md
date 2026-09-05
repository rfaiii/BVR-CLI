---
name: cline
description: Use when a task is an autonomous, multi-step coding job — code review, refactors, fixing failing tests or CI, generating a feature end-to-end, migrating code, or bootstrapping a project — and the current local model is struggling or the user explicitly wants a stronger coding agent. Cline is a full autonomous terminal coding agent (`npm i -g cline`) that drives Claude, OpenAI, Google, OpenRouter, Bedrock, and others with real tool use, checkpoints, and plan/act modes.
user-invocable: true
license: Apache-2.0
compatibility: Requires Node.js/npm. Install with `npm install -g cline`. Works with Claude, OpenAI, Google, OpenRouter, AWS Bedrock, and OpenAI-compatible endpoints. MCP-supported.
---

# Cline — Autonomous Coding Agent

Cline is a standalone autonomous coding agent that runs in your terminal and
your editor. It plans, reads and edits files, runs shell commands, manages
MCP servers, takes checkpoints, and supports plan/act modes. When the current
BVR coder model is a small local model that struggles with multi-step coding,
delegate to Cline for reliable, autonomous execution.

## Install

```sh
npm install -g cline
```

This pulls a platform binary (no Node runtime required at install time). Verify:

```sh
cline --version
```

## Authenticate

Cline is provider-agnostic. Pick any provider and supply credentials:

```sh
cline auth                                   # interactive setup
cline auth --provider anthropic --apikey sk-... --modelid claude-sonnet-4-6
cline auth --provider openai-native --apikey sk-... --modelid gpt-5
cline auth --provider openrouter --apikey sk-...
```

## Invoke from BVR

Run Cline as an autonomous sub-agent through the `bash` tool. Cline inherits
BVR's current working directory, so no `--cwd` is normally needed.

Run a single prompt end-to-end with full tool use, auto-approved:

```sh
cline --yolo "Fix the failing tests and add a README example"
```

One-shot (asks before each tool call):

```sh
cline "Audit package.json and propose dependency updates"
```

Plan first, act later:

```sh
cline --plan "Design a migration strategy for this database schema"
```

Stream structured events for tooling:

```sh
cline --json "List all TODO comments" | jq -r '.event.text'
```

## Flags

| Flag | Notes |
|------|-------|
| `--yolo` | Auto-approve all tools; disable spawn/team tools; exit when done |
| `--auto-approve false` | Require review before each tool call |
| `--cwd <path>` | Working directory (inherits BVR's cwd by default) |
| `-P, --provider <id>` | Provider id (`cline`, `anthropic`, `openai`, `openrouter`, `bedrock`, `vertex`, `openai-native`, …) |
| `-m, --model <id>` | Model id, e.g. `anthropic/claude-sonnet-4-6` |
| `--thinking [none\|low\|medium\|high\|xhigh]` | Thinking budget when supported |
| `--retries <n>` | Max consecutive mistakes before halting (default 3) |
| `--json` | NDJSON event output (non-interactive) |
| `-i, --tui` | Interactive TUI (don't use from BVR; use headless/yolo instead) |
| `-s, --system <prompt>` | Override the system prompt |

## Rules and context

Cline reads `.clinerules` / `.clinerules/` files in the project root, the same
way BVR reads `AGENTS.md`. A project can therefore share one set of rules
between BVRs and Cline by keeping coding standards in `.clinerules` and
BVR-specific shell/workspace rules in `AGENTS.md`.

## MCP and checkpoints

- Cline shares MCP servers configured via `cline mcp`. It can use the same MCP
  servers BVR exposes, so tool access is consistent across both.
- Cline takes checkpoints per session; `cline history` lists past sessions and
  `/undo` rewinds workspace state. For a throwaway run from BVR, prefer a
  fresh `--data-dir` so checkpoints stay isolated:

```sh
cline --yolo --data-dir "$TMPDIR/cline-bvr" "Add unit tests for the auth handler"
```

## When to delegate

- Local Ollama models emit chat instead of tool calls or skip read-before-edit.
- The task spans many files and needs reliable edit/grep/bash coordination.
- You want checkpoints and plan/act separation.
- CI is red and needs a fix that survives multiple iterations.

After Cline finishes, capture key changes by reading the diff:

```sh
git --no-pager diff --stat
```
