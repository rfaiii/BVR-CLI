import SwiftUI

struct ConnectionRow: View {
    @EnvironmentObject var connectionStore: ConnectionStore

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("Host")
                    .foregroundStyle(Theme.textSecondary)
                Spacer()
                Text(connectionStore.hostURL?.host ?? "Not connected")
                    .foregroundStyle(Theme.textPrimary)
                    .monospacedDigit()
            }

            HStack {
                Text("Status")
                    .foregroundStyle(Theme.textSecondary)
                Spacer()
                StatusBadge(status: connectionStore.status)
            }

            Button(action: connectionStore.toggleConnection) {
                Text(connectionStore.status == .connected ? "Disconnect" : "Connect")
                    .frame(maxWidth: .infinity)
            }
            .buttonStyle(.borderedProminent)
            .tint(Theme.primary)
            .disabled(!connectionStore.canConnect)
        }
        .padding()
        .background(Theme.surface, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
    }
}
