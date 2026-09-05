import SwiftUI

struct HostPickerView: View {
    @Environment(\.dismiss) private var dismiss
    @EnvironmentObject var connectionStore: ConnectionStore

    @State private var hostText: String = "localhost"
    @State private var portText: String = "8080"
    @State private var useTLS: Bool = false

    var body: some View {
        NavigationStack {
            Form {
                Section("Host") {
                    TextField("Host", text: $hostText)
                        .keyboardType(.URL)
                        .autocorrectionDisabled(true)
                        .textInputAutocapitalization(.never)

                    TextField("Port", text: $portText)
                        .keyboardType(.numberPad)
                }

                Section {
                    Toggle("Use TLS", isOn: $useTLS)
                }

                Section {
                    Button("Save Host") {
                        var components = URLComponents()
                        components.scheme = useTLS ? "wss" : "ws"
                        components.host = hostText
                        components.port = portText
                        components.path = "/companion"
                        connectionStore.hostURL = components.url
                        dismiss()
                    }
                    .disabled(hostText.isEmpty || portText.isEmpty)
                }
            }
            .navigationTitle("Host")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel", role: .cancel) {
                        dismiss()
                    }
                }
            }
        }
    }
}
