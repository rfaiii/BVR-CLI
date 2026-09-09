import SwiftUI

struct ContentView: View {
    @EnvironmentObject var connectionStore: ConnectionStore
    @EnvironmentObject var sessionStore: SessionStore
    @State private var messageText: String = ""

    var body: some View {
        NavigationStack {
            VStack(spacing: 0) {
                // Header (Node / MCP Status)
                HeaderView()
                
                // Model Banner
                ModelListView()
                    .frame(height: 60)
                
                // Chat Area
                ScrollView {
                    VStack(alignment: .leading, spacing: 12) {
                        Text("AWAITING NEURAL LINK...")
                            .font(.system(size: 14, weight: .bold, design: .monospaced))
                            .foregroundColor(Theme.primary)
                            .frame(maxWidth: .infinity, alignment: .center)
                            .padding(.vertical)
                            
                        // Placeholder messages
                        HStack {
                            Spacer()
                            Text("Can you help me parse this JSON file?")
                                .padding()
                                .background(Theme.surface)
                                .cornerRadius(12)
                                .foregroundColor(Theme.textPrimary)
                        }
                        
                        HStack {
                            Text("Sure! Let's take a look at the structure of your JSON file. You can use the `jq` tool or write a small Go script.")
                                .padding()
                                .background(Theme.surface)
                                .cornerRadius(12)
                                .foregroundColor(Theme.textPrimary)
                            Spacer()
                        }
                    }
                    .padding()
                }
                
                // Input Bar
                HStack {
                    Button(action: {}) {
                        Image(systemName: "plus")
                            .foregroundColor(Theme.textSecondary)
                    }
                    
                    TextField("Message BVR...", text: $messageText)
                        .padding(10)
                        .background(Theme.surface)
                        .cornerRadius(20)
                        .foregroundColor(Theme.textPrimary)
                        
                    Button(action: {}) {
                        Image(systemName: "arrow.up.circle.fill")
                            .foregroundColor(Theme.primary)
                            .font(.system(size: 24))
                    }
                }
                .padding()
                .background(Color.background)
            }
            .navigationTitle("BVR Companion")
            .navigationBarTitleDisplayMode(.inline)
            .background(Color.background.ignoresSafeArea())
        }
    }
}
