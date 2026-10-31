import Foundation

/// The NODE hosts the app can see, which one is picked, and the live link.
///
/// The bearer token is saved in the Keychain under the host address, so the
/// next launch of the same host fills the field. The host-side driver passes
/// `InMemoryTokenStore` and does not touch the login keychain.
@Observable
final class NodeStore {
    /// Port assumed for a hand-typed address with no port, matching the
    /// `bvr node serve --host` default.
    static let defaultPort = 7777

    private(set) var discovered: [DiscoveredNode] = []
    private(set) var manualHosts: [DiscoveredNode] = []
    private(set) var isScanning = false

    private(set) var selectedID: String?
    var token = ""
    private(set) var state: NodeConnectionState = .offline
    private(set) var statusMessage: String?
    private(set) var probeOutput = ""

    private let discovery = NodeDiscovery()
    private let tokens: any NodeTokenStoring
    let client = NodeClient()

    /// Fires after the selection changes, including a cleared selection.
    var onSelectedHostChange: ((DiscoveredNode?) -> Void)?

    /// Fires with the text the probe is showing.
    var onProbe: ((String) -> Void)?

    var hosts: [DiscoveredNode] { discovered + manualHosts }

    var selectedHost: DiscoveredNode? {
        get { hosts.first { $0.id == selectedID } }
        set {
            guard let value = newValue else {
                select(nil)
                return
            }
            if !hosts.contains(where: { $0.id == value.id }) {
                manualHosts.append(value)
            }
            select(value.id)
        }
    }

    var isConnected: Bool { state == .online }

    var badgeLabel: String {
        guard let host = selectedHost else { return "No NODE" }
        let suffix = state == .online ? "" : " (\(state.rawValue))"
        return host.displayName + suffix
    }

    init(tokens: any NodeTokenStoring = KeychainTokenStore()) {
        self.tokens = tokens
        discovery.onNodesChanged = { [weak self] found in
            guard let self else { return }
            self.discovered = found
            // The first host seen is almost always the one just advertised,
            // so preselect it. A token already saved for that host is filled
            // in, and a manual choice is left alone.
            if self.selectedID == nil, let id = found.first?.id {
                self.select(id)
            }
        }
        client.onStateChange = { [weak self] next in
            guard let self else { return }
            self.state = next
            switch next {
            case .online: self.statusMessage = nil
            case .connecting: self.statusMessage = "Connecting to \(self.badgeLabel)…"
            case .failed: self.statusMessage = self.client.lastError?.localizedDescription ?? "Could not reach the host."
            case .offline: self.statusMessage = nil
            }
        }
    }

    /// Remembers `token` for the selected host. With no host yet, the value
    /// stays in memory until a host is selected.
    func updateToken(_ value: String) {
        token = value
        guard let host = selectedHost else { return }
        if !tokens.save(token, for: host.tokenKey) {
            statusMessage = "The token works for this launch, but the Keychain did not keep it."
        }
    }

    /// Switches hosts and loads that host's saved token.
    func select(_ id: String?) {
        let previous = selectedID
        guard previous != id else { return }
        if let previous, let host = host(for: previous) {
            tokens.save(token, for: host.tokenKey)
        }
        selectedID = id
        guard let selected = selectedHost else {
            onSelectedHostChange?(nil)
            return
        }
        let saved = tokens.token(for: selected.tokenKey)
        if !saved.isEmpty {
            token = saved
        } else if previous == nil, !token.isEmpty {
            tokens.save(token, for: selected.tokenKey)
        } else if previous != nil {
            token = ""
        }
        onSelectedHostChange?(selected)
    }

    func startScanning() {
        guard !isScanning else { return }
        isScanning = true
        statusMessage = "Looking for NODE agents…"
        discovery.start()
    }

    func stopScanning() {
        guard isScanning else { return }
        isScanning = false
        discovery.stop()
    }

    /// Accepts `host`, `host:port`, or a bare IPv6 literal in brackets.
    func addManualHost(_ input: String) {
        let trimmed = input.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return }
        guard let parsed = Self.parse(trimmed) else {
            statusMessage = "Needs a host, and a port if the agent is not on \(Self.defaultPort)."
            return
        }
        let node = DiscoveredNode(nodeID: "", displayName: "\(parsed.host):\(parsed.port)",
                                  hostName: parsed.host, port: parsed.port, proto: "1")
        if let existing = manualHosts.first(where: { $0.id == node.id }) {
            select(existing.id)
        } else {
            manualHosts.append(node)
            select(node.id)
        }
        statusMessage = nil
    }

    /// Returns the host in the exact spelling a `ws://` authority needs, so an
    /// IPv6 literal comes back with its brackets attached.
    static func parse(_ input: String) -> (host: String, port: Int)? {
        if input.hasPrefix("[") {
            guard let close = input.firstIndex(of: "]") else { return nil }
            let host = String(input[input.index(after: input.startIndex)..<close])
            guard !host.isEmpty else { return nil }
            let suffix = input[input.index(after: close)...]
            guard !suffix.isEmpty else { return ("[\(host)]", defaultPort) }
            guard suffix.hasPrefix(":"), let port = Int(suffix.dropFirst()), (1...65_535).contains(port) else { return nil }
            return ("[\(host)]", port)
        }
        // A bare IPv6 address is full of colons, so only one colon can be a port.
        let parts = input.split(separator: ":", omittingEmptySubsequences: false)
        switch parts.count {
        case 1:
            guard !parts[0].isEmpty else { return nil }
            return (String(parts[0]), defaultPort)
        case 2:
            guard !parts[0].isEmpty, let port = Int(parts[1]), (1...65_535).contains(port) else { return nil }
            return (String(parts[0]), port)
        default:
            return nil
        }
    }

    func connect() {
        // A second tap while the handshake is in flight used to cancel the
        // socket and then treat that cancel as a failure.
        if state == .connecting || state == .online { return }
        guard let host = selectedHost else { statusMessage = "Pick a host first."; return }
        guard let url = host.webSocketURL else { statusMessage = "The host address is not usable."; return }
        guard host.isProtocolCompatible else {
            statusMessage = "Host advertises NODE proto \(host.proto); this app speaks proto 1."
            return
        }
        if !tokens.save(token, for: host.tokenKey) {
            statusMessage = "The token works for this launch, but the Keychain did not keep it."
        } else {
            statusMessage = nil
        }
        probeOutput = ""
        // Browsing is only useful while there is nothing to talk to; it keeps a
        // multicast socket and a timer alive for the whole session.
        stopScanning()
        client.connect(to: url, token: token)
    }

    func disconnect() {
        client.disconnect()
    }

    /// Runs one cheap command on the host so the picker can prove the link works
    /// before anything real is typed into it.
    ///
    /// `uname` stays on the execute allow-list added in the NODE file-share
    /// work, so this probe works against that server and against main.
    func runProbe() {
        let request = NodeCommandRequest(command: "uname", args: ["-sr"], workingDir: nil)
        client.execute(request) { [weak self] result in
            guard let self else { return }
            switch result {
            case .success(let value):
                self.probeOutput = (value.stdout ?? "").trimmingCharacters(in: .whitespacesAndNewlines)
                if self.probeOutput.isEmpty { self.probeOutput = "Connected, but `uname` returned nothing." }
            case .failure(let error):
                self.probeOutput = error.localizedDescription
            }
            self.onProbe?(self.probeOutput)
        }
    }

    private func host(for id: String) -> DiscoveredNode? {
        hosts.first { $0.id == id }
    }
}
