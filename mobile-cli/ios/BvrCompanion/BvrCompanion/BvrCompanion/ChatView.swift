import SwiftUI

struct ChatView: View {
    @Environment(MockStore.self) private var store
    
    var body: some View {
        ScrollViewReader { proxy in
            ScrollView {
                VStack(spacing: 16) {
                    ForEach(store.messages) { message in
                        HStack {
                            if message.isUser {
                                Spacer()
                                Text(message.text)
                                    .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 14))
                                    .padding()
                                    .background(Theme.surface)
                                    .foregroundColor(Theme.textPrimary)
                                    .cornerRadius(16)
                                    .overlay(
                                        RoundedRectangle(cornerRadius: 16)
                                            .stroke(Theme.border, lineWidth: 1)
                                    )
                                    .padding(.leading, 40)
                            } else {
                                Text(message.text)
                                    .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 14))
                                    .padding()
                                    .background(Color.clear)
                                    .foregroundColor(Theme.primary)
                                    .cornerRadius(16)
                                    .overlay(
                                        RoundedRectangle(cornerRadius: 16)
                                            .stroke(Theme.primary.opacity(0.3), lineWidth: 1)
                                    )
                                    .padding(.trailing, 40)
                                Spacer()
                            }
                        }
                        .id(message.id)
                    }
                }
                .padding(.horizontal)
                .padding(.vertical, 8)
            }
            .onChange(of: store.messages.count) {
                if let last = store.messages.last {
                    withAnimation {
                        proxy.scrollTo(last.id, anchor: .bottom)
                    }
                }
            }
        }
    }
}
