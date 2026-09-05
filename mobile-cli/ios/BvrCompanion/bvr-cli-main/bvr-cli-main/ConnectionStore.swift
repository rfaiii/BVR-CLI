import Foundation
import Combine


final class ConnectionStore: ObservableObject {
    @Published var hostURL: URL?
    @Published var status: ConnectionStatus = .disconnected
    @Published var canConnect: Bool = true
    @Published var lastError: String?

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

        guard let url = hostURL ?? URL(string: "ws://localhost:8080/companion") else {
            status = .failed
            lastError = "Invalid host URL"
            return
        }
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
            DispatchQueue.main.async {
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
}
