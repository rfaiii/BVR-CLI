import Foundation

/// Sent to by `NodeClient.onStateChange` from URLSession callbacks, so it is
/// read off the main actor along with the wire types.
nonisolated enum NodeConnectionState: String {
    case offline
    case connecting
    case online
    case failed
}

nonisolated enum NodeError: LocalizedError {
    case notConnected
    case unauthorized
    case timeout
    case hostClosed(String)

    var errorDescription: String? {
        switch self {
        case .notConnected: return "Not connected to a NODE host."
        case .unauthorized: return "The host rejected the token."
        case .timeout: return "The host did not answer."
        case .hostClosed(let reason): return reason
        }
    }
}

/// Wire types for the NODE frame protocol. Field names, including the
/// `working_dir` and `exit_code` spellings, have to match `internal/node`
/// (transport.go and websocket.go); `internal/node/ios_wire_test.go` pins them.
///
/// `nonisolated` keeps the Codable conformance off the main actor: the target
/// defaults every type to MainActor, but these are encoded and decoded inside
/// URLSession callbacks.
nonisolated struct NodeCommandRequest: Codable {
    var command: String
    var args: [String]?
    var workingDir: String?

    enum CodingKeys: String, CodingKey {
        case command, args
        case workingDir = "working_dir"
    }
}

nonisolated struct NodeCommandResult: Codable {
    var stdout: String?
    var stderr: String?
    var exitCode: Int?

    enum CodingKeys: String, CodingKey {
        case stdout, stderr
        case exitCode = "exit_code"
    }
}

nonisolated struct NodeWSFrame: Codable {
    var id: String
    var type: String
    var stream: String?
    var data: String?
    var request: NodeCommandRequest?
    var result: NodeCommandResult?
    var error: String?
}

/// Speaks the NODE frame protocol over one WebSocket connection.
///
/// Every reply carries the `id` of the frame it answers, so requests are matched
/// by id rather than assumed to arrive in order. The agent handles one frame at a
/// time (`NodeWebSocketServer.handle` blocks while it streams), so a second
/// request waits behind the first.
final class NodeClient: NSObject, URLSessionWebSocketDelegate {
    /// How long to wait for the handshake ping. A host that accepts the TCP
    /// connection but never answers would otherwise sit on "connecting".
    static let handshakeTimeout: TimeInterval = 10

    var onStateChange: ((NodeConnectionState) -> Void)?

    private(set) var state: NodeConnectionState = .offline

    /// Why the last handshake failed, so the UI can name the reason instead of
    /// repeating the connecting message.
    private(set) var lastError: (any Error)?

    private struct Pending {
        var onLine: ((String, String) -> Void)?
        var onComplete: (Result<NodeCommandResult, Error>) -> Void
        // Output arrives one line per frame and the closing `result` frame
        // carries only the exit code, so the text is collected here the way
        // Go's own WebSocketTransport.Execute does.
        var stdout = ""
        var stderr = ""
    }

    private var session: URLSession?
    private var task: URLSessionWebSocketTask?
    private var pending: [String: Pending] = [:]
    private var handshakeWatchdog: DispatchWorkItem?

    func connect(to url: URL, token: String) {
        disconnect()
        lastError = nil

        var request = URLRequest(url: url)
        if !token.isEmpty {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        // Handing URLSession the main queue keeps every delegate and completion
        // callback on the same actor as the UI reading this state.
        let session = URLSession(configuration: .default, delegate: self, delegateQueue: OperationQueue.main)
        let task = session.webSocketTask(with: request)
        self.session = session
        self.task = task
        setState(.connecting)
        task.resume()
        receiveNext()

        let deadline = DispatchWorkItem { [weak self] in
            guard let self, self.state != .online else { return }
            self.failAll(with: NodeError.timeout)
            self.setState(.failed)
        }
        handshakeWatchdog = deadline
        DispatchQueue.main.asyncAfter(deadline: .now() + Self.handshakeTimeout, execute: deadline)

        // A successful upgrade is not proof of a working link: the token could
        // still be wrong and the frame protocol could differ. `ping`/`pong` is
        // the cheapest thing that proves both ends speak NODE.
        ping { [weak self] result in
            guard let self else { return }
            switch result {
            case .success:
                self.setState(.online)
            case .failure(let error):
                // Every failure path calls failAll before changing state, so the
                // ping completion is where the real reason is still attached.
                self.lastError = error
                self.setState(.failed)
            }
        }
    }

    func disconnect() {
        handshakeWatchdog?.cancel()
        handshakeWatchdog = nil
        guard task != nil || session != nil else { return }
        task?.cancel(with: .goingAway, reason: nil)
        task = nil
        pending.removeAll()
        // The session retains its delegate, so it has to be torn down or this
        // object outlives every reference to it.
        session?.finishTasksAndInvalidate()
        session = nil
        setState(.offline)
    }

    func ping(_ completion: @escaping (Result<Void, Error>) -> Void) {
        send(NodeWSFrame(id: UUID().uuidString, type: "ping")) { result in
            completion(result.map { _ in })
        }
    }

    func execute(_ request: NodeCommandRequest,
                 onLine: ((String, String) -> Void)? = nil,
                 completion: @escaping (Result<NodeCommandResult, Error>) -> Void) {
        send(NodeWSFrame(id: UUID().uuidString, type: "execute", request: request),
             onLine: onLine,
             completion: completion)
    }

    // MARK: - Framing

    private func send(_ frame: NodeWSFrame,
                      onLine: ((String, String) -> Void)? = nil,
                      completion: @escaping (Result<NodeCommandResult, Error>) -> Void) {
        guard let task, state == .online || frame.type == "ping" else {
            completion(.failure(NodeError.notConnected))
            return
        }
        guard let data = try? JSONEncoder().encode(frame), let text = String(data: data, encoding: .utf8) else {
            completion(.failure(NodeError.hostClosed("Could not encode the NODE frame.")))
            return
        }

        pending[frame.id] = Pending(onLine: onLine, onComplete: completion)
        task.send(.string(text)) { [weak self] error in
            guard let self, let error else { return }
            self.fail(frame.id, with: NodeError.hostClosed(error.localizedDescription))
        }
    }

    private func receiveNext() {
        guard let task else { return }
        task.receive { [weak self] result in
            guard let self, self.task != nil else { return }
            switch result {
            case .failure(let error):
                self.failAll(with: self.failureReason(for: error))
            case .success(let message):
                self.handle(message)
                self.receiveNext()
            }
        }
    }

    private func handle(_ message: URLSessionWebSocketTask.Message) {
        let data: Data
        switch message {
        case .string(let text): data = Data(text.utf8)
        case .data(let raw): data = raw
        @unknown default: return
        }
        guard let frame = try? JSONDecoder().decode(NodeWSFrame.self, from: data) else { return }

        switch frame.type {
        case "stdout", "stderr":
            guard let payload = frame.data, var entry = pending[frame.id] else { return }
            if frame.type == "stdout" {
                entry.stdout += payload
            } else {
                entry.stderr += payload
            }
            pending[frame.id] = entry
            entry.onLine?(frame.type, payload)
        case "result":
            var value = frame.result ?? NodeCommandResult()
            if value.stdout == nil { value.stdout = pending[frame.id]?.stdout }
            if value.stderr == nil { value.stderr = pending[frame.id]?.stderr }
            finish(frame.id, with: .success(value), reason: frame.error)
        case "pong":
            finish(frame.id, with: .success(NodeCommandResult()), reason: nil)
        case "error":
            finish(frame.id, with: .failure(NodeError.hostClosed(frame.error ?? "The host rejected the request.")), reason: nil)
        default:
            break
        }
    }

    /// `reason` is set when a `result` frame also carried an error string: the
    /// output is real, but the command failed, so the caller gets both.
    private func finish(_ id: String, with result: Result<NodeCommandResult, Error>, reason: String?) {
        guard let entry = pending.removeValue(forKey: id) else { return }
        switch result {
        case .success(let value):
            if let reason, !reason.isEmpty {
                entry.onComplete(.failure(NodeError.hostClosed(reason)))
            } else {
                entry.onComplete(.success(value))
            }
        case .failure(let error):
            entry.onComplete(.failure(error))
        }
    }

    private func fail(_ id: String, with error: Error) {
        guard let entry = pending.removeValue(forKey: id) else { return }
        entry.onComplete(.failure(error))
    }

    private func failAll(with error: Error) {
        let waiting = pending
        pending.removeAll()
        waiting.values.forEach { $0.onComplete(.failure(error)) }
    }

    /// A rejected upgrade arrives as an opaque transport error, so the HTTP
    /// status left on the task is what separates "wrong token" from "host gone".
    private func failureReason(for error: Error) -> NodeError {
        if let http = task?.response as? HTTPURLResponse, http.statusCode == 401 {
            return .unauthorized
        }
        return .hostClosed(error.localizedDescription)
    }

    private func setState(_ next: NodeConnectionState) {
        guard state != next else { return }
        state = next
        // Any resolved state ends the handshake, successful or not.
        handshakeWatchdog?.cancel()
        handshakeWatchdog = nil
        onStateChange?(next)
    }

    // MARK: - URLSessionWebSocketDelegate

    func urlSession(_ session: URLSession, task: URLSessionTask, didCompleteWithError error: (any Error)?) {
        guard let error else { return }
        // A cancelled task is our own disconnect(), not a failure worth
        // reporting to whoever has a request outstanding.
        if (error as NSError).code == NSURLErrorCancelled {
            setState(.offline)
            return
        }
        failAll(with: failureReason(for: error))
        setState(.failed)
    }

    func urlSession(_ session: URLSession, webSocketTask: URLSessionWebSocketTask,
                    didCloseWith closeCode: URLSessionWebSocketTask.CloseCode, reason: Data?) {
        let why = closeCode == .unsupportedData || closeCode == .protocolError
            ? "The host does not speak the NODE protocol."
            : "The host closed the connection (\(closeCode.rawValue))."
        failAll(with: NodeError.hostClosed(why))
        setState(.offline)
    }
}
