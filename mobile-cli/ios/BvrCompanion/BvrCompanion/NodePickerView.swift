import SwiftUI

/// Sheet for choosing a NODE host, authenticating, and proving the link.
struct NodePickerView: View {
    @Environment(MockStore.self) private var store
    @Environment(\.dismiss) private var dismiss

    @State private var address = ""

    private var node: NodeStore { store.node }

    var body: some View {
        NavigationStack {
            Form {
                Section("On the local network") {
                    if node.hosts.isEmpty {
                        Text(node.isScanning
                             ? "Listening for _bvr._tcp advertisements…"
                             : "Not scanning. Start a browse to see advertised hosts.")
                            .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 12))
                            .foregroundColor(Theme.textSecondary)
                    }
                    ForEach(node.hosts) { host in
                        hostRow(host)
                    }
                    Button(node.isScanning ? "Stop scanning" : "Scan for hosts") {
                        if node.isScanning {
                            node.stopScanning()
                        } else {
                            node.startScanning()
                        }
                    }
                    .foregroundColor(Theme.primary)
                }

                Section("Add by address") {
                    TextField("host or host:port", text: $address)
                        .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 12))
                        .keyboardType(.URL)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled()
                        .submitLabel(.done)
                        .accessibilityIdentifier("node-address")
                        .onSubmit { node.addManualHost(address) }
                    Button("Add") {
                        node.addManualHost(address)
                        if node.selectedHost != nil { address = "" }
                    }
                    .foregroundColor(Theme.primary)
                    .accessibilityIdentifier("node-add")
                }

                Section("Connection") {
                    SecureField("agent bearer token", text: bindToken)
                        .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 12))
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled()
                        .accessibilityIdentifier("node-token")

                    HStack(spacing: 8) {
                        Circle()
                            .fill(stateColor)
                            .frame(width: 8, height: 8)
                        Text(node.state.rawValue.uppercased())
                            .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 11))
                            .foregroundColor(Theme.textPrimary)
                            .accessibilityIdentifier("node-state")
                        Spacer()
                    }
                    if node.isConnected {
                        // `.borderless` keeps these from becoming full-width
                        // form rows that overlap and both receive one tap.
                        HStack {
                            Button("Disconnect") { node.disconnect() }
                                .buttonStyle(.borderless)
                                .foregroundColor(Theme.error)
                                .accessibilityIdentifier("node-disconnect")
                            Spacer()
                            Button("Test `uname`") { node.runProbe() }
                                .buttonStyle(.borderless)
                                .foregroundColor(Theme.primary)
                                .accessibilityIdentifier("node-test")
                        }
                    } else {
                        Button("Connect") { node.connect() }
                            .buttonStyle(.borderless)
                            .disabled(node.selectedHost == nil)
                            .foregroundColor(Theme.primary)
                            .accessibilityIdentifier("node-connect")
                    }

                    if let status = node.statusMessage {
                        Text(status)
                            .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 11))
                            .foregroundColor(Theme.error)
                    }
                    if !node.probeOutput.isEmpty {
                        Text(node.probeOutput)
                            .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 11))
                            .foregroundColor(Theme.textSecondary)
                            .accessibilityIdentifier("node-probe")
                    }
                }

                Section {
                    TextField("http://host:8000/sse", text: bindMCP)
                        .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 12))
                        .keyboardType(.URL)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled()
                        .submitLabel(.done)
                        .accessibilityIdentifier("mcp-url")
                        .onSubmit { store.refreshMCP() }
                    Text(store.mcpStatus)
                        .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 11))
                        .foregroundColor(Theme.textSecondary)
                        .accessibilityIdentifier("mcp-status")
                    Button("Reconnect MCP") { store.refreshMCP() }
                        .foregroundColor(Theme.primary)
                        .accessibilityIdentifier("mcp-reconnect")
                } header: {
                    Text("MCP server")
                } footer: {
                    Text("Leave this empty to use the selected NODE host on port 8000. The app will not dial 127.0.0.1 from a phone.")
                }
            }
            .scrollDismissesKeyboard(.interactively)
            .navigationTitle("NODE hosts")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("Done") { dismiss() }
                        .foregroundColor(Theme.primary)
                }
            }
        }
        .presentationDragIndicator(.visible)
    }

    private func hostRow(_ host: DiscoveredNode) -> some View {
        Button {
            node.select(host.id)
        } label: {
            HStack(spacing: 10) {
                VStack(alignment: .leading, spacing: 2) {
                    Text(host.displayName)
                        .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 13))
                        .foregroundColor(Theme.textPrimary)
                    Text("\(host.host):\(host.port)")
                        .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 11))
                        .foregroundColor(Theme.textSecondary)
                }
                Spacer()
                if !host.isProtocolCompatible {
                    Text("proto \(host.proto)")
                        .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 10))
                        .foregroundColor(Theme.error)
                }
                if node.selectedID == host.id {
                    Image(systemName: "checkmark.circle.fill")
                        .foregroundColor(Theme.primary)
                }
            }
        }
        .buttonStyle(.plain)
    }

    private var bindToken: Binding<String> {
        Binding(get: { node.token }, set: { node.updateToken($0) })
    }

    private var bindMCP: Binding<String> {
        Binding(get: { store.mcpOverride }, set: { store.mcpOverride = $0 })
    }

    private var stateColor: Color {
        switch node.state {
        case .online: return Theme.success
        case .connecting: return Theme.folderYellow
        case .failed: return Theme.error
        case .offline: return Theme.textSecondary
        }
    }
}
