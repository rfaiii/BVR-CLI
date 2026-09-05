import SwiftUI

struct ModelListView: View {
    @Environment(MockStore.self) private var store
    @State private var isExpanded: Bool = false
    
    var body: some View {
        VStack(spacing: 0) {
            // Header for the selected model
            Button(action: {
                withAnimation(.easeInOut(duration: 0.2)) {
                    isExpanded.toggle()
                }
            }) {
                HStack {
                    Text(store.currentModelName)
                        .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 14))
                        .foregroundColor(Theme.textPrimary)
                    Spacer()
                    Image(systemName: "chevron.right")
                        .rotationEffect(.degrees(isExpanded ? 90 : 0))
                        .foregroundColor(Theme.textSecondary)
                }
                .padding()
                .background(Theme.surface)
            }
            
            if isExpanded {
                Divider().background(Theme.border)
                
                // List of models
                VStack(spacing: 0) {
                    ForEach(store.models) { model in
                        HStack {
                            Text(model.name)
                                .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 14))
                                .foregroundColor(Theme.textSecondary)
                            if let subtitle = model.subtitle {
                                Text(subtitle)
                                    .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 14))
                                    .foregroundColor(Theme.textSecondary)
                            }
                            Spacer()
                            Image(systemName: model.status.iconName)
                                .foregroundColor(model.status.iconColor)
                                .font(.system(size: 14))
                        }
                        .padding(.horizontal)
                        .padding(.vertical, 10)
                        
                        if model.id != store.models.last?.id {
                            Divider().background(Theme.border).padding(.leading)
                        }
                    }
                }
                .background(Theme.surface)
                
                Divider().background(Theme.border)
                
                // Plugin models at the bottom
                VStack(spacing: 0) {
                    ForEach(store.pluginModels) { model in
                        HStack {
                            Image(systemName: "folder")
                                .foregroundColor(Theme.primary)
                                .font(.system(size: 14))
                            Text(model.name)
                                .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 14))
                                .foregroundColor(Theme.textSecondary)
                            Spacer()
                            Image(systemName: model.status.iconName)
                                .foregroundColor(model.status.iconColor)
                                .font(.system(size: 10))
                        }
                        .padding(.horizontal)
                        .padding(.vertical, 10)
                    }
                }
                .background(Theme.surface)
            }
        }
        .cornerRadius(8)
        .overlay(
            RoundedRectangle(cornerRadius: 8)
                .stroke(Theme.border, lineWidth: 1)
        )
        .padding(.horizontal)
    }
}
