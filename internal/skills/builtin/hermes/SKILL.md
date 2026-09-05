---
name: hermes
description: Use when the task is an autonomous, long-running coding or research job where the user wants the Hermes agent — notable for self-improving skills, persistent cross-session memory, cloud backends (Docker/SSH/Modal/Daytona/Sandbox), and a built-in scheduler. Hermes (`curl .../install.sh | bash`) is a full terminal + gateway agent by Nous Research that works with 300+ models via Nous Portal or any OpenAI-compatible endpoint.
user-invocable: true
license: MIT
compatibility: "Requires the Hermes CLI. Install: curl -fsSL https://hermes-agent.nousresearch.com/install.sh | bash (Linux/macOS) or the PowerShell one-liner (Windows). Cross-platform; bundles uv, Python 3.11, Node.js, ripgrep, ffmpeg, and Git Bash on Windows."
---

# Hermes — Self-Improving Agent

Hermes is an autonomous agent built by Nous Research. It is not a single model
— it is an agent runtime that runs any model you choose, keeps state across
sessions, curates memory, creates and improves its own skills, and can run on a
local machine, a Docker container, or a serverless cloud backend. Hermes is
compatible with the <https://agentskills.io> open standard, the same format
BVR skills use, so project rules and skills can be shared.

## Install

Linux, macOS, WSL2, Termux:

```sh
curl -fsSL https://hermes-agent.nousresearch.com/install.sh | bash
source ~/.bashrc   # reload shell
hermes             # start chatting
```

Windows (PowerShell, native — no WSL required):

```powershell
iex (irm https://hermes-agent.nousresearch.com/install.ps1)
```

The installer handles uv, Python 3.11, Node.js, ripgrep, ffmpeg, and Git. Verify:

```sh
hermes --version
hermes doctor
```

One subscription, many providers:

```sh
hermes setup --portal   # OAuth to Nous Portal; covers model, web search, image gen, TTS, cloud browser
```

## Invoke from BVR

Hermes is a long-lived agent, not a one-shot CLI. To delegate work without
blocking BVR's terminal, run it as a backgrounded yolo/headless session or
schedule it, then check results from the shared project directory.

Run a fresh, isolated coding session (sandbox data dir so checkpoints stay
isolated from your main Hermes history):

```sh
hermes --data-dir "$TMPDIR/hermes-bvr" "Fix the failing CI on main and add a test"
```

Start the messaging gateway so you can steer the run from anywhere (Telegram,
Discord, Slack, WhatsApp, Signal), useful for long jobs:

```sh
hermes gateway start
hermes gateway setup   # picks a platform
```

Schedule recurring agentic work (the schedule writes results into the project):

```sh
hermes schedule create "Daily code review" \
  --cron "0 9 * * MON-FRI" \
  --prompt "Review PRs opened yesterday and summarize issues" \
  --workspace "$BVR_PROJECT_DIR"
```

## Slash commands (in the Hermes TUI or a message thread)

| Command | Effect |
|---------|--------|
| `/new`, `/reset` | Fresh conversation |
| `/model <provider:model>` | Switch provider/model |
| `/skills` or `/<skill-name>` | Browse or load a skill |
| `/compress` | Compact context |
| `/usage`, `/insights` | Token/run stats |
| `/retry`, `/undo` | Regenerate or rewind a turn |
| `/stop` | Interrupt current work |

## Run anywhere

Hermes backends: **local**, **Docker**, **SSH**, **Singularity**, **Modal**,
**Daytona**, and **Vercel Sandbox**. For a coding task, pick a backend with
the toolchain you need and let it persist between turns:

```sh
hermes --backend daytona "Scaffold a Go service with a Postgres migration"
```

## MCP and skills

- Hermes consumes MCP servers; it can use the same MCP servers BVR exposes.
- Hermes reads project rules from `AGENTS.md`, `CLAUDE.md`, `GEMINI.md`, and
  `BVR.md` (the same files BVR reads), plus `SOUL.md` for personality.
- Like BVR, Hermes follows the agentskills.io skill standard, so a skill you
  author for BVR is understood by Hermes and vice versa.

## When to delegate

- The task is long-running or long-lived and you want cross-session memory.
- You want the agent to learn and persist better behavior across runs.
- The work should run on a cloud backend (not your laptop) while you move on.
- You need scheduled/repeated automation that writes results back to the repo.
- You prefer the Nous Portal single subscription over managing five API keys.

After a Hermes run, inspect workspace changes:

```sh
git --no-pager diff --stat
git --no-pager log --oneline -n 5
```

> [!NOTE] Hermes does not yet ship a single "one-shot and exit" mode on every
> backend; for fire-and-forget coding from BVR prefer Cline `--yolo`, or run
> Hermes through a scheduler/gateway and poll the project directory.
