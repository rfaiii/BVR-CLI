import Foundation

/// Where the companion looks for the FastMCP SSE server.
///
/// The NODE agent does not serve MCP. The MCP process listens on port
/// 8000 of the same machine. On a phone, `127.0.0.1` is the phone, so a
/// missing host stays disconnected instead of dialing localhost.
enum MCPEndpoint {
    static let defaultPort = 8000
    static let ssePath = "/sse"
    static let overrideDefaultsKey = "mcpBaseURL"

    /// The simulator shares the Mac's loopback. A physical phone does not.
    static var allowsLoopbackHere: Bool {
        #if targetEnvironment(simulator)
        true
        #else
        false
        #endif
    }

    enum Decision: Equatable {
        case connect(URL)
        case blocked(String)
    }

    /// An explicit URL wins. Otherwise the selected NODE host is used on
    /// port 8000. Loopback is refused when `allowsLoopback` is false.
    static func decide(override: String?, nodeHost: String?, allowsLoopback: Bool) -> Decision {
        let typed = override?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
        if !typed.isEmpty {
            guard let url = URL(string: typed),
                  url.scheme == "http" || url.scheme == "https",
                  let host = url.host, !host.isEmpty else {
                return .blocked("MCP URL needs an http or https host, for example http://mac.local:8000/sse.")
            }
            if isLoopback(host), !allowsLoopback {
                return .blocked("That MCP URL points at this phone. Use the Mac's address, for example http://\(suggestedHost(nodeHost)):8000/sse.")
            }
            return .connect(url)
        }
        guard let nodeHost, !nodeHost.isEmpty else {
            return .blocked("MCP stays off until a NODE host is selected. A localhost URL would dial this phone.")
        }
        if isLoopback(nodeHost), !allowsLoopback {
            return .blocked("The selected NODE host is loopback. On a phone that is the phone itself. Add the Mac by its LAN address.")
        }
        guard let url = URL(string: "http://\(nodeHost):\(defaultPort)\(ssePath)") else {
            return .blocked("Could not build an MCP URL for \(nodeHost).")
        }
        return .connect(url)
    }

    /// Joins an SSE `endpoint` event to the configured server.
    ///
    /// A relative path stays on that server. An absolute loopback URL is
    /// rewritten onto the configured host, because FastMCP often advertises
    /// `http://127.0.0.1:8000/messages?...` even when reached from the LAN.
    static func messageURL(base: URL, endpoint: String) -> URL? {
        let trimmed = endpoint.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return nil }
        if trimmed.hasPrefix("http://") || trimmed.hasPrefix("https://") {
            guard var absolute = URLComponents(string: trimmed) else { return nil }
            if let host = absolute.host, isLoopback(host),
               let baseHost = base.host, !isLoopback(baseHost) {
                absolute.host = baseHost
            }
            return absolute.url
        }
        if trimmed.hasPrefix("/") {
            guard var parts = URLComponents(url: base, resolvingAgainstBaseURL: false),
                  let extra = URLComponents(string: trimmed) else { return nil }
            parts.path = extra.path
            parts.query = extra.query
            parts.fragment = extra.fragment
            return parts.url
        }
        return URL(string: trimmed, relativeTo: base)?.absoluteURL
    }

    static func isLoopback(_ host: String) -> Bool {
        var name = host.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        if name.hasPrefix("[") && name.hasSuffix("]") {
            name = String(name.dropFirst().dropLast())
        }
        if name == "localhost" || name == "::1" || name.hasPrefix("127.") {
            return true
        }
        return false
    }

    private static func suggestedHost(_ nodeHost: String?) -> String {
        guard let nodeHost, !nodeHost.isEmpty, !isLoopback(nodeHost) else { return "<mac>" }
        return nodeHost
    }
}
