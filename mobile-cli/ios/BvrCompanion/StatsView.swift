import SwiftUI

struct StatsView: View {
    @Environment(MockStore.self) private var store
    
    var body: some View {
        VStack(spacing: 8) {
            HStack {
                Text("NODE:").foregroundColor(Theme.textSecondary) + Text(" \(store.nodeName)").foregroundColor(Theme.textPrimary)
                Spacer()
                Text("CPU:").foregroundColor(Theme.textSecondary)
                ProgressBar(value: store.cpuUsage)
                    .frame(width: 60, height: 10)
                Text("\(Int(store.cpuUsage * 100))%").foregroundColor(Theme.primary).font(.caption.monospaced())
            }
            .font(.caption)
            
            HStack {
                Text("MCPs:").foregroundColor(Theme.textSecondary) + Text(" \(store.mcps)").foregroundColor(Theme.textPrimary)
                Spacer()
                Text("RAM:").foregroundColor(Theme.textSecondary)
                ProgressBar(value: store.ramUsage)
                    .frame(width: 60, height: 10)
                Text("\(Int(store.ramUsage * 100))%").foregroundColor(Theme.primary).font(.caption.monospaced())
            }
            .font(.caption)
            
            HStack {
                Circle()
                    .fill(store.isReady ? Theme.primary : Theme.error)
                    .frame(width: 8, height: 8)
                Text("Ready?").foregroundColor(Theme.textSecondary).font(.caption)
                Spacer()
            }
        }
        .padding()
        .background(Theme.surface)
        .cornerRadius(8)
        .overlay(
            RoundedRectangle(cornerRadius: 8)
                .stroke(Theme.border, lineWidth: 1)
        )
        .padding(.horizontal)
    }
}

struct ProgressBar: View {
    var value: Double
    
    var body: some View {
        GeometryReader { geometry in
            ZStack(alignment: .leading) {
                Rectangle().frame(width: geometry.size.width, height: geometry.size.height)
                    .opacity(0.3)
                    .foregroundColor(Theme.border)
                
                Rectangle().frame(width: min(CGFloat(self.value)*geometry.size.width, geometry.size.width), height: geometry.size.height)
                    .foregroundColor(Theme.primary)
            }
        }
    }
}
