import Foundation

class MCPClient {
    var isConnected: Bool = false
    
    private var messageEndpoint: URL?
    private var messageId = 1
    private var pendingRequests: [Int: (String) -> Void] = [:]
    
    func connect(onConnect: @escaping () -> Void) {
        guard let url = URL(string: "http://127.0.0.1:8000/sse") else { return }
        
        Task {
            do {
                let (asyncBytes, _) = try await URLSession.shared.bytes(from: url)
                
                for try await line in asyncBytes.lines {
                    if line.hasPrefix("data: ") {
                        let dataString = String(line.dropFirst(6))
                        
                        // If it's the initial endpoint URL
                        if dataString.contains("/messages?") {
                            if dataString.hasPrefix("http") {
                                self.messageEndpoint = URL(string: dataString)
                            } else {
                                self.messageEndpoint = URL(string: "http://127.0.0.1:8000\(dataString)")
                            }
                            
                            try await self.initializeConnection()
                            
                            DispatchQueue.main.async {
                                self.isConnected = true
                                onConnect()
                            }
                        } 
                        // If it's a JSON-RPC response
                        else if let data = dataString.data(using: .utf8),
                                  let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any] {
                            
                            if let id = json["id"] as? Int, let handler = pendingRequests[id] {
                                if let result = json["result"] as? [String: Any],
                                   let content = result["content"] as? [[String: Any]],
                                   let first = content.first,
                                   let text = first["text"] as? String {
                                    handler(text)
                                } else if let error = json["error"] as? [String: Any] {
                                    handler("Error: \(error)")
                                } else {
                                    // Handle initialize response
                                    handler("success")
                                }
                                pendingRequests.removeValue(forKey: id)
                            }
                        }
                    }
                }
            } catch {
                print("MCP Connection error: \(error)")
                DispatchQueue.main.async {
                    self.isConnected = false
                }
            }
        }
    }
    
    private func initializeConnection() async throws {
        let params: [String: Any] = [
            "protocolVersion": "2024-11-05",
            "capabilities": [String: Any](),
            "clientInfo": ["name": "bvr-ios", "version": "1.0"]
        ]
        
        try await sendRequest(method: "initialize", params: params) { _ in
            Task {
                try? await self.sendNotification(method: "notifications/initialized")
            }
        }
    }
    
    func callTool(name: String, args: [String: Any], completion: @escaping (String) -> Void) {
        Task {
            let params: [String: Any] = [
                "name": name,
                "arguments": args
            ]
            try? await sendRequest(method: "tools/call", params: params, completion: completion)
        }
    }
    
    private func sendRequest(method: String, params: [String: Any], completion: @escaping (String) -> Void) async throws {
        guard let endpoint = messageEndpoint else { return }
        
        let id = messageId
        messageId += 1
        
        pendingRequests[id] = completion
        
        let payload: [String: Any] = [
            "jsonrpc": "2.0",
            "id": id,
            "method": method,
            "params": params
        ]
        
        var request = URLRequest(url: endpoint)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONSerialization.data(withJSONObject: payload)
        
        let _ = try await URLSession.shared.data(for: request)
    }
    
    private func sendNotification(method: String) async throws {
        guard let endpoint = messageEndpoint else { return }
        
        let payload: [String: Any] = [
            "jsonrpc": "2.0",
            "method": method
        ]
        
        var request = URLRequest(url: endpoint)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONSerialization.data(withJSONObject: payload)
        
        let _ = try await URLSession.shared.data(for: request)
    }
}
