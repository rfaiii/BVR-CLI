# BVR-CLI Mobile Companion

This document is the consolidated reference for the mobile work: the iOS and
Android companion plans and the host-side companion bridge design. It merges
the former `MOBILE-CLI.md` and `BVR-SERVER.md`. Device-to-device transport
(what the companion rides on) is documented in [`NODES.md`](NODES.md).

## Companion app plan

Plans and outline for the iPhone and Android companion variants, plus the
host-side companion bridge that connects them to `bvr-cli`.

### Goals

- Provide a mobile-first control surface for BVR-CLI sessions.
- Keep the existing Go CLI intact; add a lightweight companion bridge instead
  of rewriting the core in another runtime.
- Support secure local/network transport between the host CLI and mobile
  clients.

### Topology

- Host machine runs the normal BVR-CLI.
- Optional companion server/bridge runs alongside the CLI to expose secure
  mobile-facing commands.
- Mobile clients connect via a private/link-local or authenticated channel.

### Essentials

1. Host-side bridge module
2. Mobile client variant(s)
3. Secure auth/transport model
4. Session and telemetry bridge
5. Installation and docs

### Repo structure

```
mobile-cli/
  README.md
  ios/BvrCompanion/     Xcode project, sources, Info.plist
  ios/node-driver/      host-side driver for the app's networking stack
  android/
  web-preview/
```

### Roadmap

1. Capture requirements and constraints.
2. Define the bridge protocol and data model.
3. Prototype the host-side entrypoint.
4. Prototype one mobile client variant (iOS companion).
5. Build a web-based HTML/CSS prototype for rapid UI iteration without Xcode.
6. Migrate desktop settings and commands into the mobile surface.
7. Add tests, docs, and packaging.

### Current status

- **iOS Companion App** in `mobile-cli/ios/BvrCompanion/` builds clean for the
  simulator: chat, command palette, theme switching, MCP client, and stats
  surfaces. Chat still renders `MockStore` data.
- **NODE host connection is real**: `NodeDiscovery` browses `_bvr._tcp`,
  `NodeClient` speaks the `/v1/node/ws` frame protocol, and `NodeStore` plus
  `NodePickerView` (opened from the header badge) hold the host list, the
  selected host, the token field, and the connection state. The stats row and
  header badge read that state instead of a hardcoded "Local Device".
- **Host-side bridge** is the NODE agent, not a separate companion server:
  `bvr node serve --host 0.0.0.0:7777 --token <token> --advertise`.
  See [`NODES.md`](NODES.md#local-network-discovery).
- **Discovery** works end to end on macOS: the agent publishes
  `_bvr._tcp.local` with `proto`/`id`/`name` TXT records and `bvr node scan`
  resolves it back to a reachable IPv4 and port. The app's `Info.plist` already
  declares `NSBonjourServices` and `NSLocalNetworkUsageDescription`.
- **Web UI Prototype** built in `mobile-cli/web-preview/` for instant
  zero-compile UI testing.
- **FastMCP Server** implemented in `mcp-server/`, exposing tools over SSE.

### Testing the app's networking without an iPhone

A Mac cannot receive its own mDNS advertisements, so Bonjour browse cannot be
exercised on the advertising host; `dns-sd -B _bvr._tcp` returns nothing there
even for a service registered with `dns-sd -P`. Everything behind discovery can
still be driven for real with `mobile-cli/ios/node-driver/main.swift`, which
compiles against the app's own sources:

```sh
cd mobile-cli/ios
xcrun --sdk macosx swiftc -o /tmp/nodedrv node-driver/main.swift \
  BvrCompanion/BvrCompanion/NodeDiscovery.swift \
  BvrCompanion/BvrCompanion/NodeClient.swift \
  BvrCompanion/BvrCompanion/NodeStore.swift
BVR_TOKEN=<token> BVR_MANUAL=<host>:<port> /tmp/nodedrv
```

`BVR_MANUAL` feeds the address through `NodeStore.addManualHost`, the same path
the picker's manual field uses. Verified with this recipe: connect → handshake
`ping` → `online`, `uname -sr` returning `Darwin 25.5.0`, a wrong token
reporting `The host rejected the token.`, a closed port reporting
`Could not connect to the server.`, and a 200-line `seq` command arriving as
200 in-order `stdout` frames.

### Current limitations

- Bonjour browse is the one path that still needs a second machine or a real
  device to confirm.
- Chat prompts do not reach the NODE host yet: the agent executes shell
  commands per frame, while the app's chat is session-oriented, so wiring one to
  the other needs session-scoped frames (below).
- The bearer token has to be typed by hand; there is no pairing flow. It is held
  in memory only — never Keychain or UserDefaults — so it is retyped each launch.
- The picker's `Test \`uname\`` button is the only command the app sends.
- Android client is not yet scaffolded.

### Next steps

- Session-scoped frames, then route chat input to the attached session.
- Pair the token instead of typing it — GGWave or a QR code both already exist
  as primitives elsewhere in the app.
- Refine command palette parity: `/themes`, `/node`, `/finder`, `/mcp`, `/cd`.
- Lock down the minimum viable command set.

## Host-side bridge

The mobile apps are remote controls for the `bvr-cli` already running on the
host, so the bridge is the NODE agent rather than a second server:

```sh
bvr node serve --host 0.0.0.0:7777 --token "$BVR_NODE_TOKEN" --advertise
```

**Connection flow:**

1. The agent starts an HTTP listener and, with `--advertise`, publishes
   `_bvr._tcp.local` with `proto`, `id`, and `name` TXT records.
2. The companion browses that service type, resolves an IPv4 and port, and
   dials `ws://<host>:<port>/v1/node/ws` with `Authorization: Bearer <token>`.
3. The agent upgrades the connection and exchanges `WSFrame` JSON messages:
   the client sends `ping` or `execute`, the host answers with `stdout` /
   `stderr` lines and a final `result`.
4. Without a token, `--advertise` refuses to start: the agent runs arbitrary
   commands for any client that authenticates.

Frame types and the HTTP endpoints are specified in
[`NODES.md`](NODES.md#websocket-protocol); `internal/node/websocket.go` is the
implementation on both sides.

### Remaining host-side work

- [ ] Feed `Scan` results into the device registry, and persist devices across
      runs (the registry is currently per-process).
- [ ] Pair the token instead of typing it.
- [ ] Session-scoped frames (`SESSION_LIST`, `SESSION_CONNECT`) so the phone
      attaches to a running chat instead of executing shell commands.
