import Foundation

enum ConnectionStatus {
    case disconnected
    case connecting
    case connected
    case failed

    var label: String {
        switch self {
        case .disconnected: return "Disconnected"
        case .connecting:   return "Connecting"
        case .connected:    return "Connected"
        case .failed:       return "Failed"
        }
    }
}
