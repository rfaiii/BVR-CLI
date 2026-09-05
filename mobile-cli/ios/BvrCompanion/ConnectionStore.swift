import SwiftUI
import Combine

enum ConnectionStatus: String {
    case disconnected
    case connecting
    case connected
    case failed

    var label: String {
        rawValue.capitalized
    }
}

@Observable
final class ConnectionStore {
    var hostURL: URL?
    var status: ConnectionStatus = .disconnected
    var canConnect: Bool = true
    var lastError: String?

    private var webSocketTask: URLSessionWebSocketTask?
    private var session = URLSession.shared
    private var cancellables = Set<AnyCancellable>()

    func toggleConnection() {
        switch status {
        case .disconnected, .failed:
            connect()
        case .connecting, .connected:
            disconnect()
        }
    }

    func connect() {
        status = .connecting
        lastError = nil

        guard let url = URL(string: "ws://localhost:8080/companion") else {
            status = .failed
            lastError = "Invalid host URL"
            return
        }

        hostURL = url
        webSocketTask = session.webSocketTask(with: url)
        webSocketTask?.resume()
        receive()
        status = .connected
    }

    func disconnect() {
        webSocketTask?.cancel(with: .goingAway, reason: nil)
        webSocketTask = nil
        status = .disconnected
    }

    private func receive() {
        webSocketTask?.receive { [weak self] result in
            switch result {
            case .success:
                self?.receive()
            case .failure:
                self?.status = .failed
                self?.lastError = "Connection lost"
            }
        }
    }
}
