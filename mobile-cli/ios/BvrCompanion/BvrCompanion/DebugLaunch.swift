import Foundation

#if DEBUG
/// Debug-only stand-in for the picker's manual-address path.
///
/// `simctl launch` passes `SIMCTL_CHILD_BVR_MANUAL` and
/// `SIMCTL_CHILD_BVR_TOKEN` through as `BVR_MANUAL` and `BVR_TOKEN`.
/// `BVR_PROBE=1` runs `uname -sr` after the handshake and prints
/// `BVR_PROBE` on stdout. Release builds omit this file's body.
enum DebugLaunch {
    static func apply(to store: MockStore) {
        let env = ProcessInfo.processInfo.environment
        if let mcp = env["BVR_MCP_URL"] {
            store.mcpOverride = mcp
        }
        let manual = env["BVR_MANUAL"] ?? ""
        guard !manual.isEmpty else {
            if env["BVR_MCP_URL"] != nil {
                store.refreshMCP()
            }
            return
        }
        if let token = env["BVR_TOKEN"] {
            store.node.updateToken(token)
        }
        let existing = store.node.client.onStateChange
        store.node.client.onStateChange = { state in
            existing?(state)
            print("BVR_STATE \(state.rawValue)")
            if let message = store.node.statusMessage {
                print("BVR_MESSAGE \(message)")
            }
            fflush(stdout)
            if state == .online, env["BVR_PROBE"] == "1" {
                store.node.runProbe()
            }
        }
        store.node.onProbe = { output in
            print("BVR_PROBE \(output)")
            fflush(stdout)
        }
        store.node.addManualHost(manual)
        store.node.connect()
    }
}
#endif
