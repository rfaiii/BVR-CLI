import SwiftUI

struct SessionRow: View {
    let session: Session

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(session.title)
                .foregroundStyle(Theme.textPrimary)
                .font(.headline)

            HStack(spacing: 12) {
                Label(session.status.label, systemImage: session.status.systemImage)
                    .foregroundStyle(Theme.textSecondary)
                    .font(.caption)
                Spacer()
                Text(session.updatedAt, style: .relative)
                    .foregroundStyle(Theme.textSecondary)
                    .font(.caption)
            }
        }
        .padding()
        .background(Theme.surface, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 12, style: .continuous).stroke(Theme.border, lineWidth: 1))
    }
}
