import SwiftUI

struct ContentView: View {
    @State private var mockStore = MockStore()

    var body: some View {
        ZStack {
            Theme.background.ignoresSafeArea()
            
            VStack(spacing: 0) {
                // Top dashboard section
                VStack(spacing: 16) {
                    HeaderView()
                    StatsView()
                    TaskPickerView()
                    ModelListView()
                }
                .padding(.top, 10)
                .padding(.bottom, 8)
                
                // Chat area takes up remaining flexible space
                ChatView()
                
                // Bottom input and toolbar section
                VStack(spacing: 8) {
                    ChatInputBar()
                    BottomToolbarView()
                }
                .padding(.bottom, 8)
            }
        }
        .environment(mockStore)
        .preferredColorScheme(.dark)
    }
}

#Preview {
    ContentView()
}
