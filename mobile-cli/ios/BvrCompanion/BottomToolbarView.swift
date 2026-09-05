import SwiftUI

struct BottomToolbarView: View {
    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            // Command indicator
            VStack(alignment: .leading, spacing: 4) {
                HStack(spacing: 4) {
                    Image(systemName: "chevron.right")
                        .foregroundColor(Theme.textSecondary)
                        .font(.system(size: 12))
                    Text("Ready?")
                        .foregroundColor(Theme.textSecondary)
                }
                Image(systemName: "circle.grid.3x3.fill")
                    .foregroundColor(Theme.textSecondary)
                    .font(.system(size: 14))
            }
            .padding(.horizontal)
            
            // Open File Finder
            Button(action: {}) {
                HStack {
                    Image(systemName: "folder")
                    Text("OPEN FILE FINDER")
                }
                .font(.system(size: 14, weight: .medium, design: .monospaced))
                .foregroundColor(Theme.primary)
                .frame(maxWidth: .infinity)
                .padding(.vertical, 12)
                .background(Theme.surface)
                .cornerRadius(8)
                .overlay(
                    RoundedRectangle(cornerRadius: 8)
                        .stroke(Theme.primary, lineWidth: 1)
                )
            }
            .padding(.horizontal)
            
            // Bottom buttons row
            HStack(spacing: 8) {
                BottomButton(title: "OPEN\nMODEL")
                BottomButton(title: "SWITCH\nMODEL")
                BottomButton(title: "REFRESH")
                BottomButton(title: "MORE\nOPTIONS")
            }
            .padding(.horizontal)
        }
        .padding(.bottom, 20)
    }
}

struct BottomButton: View {
    var title: String
    
    var body: some View {
        Button(action: {}) {
            Text(title)
                .font(.system(size: 10, design: .monospaced))
                .multilineTextAlignment(.center)
                .foregroundColor(Theme.primary)
                .frame(maxWidth: .infinity)
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
