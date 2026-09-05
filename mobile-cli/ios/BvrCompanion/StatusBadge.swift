import SwiftUI

struct StatusBadge: View {
    let status: ConnectionStatus

    var body: some View {
        HStack(spacing: 6) {
            Circle()
                .fill(status == .connected ? Theme.success : Theme.error)
                .frame(width: 8, height: 8)
            Text(status.label)
                .foregroundStyle(status == .connected ? Theme.success : Theme.error)
        }
        .padding(.horizontal, 10)
        .padding(.vertical, 6)
        .background(Theme.surface, in: Capsule())
        .overlay(Capsule().stroke(Theme.border, lineWidth: 1))
    }
}
