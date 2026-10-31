# BVR CLI Companion — iOS

The iOS companion app for BVR-CLI: a remote control for a `bvr-cli` NODE agent
running on your Mac or Linux box.

## Layout

```
ios/
  BvrCompanion/
    BvrCompanion.xcodeproj/   the project — open this
    BvrCompanion/             sources, Assets.xcassets, Fonts
    BvrCompanionTests/        unit tests for address parsing, Keychain, MCP URLs
    BvrCompanionUITests/      manual-address connect against a live NODE agent
    Info.plist
  node-driver/                host-side driver for the networking stack
```

The target uses an Xcode 16-style `PBXFileSystemSynchronizedRootGroup`, so every
file under `BvrCompanion/` is compiled automatically. Do not add a second copy
of a type to make something compile — duplicate declarations inside the
synchronized folder break the build. Tests live in sibling folders so they are
not compiled into the app.

## Build and run

Open `BvrCompanion.xcodeproj` and run the `BvrCompanion` scheme on a simulator,
or headlessly:

```sh
xcodebuild -scheme BvrCompanion -sdk iphonesimulator -configuration Debug \
  -destination 'generic/platform=iOS Simulator' CODE_SIGNING_ALLOWED=NO build
```

Signing is automatic (`DEVELOPMENT_TEAM = A6EF93MTSU`,
`PRODUCT_BUNDLE_IDENTIFIER = com.richard.BvrCompanion`). That team matches the
Apple Development certificate already on this Mac. A simulator build does not
need a provisioning profile. A device build does, and the iPhone has to be
unlocked and trusted before Xcode will install.

`Info.plist` is the real build input (`INFOPLIST_FILE` is set alongside
`GENERATE_INFOPLIST_FILE`). It carries `NSLocalNetworkUsageDescription`,
`NSBonjourServices` (`_bvr._tcp`), `NSAllowsLocalNetworking` (so `ws://` and
the MCP `http://` URL are allowed on the LAN and on loopback), and `UIAppFonts`
for both bundled JetBrains Mono Nerd Font faces. The Bonjour service type must
stay in sync with `node.ServiceType` in the Go module and with
`NodeDiscovery.serviceType`.

Unit tests:

```sh
xcodebuild -scheme BvrCompanion \
  -destination 'platform=iOS Simulator,name=iPhone 17,OS=26.5' \
  -only-testing:BvrCompanionTests test
```

The UI test types a host and token into the NODE sheet and runs `uname`. It
skips unless a NODE agent is already listening and the runner has
`TEST_RUNNER_BVR_TEST_HOST` and `TEST_RUNNER_BVR_TEST_TOKEN` set.

### Sources

| File | Responsibility |
| --- | --- |
| `NodeDiscovery.swift` | browses `_bvr._tcp`, reads the `proto`/`id`/`name` TXT record |
| `NodeClient.swift` | the `/v1/node/ws` frame protocol: `ping`, `execute`, `stdout`, `stderr`, `result` |
| `NodeTokenStore.swift` | Keychain (app) or in-memory (tests and the host driver) token store |
| `NodeStore.swift` | host list, selection, per-host token, connection state |
| `NodePickerView.swift` | the sheet behind the header badge: scan, add by address, connect, test |
| `MCPEndpoint.swift` | chooses the MCP SSE URL from an override or the NODE host |
| `MCPClient.swift` | FastMCP SSE client. It does not hardcode `127.0.0.1` |
| `MockStore.swift` | chat, CPU, RAM, and the model list. NODE and MCP are live |

## Connecting to a host

Start the agent on the machine running `bvr-cli`:

```sh
bvr-cli node serve --host 0.0.0.0:7777 --token "$BVR_NODE_TOKEN" --advertise
```

`--advertise` publishes `_bvr._tcp.local`, which the app browses on launch; iOS
asks for local-network permission the first time. Tap the header badge to pick a
host, type the token, and connect. The token is saved in the Keychain for that
`host:port` and filled in on the next launch. If the network blocks multicast,
add `host` or `host:port` in the picker instead.

The WebSocket request sets `Origin` to `http://<host>:<port>` (or `https` for
`wss`). Main's agent accepts any Origin. The agent branch that checks Origin
(`cursor/node-main-backup-files-7145`) allows a missing Origin and an Origin
whose host equals the request Host, so this value is accepted there too.
URLSession sends no Origin unless the app sets one.

`Test uname` runs `uname -sr`. That command is on the default execute
allow-list in that same branch, and main still accepts every command name.

MCP is separate from NODE. Leave the MCP field empty to use
`http://<selected NODE host>:8000/sse`, or type another `http`/`https` URL.
On a phone the app refuses `127.0.0.1` and `localhost`, because those addresses
are the phone. The simulator still allows loopback, since that is the Mac.
The app does not connect to MCP until a host or an override exists.

## Testing the networking without a simulator

A Mac does not receive its own mDNS advertisements, so browsing cannot be
verified on the machine that is advertising — `dns-sd -B _bvr._tcp` finds nothing
there even for a service registered with `dns-sd -P`. The rest of the stack can
be driven for real by compiling the app sources into a host binary:

```sh
cd mobile-cli/ios
xcrun --sdk macosx swiftc -o /tmp/nodedrv node-driver/main.swift \
  BvrCompanion/BvrCompanion/NodeDiscovery.swift \
  BvrCompanion/BvrCompanion/NodeClient.swift \
  BvrCompanion/BvrCompanion/NodeTokenStore.swift \
  BvrCompanion/BvrCompanion/NodeStore.swift
BVR_TOKEN=<token> BVR_MANUAL=<host>:<port> /tmp/nodedrv
```

`node-driver/main.swift` lives beside the project, never inside
`BvrCompanion/BvrCompanion/`: the target compiles every Swift file under that
folder and top-level code there fails the app build. `BVR_MANUAL` feeds the
address through `NodeStore.addManualHost`, the same path the picker's manual
field uses. The driver passes `InMemoryTokenStore`, so this check does not
write the Mac login keychain. The app uses `KeychainTokenStore`. See
[`docs/MOBILE.md`](../../docs/MOBILE.md) for the frame protocol and what is
still missing on the app side.

A Debug simulator launch can take the same variables. `simctl` only forwards
environment variables prefixed with `SIMCTL_CHILD_`:

```sh
SIMCTL_CHILD_BVR_MANUAL=127.0.0.1:7777 \
SIMCTL_CHILD_BVR_TOKEN=<token> \
SIMCTL_CHILD_BVR_PROBE=1 \
  xcrun simctl launch --terminate-running-process \
  --stdout=/tmp/bvr-sim-out.txt --stderr=/tmp/bvr-sim-err.txt \
  booted com.richard.BvrCompanion
```

`BVR_PROBE=1` prints `BVR_PROBE` followed by the `uname -sr` output. Release
builds do not read these variables.

## Theme

Dark ultra-contrast material theme, defined in `BvrCompanion/Theme.swift`:

| Token | Value |
| --- | --- |
| Primary / success | `#39F66B` |
| Secondary | `#B972FF` |
| Background | `#000000` |
| Surface | `#121212` |
| Border | `#2C2C2C` |
| Text primary | `#E0E0E0` |
| Text secondary | `#B0BEC5` |

## Requirements

- Xcode 26.6 with the iOS 26.5 SDK. The deployment target is iOS 26.5, the SDK
  this project builds with. An older Xcode does not have that SDK, so the
  target stays at 26.5 instead of being lowered. A phone on iOS 26.5 or later
  (this Mac's iPhone is on 26.6.1) can install the build.
- SwiftUI
- Local network permission on first browse
- For a physical iPhone: the phone unlocked, trusted, and on the same Wi-Fi as
  the Mac. Signing team `A6EF93MTSU` is already set and matches the local
  Apple Development certificate. In Xcode, pick the `BvrCompanion` scheme and
  the phone, then Run. Allow Local Network. Start the agent with
  `--host 0.0.0.0:7777` so the phone can reach it, then pick the Mac in the
  sheet or type its LAN address and port. The first token is typed by hand;
  later launches fill it from the Keychain.

## Running on a physical iPhone

The phone named `R1CH$ (2)` (iPhone 14 Plus, iOS 26.6.1) was offline when this
was last checked, so a device install was not attempted and no signing
settings were changed.

1. Connect the iPhone, unlock it, and tap Trust if the Mac asks.
2. Open `BvrCompanion.xcodeproj` in Xcode 26.6. Select the `BvrCompanion`
   scheme and the phone. The signing team is already `A6EF93MTSU`. If signing
   fails, choose that team under Signing & Capabilities. Leave the bundle id
   `com.richard.BvrCompanion` as it is.
3. Run. On the first launch, allow Local Network access.
4. On the Mac, on the same Wi-Fi:

   ```sh
   bvr-cli node serve --host 0.0.0.0:7777 --token "$BVR_NODE_TOKEN" --advertise
   ```

   `uname` works against main and against the execute allow-list on the
   file-share branch. Other command names on that branch need `--allow-exec`
   or `--unrestricted-exec`.
5. In the app, tap the NODE badge. If the Mac is listed, select it. This Mac
   does not see its own Bonjour ad; the phone should. If it does not, tap Add
   and enter the Mac's LAN address, for example `192.168.1.20:7777`.
6. Type the token, tap Connect, then Test `uname`. The next launch fills the
   token for that host from the Keychain.
7. MCP: leave the URL empty to use `http://<that host>:8000/sse`. Run the MCP
   server on the Mac bound to the LAN, not only to `127.0.0.1`. The phone
   refuses a localhost MCP URL.
