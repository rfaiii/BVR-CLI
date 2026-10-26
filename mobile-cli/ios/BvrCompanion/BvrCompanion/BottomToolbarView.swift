import SwiftUI

struct BottomToolbarView: View {
    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            // Command indicator
            VStack(alignment: .leading, spacing: 4) {
                HStack(spacing: 4) {
                    Image(systemName: "chevron.right")
                        .foregroundColor(Theme.textSecondary)
                        .font(.system(size: 12))
                    Text("Ready?")
                        .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 12))
                        .foregroundColor(Theme.textSecondary)
                }
                Image(systemName: "circle.grid.3x3.fill")
                    .foregroundColor(Theme.textSecondary)
                    .font(.system(size: 14))
            }
            .padding(.horizontal)
            
            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 8) {
                    BottomButton(title: "OPEN\nFILE FINDER", icon: "folder")
                    BottomButton(title: "OPEN\nMODEL")
                    BottomButton(title: "SWITCH\nMODEL")
                    BottomButton(title: "REFRESH")
                    BottomButton(title: "MORE\nOPTIONS")
                }
                .padding(.horizontal)
            }
        }
        .padding(.bottom, 10)
    }
}

struct BottomButton: View {
    var title: String
    var icon: String? = nil
    
    var body: some View {
        Button(action: {}) {
            HStack(spacing: 4) {
                if let icon = icon {
                    Image(systemName: icon)
                }
                Text(title)
                    .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 10))
                    .multilineTextAlignment(.center)
            }
            .foregroundColor(Theme.primary)
            .padding(.horizontal, 16)
            .frame(height: 44)
            .background(Theme.surface)
            .cornerRadius(8)
            .overlay(
                RoundedRectangle(cornerRadius: 8)
                    .stroke(Theme.border, lineWidth: 1)
            )
        }
    }
}
