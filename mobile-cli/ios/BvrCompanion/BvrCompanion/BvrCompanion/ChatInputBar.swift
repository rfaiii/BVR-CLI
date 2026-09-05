import SwiftUI

struct ChatInputBar: View {
    @Environment(MockStore.self) private var store
    @State private var text: String = ""
    @State private var showPalette: Bool = false
    
    var body: some View {
        VStack(spacing: 0) {
            CommandPaletteView(text: $text, isVisible: $showPalette)
            
            HStack(spacing: 12) {
                // Upload button
                Button(action: {}) {
                    Image(systemName: "plus")
                        .foregroundColor(Theme.primary)
                        .font(.system(size: 20))
                }
                
                // Text field
                TextField("Message BVR...", text: $text)
                    .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 14))
                    .foregroundColor(Theme.textPrimary)
                    .padding(.vertical, 10)
                    .padding(.horizontal, 12)
                    .background(Theme.surface)
                    .cornerRadius(20)
                    .overlay(
                        RoundedRectangle(cornerRadius: 20)
                            .stroke(Theme.border, lineWidth: 1)
                    )
                    .onChange(of: text) { _, newValue in
                        if newValue.hasPrefix("/") {
                            showPalette = true
                        } else {
                            showPalette = false
                        }
                    }
                    .onSubmit {
                        guard !text.isEmpty else { return }
                        if text.hasPrefix("/") {
                            let cmd = String(text.dropFirst()).trimmingCharacters(in: .whitespacesAndNewlines)
                            store.executeCommand(cmd)
                        } else {
                            store.sendMessage(text)
                        }
                        text = ""
                        showPalette = false
                    }
                
                // Share button
                Button(action: {}) {
                    Image(systemName: "square.and.arrow.up")
                        .foregroundColor(Theme.primary)
                        .font(.system(size: 20))
                }
            }
            .padding(.horizontal)
            .padding(.vertical, 8)
        }
    }
}
