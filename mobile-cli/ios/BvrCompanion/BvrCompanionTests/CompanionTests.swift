import XCTest
@testable import BvrCompanion

@MainActor
final class CompanionTests: XCTestCase {
    func testParseHostAndPort() {
        expectParse("mac.local", host: "mac.local", port: 7777)
        expectParse("mac.local:7799", host: "mac.local", port: 7799)
        expectParse("[::1]", host: "[::1]", port: 7777)
        expectParse("[fe80::1]:7799", host: "[fe80::1]", port: 7799)
        XCTAssertNil(NodeStore.parse("fe80::1"))
        XCTAssertNil(NodeStore.parse("mac.local:0"))
        XCTAssertNil(NodeStore.parse("mac.local:99999"))
        XCTAssertNil(NodeStore.parse(":7777"))
        XCTAssertNil(NodeStore.parse("[]"))
        XCTAssertNil(NodeStore.parse(""))
    }

    func testTokenIsRememberedPerHost() {
        let memory = InMemoryTokenStore()
        let store = NodeStore(tokens: memory)
        store.updateToken("typed-before-host")
        store.addManualHost("  10.0.0.1:7777  ")
        XCTAssertEqual(store.token, "typed-before-host")
        XCTAssertEqual(memory.token(for: "10.0.0.1:7777"), "typed-before-host")

        store.addManualHost("10.0.0.2:7778")
        XCTAssertEqual(store.token, "")
        store.updateToken("beta")
        XCTAssertEqual(memory.token(for: "10.0.0.2:7778"), "beta")
        XCTAssertEqual(memory.token(for: "10.0.0.1:7777"), "typed-before-host")

        store.addManualHost("10.0.0.1:7777")
        XCTAssertEqual(store.token, "typed-before-host")
        let other = store.hosts.first { $0.port == 7778 }
        store.select(other?.id)
        XCTAssertEqual(store.token, "beta")

        store.updateToken("")
        XCTAssertEqual(memory.token(for: "10.0.0.2:7778"), "")
    }

    func testKeychainRoundTripUsesAPrivateService() {
        let service = "com.richard.BvrCompanion.node-token.test.\(UUID().uuidString)"
        let store = KeychainTokenStore(service: service)
        let account = "10.1.1.1:7777"
        defer { _ = store.save("", for: account) }

        XCTAssertTrue(store.save("s3cret-token", for: account), "status \(store.lastStatus)")
        XCTAssertEqual(store.token(for: account), "s3cret-token")
        XCTAssertTrue(store.save("rotated", for: account), "status \(store.lastStatus)")
        XCTAssertEqual(store.token(for: account), "rotated")
        XCTAssertTrue(store.save("", for: account), "status \(store.lastStatus)")
        XCTAssertEqual(store.token(for: account), "")
    }

    func testOriginHeaderMatchesTheRequestHost() {
        let v4 = URL(string: "ws://10.0.0.8:7777/v1/node/ws")
        XCTAssertEqual(NodeClient.originHeader(for: v4!), "http://10.0.0.8:7777")
        let v4Parts = URLComponents(string: "http://10.0.0.8:7777")
        XCTAssertEqual(originHost(v4Parts), "10.0.0.8:7777")

        let v6 = URL(string: "ws://[fe80::1]:7777/v1/node/ws")
        XCTAssertEqual(NodeClient.originHeader(for: v6!), "http://[fe80::1]:7777")
        let v6Parts = URLComponents(string: "http://[fe80::1]:7777")
        XCTAssertEqual(originHost(v6Parts), "[fe80::1]:7777")

        let secure = URL(string: "wss://mac.local/v1/node/ws")
        XCTAssertEqual(NodeClient.originHeader(for: secure!), "https://mac.local")
    }

    func testMCPStaysOffWithoutAHostAndRejectsPhoneLoopback() {
        XCTAssertEqual(
            MCPEndpoint.decide(override: nil, nodeHost: nil, allowsLoopback: false),
            MCPEndpoint.Decision.blocked("MCP stays off until a NODE host is selected. A localhost URL would dial this phone.")
        )
        XCTAssertEqual(
            MCPEndpoint.decide(override: "   ", nodeHost: "mac.local", allowsLoopback: false),
            MCPEndpoint.Decision.connect(URL(string: "http://mac.local:8000/sse")!)
        )
        guard case .blocked = MCPEndpoint.decide(override: nil, nodeHost: "127.0.0.1", allowsLoopback: false) else {
            return XCTFail("loopback NODE host should be blocked on a phone")
        }
        XCTAssertEqual(
            MCPEndpoint.decide(override: nil, nodeHost: "127.0.0.1", allowsLoopback: true),
            MCPEndpoint.Decision.connect(URL(string: "http://127.0.0.1:8000/sse")!)
        )
        XCTAssertEqual(
            MCPEndpoint.decide(override: "https://mcp.example:9000/sse", nodeHost: "mac.local", allowsLoopback: false),
            MCPEndpoint.Decision.connect(URL(string: "https://mcp.example:9000/sse")!)
        )
        guard case .blocked = MCPEndpoint.decide(override: "http://localhost:8000/sse", nodeHost: "mac.local", allowsLoopback: false) else {
            return XCTFail("explicit loopback should be blocked on a phone")
        }
        guard case .blocked = MCPEndpoint.decide(override: "not a url", nodeHost: "mac.local", allowsLoopback: true) else {
            return XCTFail("a URL with no host should be blocked")
        }
    }

    func testMCPMessageURLStaysOnTheConfiguredHost() {
        let base = URL(string: "http://mac.local:8000/sse")!
        XCTAssertEqual(
            MCPEndpoint.messageURL(base: base, endpoint: "/messages/?session_id=abc")?.absoluteString,
            "http://mac.local:8000/messages/?session_id=abc"
        )
        XCTAssertEqual(
            MCPEndpoint.messageURL(base: base, endpoint: "http://127.0.0.1:8000/messages/?session_id=abc")?.absoluteString,
            "http://mac.local:8000/messages/?session_id=abc"
        )
        let loopback = URL(string: "http://127.0.0.1:8000/sse")!
        XCTAssertEqual(
            MCPEndpoint.messageURL(base: loopback, endpoint: "http://127.0.0.1:8000/messages/?session_id=abc")?.absoluteString,
            "http://127.0.0.1:8000/messages/?session_id=abc"
        )
    }

    private func expectParse(_ input: String, host: String, port: Int, file: StaticString = #filePath, line: UInt = #line) {
        let value = NodeStore.parse(input)
        XCTAssertEqual(value?.host, host, file: file, line: line)
        XCTAssertEqual(value?.port, port, file: file, line: line)
    }

    /// Go's `url.URL.Host`, which is what the NODE origin check compares.
    private func originHost(_ parts: URLComponents?) -> String? {
        guard var host = parts?.host, !host.isEmpty else { return nil }
        if host.hasPrefix("[") && host.hasSuffix("]") {
            host = String(host.dropFirst().dropLast())
        }
        let name = host.contains(":") ? "[\(host)]" : host
        if let port = parts?.port {
            return "\(name):\(port)"
        }
        return name
    }
}
