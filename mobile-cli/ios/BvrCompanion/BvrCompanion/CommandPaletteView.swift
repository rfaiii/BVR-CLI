import SwiftUI

struct AppCommand: Identifiable {
    let id = UUID()
    let command: String
    let description: String
    let icon: String
}

struct CommandPaletteView: View {
    @Environment(MockStore.self) private var store
    @Binding var text: String
    @Binding var isVisible: Bool
    
    var filteredCommands: [AppCommand] {
        if text == "/" {
            return store.availableCommands
        }
        let search = text.dropFirst().lowercased()
        return store.availableCommands.filter { $0.command.lowercased().contains(search) || $0.description.lowercased().contains(search) }
    }
    
    var body: some View {
        if isVisible && !filteredCommands.isEmpty {
            VStack(spacing: 0) {
                ScrollView {
                    VStack(alignment: .leading, spacing: 0) {
                        ForEach(filteredCommands) { cmd in
                            Button(action: {
                                store.executeCommand(cmd.command)
                                text = ""
                                isVisible = false
                            }) {
                                HStack(spacing: 12) {
                                    Image(systemName: cmd.icon)
                                        .foregroundColor(Theme.primary)
                                        .frame(width: 24)
                                    
                                    VStack(alignment: .leading, spacing: 2) {
                                        Text("/\(cmd.command)")
                                            .font(.custom("JetBrainsMonoNerdFontMono-Bold", size: 14))
                                            .foregroundColor(Theme.textPrimary)
                                        
                                        Text(cmd.description)
                                            .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 12))
                                            .foregroundColor(Theme.textSecondary)
                                    }
                                    Spacer()
                                }
                                .padding(.vertical, 10)
                                .padding(.horizontal, 16)
                                .contentShape(Rectangle())
                            }
                            .buttonStyle(PlainButtonStyle())
                            
                            if cmd.id != filteredCommands.last?.id {
                                Divider()
                                    .background(Theme.border.opacity(0.5))
                            }
                        }
                    }
                }
                .frame(maxHeight: 250)
            }
            .background(Theme.surface)
            .cornerRadius(12)
            .overlay(
                RoundedRectangle(cornerRadius: 12)
                    .stroke(Theme.border, lineWidth: 1)
            )
            .shadow(color: Color.black.opacity(0.2), radius: 10, y: -5)
            .padding(.horizontal)
            .padding(.bottom, 8)
        }
    }
}
