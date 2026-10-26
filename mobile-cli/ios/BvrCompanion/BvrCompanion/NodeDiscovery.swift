import Foundation

/// A NODE agent found on the local network over Bonjour. Mirrors `node.Node`
/// in the Go module: same `_bvr._tcp` service type, same TXT keys.
struct DiscoveredNode: Identifiable, Hashable {
    var nodeID: String
    var displayName: String
    var hostName: String
    var port: Int
    var proto: String

    var id: String { nodeID.isEmpty ? "\(hostName):\(port)" : nodeID }

    /// mDNS host names are fully qualified, and a trailing dot is not valid in
    /// a URL authority.
    var host: String {
        hostName.hasSuffix(".") ? String(hostName.dropLast()) : hostName
    }

    var webSocketURL: URL? { URL(string: "ws://\(host):\(port)/v1/node/ws") }
    var isProtocolCompatible: Bool { proto.isEmpty || proto == "1" }
}

/// Browses `_bvr._tcp` and reports the agents it can see.
///
/// The Go side advertises this service from `bvr node serve --advertise`; the
/// app declares it in `NSBonjourServices`, and iOS only shows results after the
/// user grants local-network permission on the first browse.
///
/// Apple has marked `NetService` "API to be deprecated" in favour of
/// Network.framework, but `NWBrowser` hands back endpoints rather than the TXT
/// records this needs to check the NODE protocol version and node id, so the
/// older API stays until that moves.
final class NodeDiscovery: NSObject, NetServiceBrowserDelegate, NetServiceDelegate {
    static let serviceType = "_bvr._tcp"
    static let serviceDomain = "local."

    var onNodesChanged: (([DiscoveredNode]) -> Void)?

    private(set) var nodes: [DiscoveredNode] = []
    private let browser = NetServiceBrowser()
    // The browser hands over a fresh service per announcement and nothing else
    // retains it, but resolution is asynchronous -- an unreferenced service
    // disappears before its reply arrives.
    private var resolving: [NetService] = []
    private var started = false

    func start() {
        guard !started else { return }
        started = true
        browser.delegate = self
        browser.searchForServices(ofType: Self.serviceType, inDomain: Self.serviceDomain)
    }

    func stop() {
        guard started else { return }
        started = false
        browser.stop()
        resolving.forEach { $0.stop() }
        resolving.removeAll()
    }

    func netServiceBrowser(_ browser: NetServiceBrowser, didFind service: NetService, moreComing: Bool) {
        service.delegate = self
        resolving.append(service)
        service.resolve(withTimeout: 5)
    }

    func netServiceBrowser(_ browser: NetServiceBrowser, didRemove service: NetService, moreComing: Bool) {
        nodes.removeAll { $0.hostName == service.hostName && $0.port == Int(service.port) }
        publish()
    }

    func netServiceDidResolveAddress(_ sender: NetService) {
        if let node = node(from: sender) {
            upsert(node)
        }
        release(sender)
    }

    func netService(_ sender: NetService, didNotResolve errorDict: [String: NSNumber]) {
        // The SRV answer usually carries the host name and port even when the
        // A record did not answer in time, so show what we have: the URL is
        // built from the host name, which iOS resolves over mDNS itself.
        if let node = node(from: sender) {
            upsert(node)
        }
        release(sender)
    }

    private func release(_ service: NetService) {
        service.delegate = nil
        resolving.removeAll { $0 === service }
    }

    private func node(from service: NetService) -> DiscoveredNode? {
        let port = Int(service.port)
        guard port > 0, let hostName = service.hostName, !hostName.isEmpty else { return nil }
        let txt = txtValues(for: service)
        // An advertisement with `name=` unset publishes an empty TXT value, and
        // the instance name is the next best thing to show a human.
        let advertised = txt["name"].flatMap { $0.isEmpty ? nil : $0 }
        let instance = service.name.isEmpty ? hostName : service.name
        return DiscoveredNode(
            nodeID: txt["id"] ?? "",
            displayName: advertised ?? instance,
            hostName: hostName,
            port: port,
            proto: txt["proto"] ?? "")
    }

    private func txtValues(for service: NetService) -> [String: String] {
        guard let data = service.txtRecordData() else { return [:] }
        // The overlay types every TXT value as raw bytes, decoded here because
        // the Go publisher writes plain ASCII.
        return NetService.dictionary(fromTXTRecord: data).reduce(into: [String: String]()) { values, pair in
            if let value = String(data: pair.value, encoding: .utf8) {
                values[pair.key] = value
            }
        }
    }

    private func upsert(_ node: DiscoveredNode) {
        if let index = nodes.firstIndex(where: { $0.id == node.id }) {
            nodes[index] = node
        } else {
            nodes.append(node)
        }
        publish()
    }

    private func publish() {
        nodes.sort { $0.displayName.localizedCaseInsensitiveCompare($1.displayName) == .orderedAscending }
        onNodesChanged?(nodes)
    }
}
