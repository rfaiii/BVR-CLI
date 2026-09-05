import SwiftUI

struct HeaderView: View {
    var body: some View {
        VStack(spacing: 8) {
            // Badges matching web prototype
            HStack(spacing: 12) {
                BadgeView(icon: "network", text: "Local Device")
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
    }
}

struct BadgeView: View {
    let icon: String
    let text: String
    
    var body: some View {
        HStack(spacing: 4) {
            Image(systemName: icon)
                .font(.system(size: 10, weight: .bold))
            Text(text)
                .font(.system(size: 10, weight: .bold))
        }
        .padding(.horizontal, 8)
        .padding(.vertical, 4)
        .background(Theme.buttonSurface)
        .foregroundColor(Theme.primary)
        .cornerRadius(12)
        .overlay(
            RoundedRectangle(cornerRadius: 12)
                .stroke(Theme.primary.opacity(0.3), lineWidth: 1)
        )
    }
}
