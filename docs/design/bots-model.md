# Backing a bot with a real model

Report only. Nothing here is wired up. The live `/bots` chat still calls
`bots.Fake` through `bots.Backend`.

## What already exists

`bots.Backend` is the seam:

```go
type Backend interface {
    List() []Bot
    Send(ctx context.Context, botID, msg string) (<-chan Reply, error)
}
```

The dialog does not know whether `Send` is canned or a model. It already
cancels the context, streams `Reply` chunks, typewrites them, and plays
the bot's WAV cues. A real backend can drop in beside `Fake` without a
new screen.

The app's model path is separate from that interface:

- **Provider catalog.** `internal/config` loads `ProviderConfig` (id,
  type, base URL, API key or OAuth, extra headers). Catwalk
  (`internal/config/catwalk.go`, `charm.land/catwalk`) supplies the model
  list and default options. The selected large model is what the main
  chat already resolved.
- **Agent.** `internal/agent/coordinator.go` `buildAgent` builds a
  `SessionAgent` with that large model, then `prompt.Build` fills the
  coder system prompt and `mcp.WaitForInit` gates the tool list.
  `sessionAgent.Run` (`internal/agent/agent.go`) calls
  `fantasy.NewAgent` and `agent.Stream`. It also writes the user message
  and the assistant message onto a session.
- **Session.** `internal/session` `Create` inserts a SQLite session.
  Messages, cost, title generation, and the busy/cancel map hang off
  that id. The main TUI chat is a view of one session.
- **MCP.** Clients live under `internal/agent/tools/mcp`. Connected
  servers contribute tools and an `<mcp-instructions>` block appended
  inside `Run`. That is the coder agent's tool belt, not a personality.

## Cleanest model backend

Add a `Voice.Prompt` string per bot (Dingus level-headed, BroBro short
and fast, Pillo one sentence, Luvah kind, Gnarf caps and sound). Implement
`ModelBackend` with the same `List`/`Send` signature:

1. Resolve the large model the way `coordinator.buildAgentModels` already
   does. No second provider stack.
2. Call `fantasy.NewAgent` with that model, **no tools**, and a system
   prompt that is only `Voice.Prompt` plus "reply in one or two short
   lines." Do not call `prompt.Build` and do not append MCP instructions.
3. `agent.Stream` one user turn. Map text deltas onto `Reply.Text`. Keep
   `Emote` and `Boop` from the catalog, not from the model.
4. Leave `Fake` as the default until a model is configured, and keep the
   `DEMO` badge on the fake path.

The dialog, sounds, and small sprites stay put. `Send` is already the
boundary.

This stays out of SQLite. History is whatever the dialog holds in
memory, which matches the current one-off NES window.

Scope: one backend type, one prompt field on `Voice`, and a constructor
the UI calls instead of `bots.NewFake` when a model is selected. It does
not touch permissions, the session table, or the main chat list.

## Could a bot chat be a full session?

Yes. It should not be the first step.

`session.Create` plus `sessionAgent.Run` would give the bot a real
transcript, a title, cost, cancel, and the main chat's message list.
The NES window would stop owning `[]chatLine` and would render that
session instead. MCP tools and the coder system prompt come along with
`Run` unless they are explicitly stripped, so a "full session" bot
becomes another coder unless the call is built with an empty tool list
and the personality prompt.

That crosses the session service, the message service, the coordinator's
busy/cancel map, permissions, and the dialog. It is the path for "this
bot is a saved conversation I can reopen from the session list." It is
the wrong path for "say one line, get one line."

Do the `ModelBackend` first. A later session-backed bot can call the
same stream helper and then persist the turn.
