import SwiftUI

struct ContentView: View {
    @EnvironmentObject var connectionStore: ConnectionStore
    @EnvironmentObject var sessionStore: SessionStore

    var body: some View {
        NavigationStack {
            List {
                Section("Connection") {
                    ConnectionRow()
                }

                Section("Sessions") {
                    if sessionStore.sessions.isEmpty {
                        ContentUnavailableView(
                            "No Sessions",
                            systemImage: "terminal",
                            description: Text("Connect to a host to view sessions.")
                        )
                    } else {
                        ForEach(sessionStore.sessions) { session in
                            SessionRow(session: session)
                        }
                    }
                }
            }
            .navigationTitle("Bvr Companion")
            .scrollContentBackground(.hidden)
            .background(Color.background.ignoresSafeArea())
        }
    }
}
