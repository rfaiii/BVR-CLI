# BVR CLI Companion — iOS

The iOS companion app for BVR-CLI: a remote control for a `bvr-cli` NODE agent
running on your Mac or Linux box.

## Layout

```
ios/
  BvrCompanion/
    BvrCompanion.xcodeproj/   the project — open this
    BvrCompanion/             sources, Assets.xcassets, Fonts
    Info.plist
  node-driver/                host-side driver for the networking stack
```

The target uses an Xcode 16 `PBXFileSystemSynchronizedRootGroup`, so every file
under `BvrCompanion/` is compiled automatically. Do not add a second copy of a
type to make something compile — duplicate declarations inside the synchronized
folder break the build.

## Build and run

Open `BvrCompanion.xcodeproj` and run the `BvrCompanion` scheme on a simulator,
or headlessly:

```sh
xcodebuild -scheme BvrCompanion -sdk iphonesimulator -configuration Debug \
  -destination 'generic/platform=iOS Simulator' CODE_SIGNING_ALLOWED=NO build
```

Signing is automatic (`DEVELOPMENT_TEAM = A6EF93MTSU`,
`PRODUCT_BUNDLE_IDENTIFIER = com.richard.BvrCompanion`), so a device build needs
a provisioning profile but a simulator build does not.

`Info.plist` is the real build input (`INFOPLIST_FILE` is set alongside
`GENERATE_INFOPLIST_FILE`) and carries the local-network keys the companion
needs: `NSLocalNetworkUsageDescription`, `NSBonjourServices` (`_bvr._tcp`), and
`UIAppFonts` for both bundled JetBrains Mono Nerd Font faces. The Bonjour
service type must stay in sync with `node.ServiceType` in the Go module and with
`NodeDiscovery.serviceType`.

### Sources

| File | Responsibility |
| --- | --- |
| `NodeDiscovery.swift` | browses `_bvr._tcp`, reads the `proto`/`id`/`name` TXT record |
| `NodeClient.swift` | the `/v1/node/ws` frame protocol: `ping`, `execute`, `stdout`, `stderr`, `result` |
| `NodeStore.swift` | host list, selection, token field, connection state |
| `NodePickerView.swift` | the sheet behind the header badge: scan, add by address, connect, test |
| `MockStore.swift` | everything the app still fakes (chat, CPU, RAM, models) |

## Connecting to a host

Start the agent on the machine running `bvr-cli`:

```sh
bvr-cli node serve --host 0.0.0.0:7777 --token "$BVR_NODE_TOKEN" --advertise
```

`--advertise` publishes `_bvr._tcp.local`, which the app browses on launch; iOS
asks for local-network permission the first time. Tap the header badge to pick a
host, type the token, and connect. If the network blocks multicast, add
`host` or `host:port` in the picker instead.

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
  BvrCompanion/BvrCompanion/NodeStore.swift
BVR_TOKEN=<token> BVR_MANUAL=<host>:<port> /tmp/nodedrv
```

`node-driver/main.swift` lives beside the project, never inside
`BvrCompanion/BvrCompanion/`: the target compiles every Swift file under that
folder and top-level code there fails the app build. `BVR_MANUAL` feeds the
address through `NodeStore.addManualHost`, the same path the picker's manual
field uses. See [`docs/MOBILE.md`](../../docs/MOBILE.md) for the frame protocol
and what is still missing on the app side.

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

- Xcode 16+, iOS 26.5 deployment target (set in the project)
- SwiftUI
- Local network permission on first connect
