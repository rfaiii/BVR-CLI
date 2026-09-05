import Foundation
import Combine

enum SessionStatus {
    case idle
    case running
    case completed

    var label: String {
        switch self {
        case .idle: return "Idle"
        case .running: return "Running"
        case .completed: return "Completed"
        }
    }

    var systemImage: String {
        switch self {
        case .idle: return "pause.circle.fill"
        case .running: return "play.circle.fill"
        case .completed: return "checkmark.circle.fill"
        }
    }
}

struct Session: Identifiable {
    let id = UUID()
    var title: String
    var status: SessionStatus
    var updatedAt = Date()
}

final class SessionStore: ObservableObject {
    @Published var sessions: [Session] = [
        .init(title: "Refactor auth flow", status: .running),
        .init(title: "Review mobile bridge spec", status: .idle),
        .init(title: "Generate release notes", status: .completed)
    ]
}
