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

- **iOS Companion App** in `mobile-cli/ios/BvrCompanion/` builds for the
  iOS 26.5 simulator with Xcode 26.6. Chat, CPU, RAM, and the model list still
  render `MockStore` data. NODE connect and the MCP URL do not.
- **NODE host connection is real**: `NodeDiscovery` browses `_bvr._tcp`,
  `NodeClient` speaks the `/v1/node/ws` frame protocol, and `NodeStore` plus
  `NodePickerView` (opened from the header badge) hold the host list, the
  selected host, the token, and the connection state. The stats row and
  header badge read that state instead of a hardcoded "Local Device".
- **The bearer token is stored in the Keychain per host** (`host:port`,
  service `com.richard.BvrCompanion.node-token`, this device only, not
  iCloud). The first launch of a host still needs the token typed in. The
  host-side driver uses an in-memory store so it does not write the Mac login
  keychain.
- **MCP does not dial `127.0.0.1` on launch.** An optional URL wins. Otherwise
  the app uses `http://<selected NODE host>:8000/sse` once a host is selected.
  On a phone, loopback is refused. The simulator allows it, because loopback
  there is the Mac. `NSAllowsLocalNetworking` permits the LAN `ws://` and
  `http://` URLs without turning ATS off for the rest of the internet.
- **WebSocket Origin** is set to `http://<host>:<port>` (or `https` for
  `wss`). URLSession sends no Origin unless the app sets one. Main accepts
  any Origin. The Origin check on `cursor/node-main-backup-files-7145` accepts
  a missing Origin and an Origin whose host equals the request Host, so the
  value above matches. `uname` remains on that branch's execute allow-list.
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
  BvrCompanion/BvrCompanion/NodeTokenStore.swift \
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
  device to confirm. A Mac does not receive its own advertisement.
- Chat prompts do not reach the NODE host yet: the agent executes shell
  commands per frame, while the app's chat is session-oriented, so wiring one to
  the other needs session-scoped frames (below).
- The bearer token is still typed the first time. There is no pairing flow.
  After that it is read back from the Keychain for that host.
- The picker's `Test \`uname\`` button is the only command the app sends.
  That command is on the default execute allow-list, so the probe works when
  the agent is not running with `--unrestricted-exec`.
- MCP is not the NODE protocol. It only connects when a NODE host or an
  explicit URL is available, and a phone will not use a loopback URL.
- Android client is not yet scaffolded.

### Next steps

- Session-scoped frames, then route chat input to the attached session.
- Pair the token instead of typing it the first time. The Keychain already
  keeps a token that was typed. GGWave or a QR code both already exist as
  primitives elsewhere in the app.
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
2. The companion browses that service type, resolves a host and port, and
   dials `ws://<host>:<port>/v1/node/ws` with `Authorization: Bearer <token>`
   and `Origin: http://<host>:<port>`. The token for that host is loaded from
   the Keychain when one was saved.
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
