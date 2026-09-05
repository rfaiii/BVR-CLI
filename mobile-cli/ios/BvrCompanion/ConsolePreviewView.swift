import SwiftUI

struct ConsolePreviewView: View {
    let lines: [ConsoleLine]

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 6) {
                ForEach(lines) { line in
                    Text(line.text)
                        .font(.system(.body, design: .monospaced))
                        .foregroundStyle(color(for: line.kind))
                }
            }
            .padding()
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .background(Theme.background.ignoresSafeArea())
        .navigationTitle("Console")
        .navigationBarTitleDisplayMode(.inline)
    }

    private func color(for kind: ConsoleLine.Kind) -> Color {
        switch kind {
        case .stdout:
            Theme.textPrimary
        case .stderr:
            Theme.error
        case .system:
            Theme.secondary
        }
    }
}

struct ConsoleLine: Identifiable {
    let id = UUID()
    let text: String
    let kind: Kind

    enum Kind {
        case stdout, stderr, system
    }
}
