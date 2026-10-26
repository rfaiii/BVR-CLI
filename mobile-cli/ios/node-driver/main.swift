import Foundation

// Host-side driver for the companion's networking stack. Built from the app's
// sources so the WebSocket client, the store, and the manual-address path can be
// exercised without a simulator:
//
//   cd mobile-cli/ios
//   xcrun --sdk macosx swiftc -o /tmp/nodedrv \
//     node-driver/main.swift \
//     BvrCompanion/BvrCompanion/NodeDiscovery.swift \
//     BvrCompanion/BvrCompanion/NodeClient.swift \
//     BvrCompanion/BvrCompanion/NodeStore.swift
//   BVR_TOKEN=<agent token> BVR_MANUAL=<host>:<port> /tmp/nodedrv
//
// This file is deliberately NOT under BvrCompanion/: the target compiles every
// Swift file in that folder, and top-level code there fails the app build.
//
// Bonjour browsing cannot be checked this way. A host does not receive its own
// mDNS advertisements, so set BVR_MANUAL and point it at a `bvr node serve
// --advertise` process running somewhere else for the browse path.

let store = NodeStore()
store.token = ProcessInfo.processInfo.environment["BVR_TOKEN"] ?? ""
let manual = ProcessInfo.processInfo.environment["BVR_MANUAL"] ?? ""
let timeout = TimeInterval(ProcessInfo.processInfo.environment["BVR_TIMEOUT"] ?? "25") ?? 25

enum Phase {
    case browsing
    case running
}

var phase = Phase.browsing
let deadline = Date().addingTimeInterval(timeout)

func report(_ code: Int32) -> Never {
    print("hosts: " + store.hosts.map { "\($0.displayName) proto=\($0.proto) id=\($0.nodeID) url=\($0.webSocketURL?.absoluteString ?? "nil")" }.joined(separator: "; "))
    print("state: \(store.state.rawValue)")
    print("probe: \(store.probeOutput)")
    if let message = store.statusMessage { print("message: \(message)") }
    exit(code)
}

func tick() {
    switch store.state {
    case .online where phase == .browsing:
        phase = .running
        store.runProbe()
    case .offline where !store.hosts.isEmpty && phase == .browsing:
        store.connect()
    case .failed:
        report(1)
    default:
        break
    }
    if phase == .running, !store.probeOutput.isEmpty { report(0) }
    if Date() > deadline {
        print("timed out")
        report(1)
    }
}

Timer.scheduledTimer(withTimeInterval: 0.5, repeats: true) { _ in tick() }
if manual.isEmpty {
    store.startScanning()
} else {
    // The same path the picker's manual address field uses.
    store.addManualHost(manual)
    store.connect()
}
RunLoop.main.run()
