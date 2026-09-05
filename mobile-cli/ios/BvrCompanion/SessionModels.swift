import SwiftUI

struct Session: Identifiable, Codable {
    let id: UUID
    let title: String
    let status: SessionStatus
    let updatedAt: Date

    init(
        id: UUID = .init(),
        title: String,
        status: SessionStatus = .idle,
        updatedAt: Date = .init()
    ) {
        self.id = id
        self.title = title
        self.status = status
        self.updatedAt = updatedAt
    }
}

enum SessionStatus: String, Codable {
    case idle
    case running
    case paused
    case completed
    case failed

    var label: String {
        rawValue.capitalized
    }

    var systemImage: String {
        switch self {
        case .idle:
            "circle.dotted"
        case .running:
            "sparkles"
        case .paused:
            "pause.circle"
        case .completed:
            "checkmark.circle"
        case .failed:
            "xmark.circle"
        }
    }
}

@Observable
final class SessionStore {
    var sessions: [Session] = [
        .init(title: "Refactor auth flow", status: .running),
        .init(title: "Review mobile bridge spec", status: .idle),
        .init(title: "Generate release notes", status: .completed)
    ]
}
