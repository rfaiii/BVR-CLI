import Foundation

/// Minimal FastMCP SSE client.
///
/// The base URL comes from `MCPEndpoint`. This type never invents
/// `127.0.0.1`: the caller passes the URL, and message endpoints that the
/// server describes as loopback are rewritten onto that URL's host.
final class MCPClient {
    private(set) var isConnected = false
    var onStatus: ((String) -> Void)?

    private var messageEndpoint: URL?
    private var baseURL: URL?
    private var messageId = 1
    private var pendingRequests: [Int: (String) -> Void] = [:]
    private var task: Task<Void, Never>?
    private var generation = 0

    func disconnect() {
        generation += 1
        task?.cancel()
        task = nil
        messageEndpoint = nil
        isConnected = false
    }

    func connect(to url: URL, onConnect: @escaping () -> Void) {
        disconnect()
        let generation = generation
        baseURL = url
        task = Task { [weak self] in
            guard let self else { return }
            do {
                let (bytes, _) = try await URLSession.shared.bytes(from: url)
                for try await line in bytes.lines {
                    if Task.isCancelled || generation != self.generation { return }
                    guard line.hasPrefix("data: ") else { continue }
                    let dataString = String(line.dropFirst(6))
                    if dataString.contains("/messages?") {
                        guard let endpoint = MCPEndpoint.messageURL(base: url, endpoint: dataString) else {
                            self.fail("MCP server sent an unusable message endpoint.", generation: generation)
                            return
                        }
                        self.messageEndpoint = endpoint
                        try await self.initializeConnection()
                        guard generation == self.generation else { return }
                        self.isConnected = true
                        onConnect()
                    } else if let data = dataString.data(using: .utf8),
                              let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any],
                              let id = json["id"] as? Int,
                              let handler = self.pendingRequests[id] {
                        handler(Self.text(from: json))
                        self.pendingRequests.removeValue(forKey: id)
                    }
                }
                if generation == self.generation, self.isConnected {
                    self.isConnected = false
                    self.onStatus?("MCP server closed the connection.")
                }
            } catch {
                if Task.isCancelled || generation != self.generation { return }
                self.fail("MCP connection failed: \(error.localizedDescription)", generation: generation)
            }
        }
    }

    func callTool(name: String, args: [String: Any], completion: @escaping (String) -> Void) {
        Task {
            let params: [String: Any] = [
                "name": name,
                "arguments": args,
            ]
            try? await sendRequest(method: "tools/call", params: params, completion: completion)
        }
    }

    private func fail(_ message: String, generation: Int) {
        guard generation == self.generation else { return }
        isConnected = false
        onStatus?(message)
    }

    private func initializeConnection() async throws {
        let params: [String: Any] = [
            "protocolVersion": "2024-11-05",
            "capabilities": [String: Any](),
            "clientInfo": ["name": "bvr-ios", "version": "1.0"],
        ]
        try await sendRequest(method: "initialize", params: params) { _ in
            Task {
                try? await self.sendNotification(method: "notifications/initialized")
            }
        }
    }

    private func sendRequest(method: String, params: [String: Any], completion: @escaping (String) -> Void) async throws {
        guard let endpoint = messageEndpoint else { return }
        let id = messageId
        messageId += 1
        pendingRequests[id] = completion
        try await post(endpoint, payload: [
            "jsonrpc": "2.0",
            "id": id,
            "method": method,
            "params": params,
        ])
    }

    private func sendNotification(method: String) async throws {
        guard let endpoint = messageEndpoint else { return }
        try await post(endpoint, payload: [
            "jsonrpc": "2.0",
            "method": method,
        ])
    }

    private func post(_ endpoint: URL, payload: [String: Any]) async throws {
        var request = URLRequest(url: endpoint)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONSerialization.data(withJSONObject: payload)
        _ = try await URLSession.shared.data(for: request)
    }

    private static func text(from json: [String: Any]) -> String {
        if let result = json["result"] as? [String: Any],
           let content = result["content"] as? [[String: Any]],
           let first = content.first,
           let text = first["text"] as? String {
            return text
        }
        if let error = json["error"] as? [String: Any] {
            return "Error: \(error)"
        }
        return "success"
    }
}
