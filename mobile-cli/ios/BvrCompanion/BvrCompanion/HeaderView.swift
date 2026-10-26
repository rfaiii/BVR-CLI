import SwiftUI

struct HeaderView: View {
    @Environment(MockStore.self) private var store
    @State private var showNodePicker = false

    var body: some View {
        VStack(spacing: 8) {
            // Badges matching web prototype
            HStack(spacing: 12) {
                Button {
                    showNodePicker = true
                } label: {
                    BadgeView(icon: "network",
                              text: store.nodeName,
                              tint: nodeColor)
                }
                .buttonStyle(.plain)
                BadgeView(icon: "powerplug.fill", text: "fastmcp")
            }
            .padding(.top, 10)
            
            // Main Banner
            Text("//// BVR-CLI ////")
                .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 32))
                .foregroundColor(Theme.primary)
                
            // Glitch sub-banner
            GlitchText(text: "AWAITING NEURAL LINK...")
                .padding(.bottom, 12)
        }
        .padding(.horizontal)
        .background(Theme.background.opacity(0.8)) // Optional blur/glass effect base
        .sheet(isPresented: $showNodePicker) {
            NodePickerView()
                .environment(store)
                .presentationDetents([.medium, .large])
        }
    }

    /// Grey until the agent answers the handshake ping, green on the link, red
    /// when a host was reached and refused or never replied.
    private var nodeColor: Color {
        switch store.node.state {
        case .online: return Theme.success
        case .connecting: return Theme.folderYellow
        case .failed: return Theme.error
        case .offline: return Theme.textSecondary
        }
    }
}

struct BadgeView: View {
    let icon: String
    let text: String
    var tint: Color = Theme.primary
    
    var body: some View {
        HStack(spacing: 4) {
            Image(systemName: icon)
                .font(.system(size: 10, weight: .bold))
                .foregroundColor(tint)
            Text(text)
                .font(.system(size: 10, weight: .bold))
        }
        .padding(.horizontal, 8)
        .padding(.vertical, 4)
        .background(Theme.buttonSurface)
        .foregroundColor(tint)
        .cornerRadius(12)
        .overlay(
            RoundedRectangle(cornerRadius: 12)
                .stroke(tint.opacity(0.3), lineWidth: 1)
        )
    }
}
