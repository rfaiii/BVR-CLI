import SwiftUI

enum ModelStatus: String {
    case ready
    case folder
    case alert
    case smallDot
    
    var iconName: String {
        switch self {
        case .ready: return "checkmark.circle.fill"
        case .folder: return "folder.fill"
        case .alert: return "exclamationmark.circle"
        case .smallDot: return "circle.fill"
        }
    }
    
    var iconColor: Color {
        switch self {
        case .ready: return Theme.primary
        case .folder: return Theme.folderYellow
        case .alert: return Theme.textSecondary
        case .smallDot: return Theme.primary
        }
    }
}

struct AIModel: Identifiable {
    let id = UUID()
    var name: String
    let subtitle: String?
    var status: ModelStatus
}

struct ChatMessage: Identifiable {
    let id = UUID()
    let isUser: Bool
    let text: String
}

@Observable
class MockStore {
    let node = NodeStore()

    var mcps: String = "Off"
    var cpuUsage: Double = 1.0 // 100%
    var ramUsage: Double = 0.64 // 64%

    /// Real state now that the app browses `_bvr._tcp` itself.
    var nodeName: String { node.badgeLabel }
    var isReady: Bool { node.isConnected }
    var selectedTaskType: Int = 0 // 0 for Large, 1 for Small
    var currentModelName: String = "OpenAI: GPT-5.6 Luna Pro"
    
    var models: [AIModel] = [
        AIModel(name: "ornith:latest", subtitle: "(Ollama - Local)", status: .ready)
    ]
    
    var pluginModels: [AIModel] = [
        AIModel(name: "fastmcp: off", subtitle: nil, status: .alert)
    ]
    
    var messages: [ChatMessage] = []
    
    let mcpClient = MCPClient()
    var mcpOverride: String {
        didSet {
            UserDefaults.standard.set(mcpOverride, forKey: MCPEndpoint.overrideDefaultsKey)
        }
    }
    var mcpStatus: String
    var mcpConnected = false
    
    var availableCommands: [AppCommand] = [
        AppCommand(command: "themes", description: "Change application theme", icon: "paintbrush.fill"),
        AppCommand(command: "ollama", description: "Manage local Ollama models", icon: "cpu"),
        AppCommand(command: "node", description: "Manage desktop NODE connections", icon: "network"),
        AppCommand(command: "finder", description: "Open file browser", icon: "folder.fill"),
        AppCommand(command: "cd", description: "Change working directory", icon: "arrow.uturn.right"),
        AppCommand(command: "mcp", description: "Manage MCP servers", icon: "server.rack")
    ]
    
    init() {
        mcpOverride = UserDefaults.standard.string(forKey: MCPEndpoint.overrideDefaultsKey) ?? ""
        mcpStatus = "MCP stays off until a NODE host is selected. A localhost URL would dial this phone."
        mcpClient.onStatus = { [weak self] message in
            guard let self else { return }
            self.mcpStatus = message
            self.mcpConnected = false
            self.mcps = "Off"
            self.pluginModels = [
                AIModel(name: "fastmcp: off", subtitle: nil, status: .alert)
            ]
        }
        node.onSelectedHostChange = { [weak self] _ in
            self?.refreshMCP()
        }
        // iOS asks for local-network permission on the first browse, so start
        // immediately rather than hiding that behind the picker. Unit tests
        // host this app and must not raise that prompt.
        if ProcessInfo.processInfo.environment["XCTestConfigurationFilePath"] == nil {
            node.startScanning()
        }
        #if DEBUG
        DebugLaunch.apply(to: self)
        #endif
    }

    /// Dials MCP at the override URL, or at the selected NODE host on port
    /// 8000. A blocked decision leaves the client disconnected.
    func refreshMCP() {
        switch MCPEndpoint.decide(override: mcpOverride,
                                  nodeHost: node.selectedHost?.host,
                                  allowsLoopback: MCPEndpoint.allowsLoopbackHere) {
        case .blocked(let reason):
            mcpClient.disconnect()
            mcpConnected = false
            mcpStatus = reason
            mcps = "Off"
            pluginModels = [
                AIModel(name: "fastmcp: off", subtitle: nil, status: .alert)
            ]
        case .connect(let url):
            mcpStatus = "Connecting to \(url.absoluteString)…"
            mcps = "Connecting"
            pluginModels = [
                AIModel(name: "fastmcp: connecting", subtitle: url.host, status: .alert)
            ]
            mcpClient.connect(to: url) { [weak self] in
                guard let self else { return }
                self.mcpConnected = true
                self.mcpStatus = "Connected to \(url.absoluteString)"
                self.mcps = url.host ?? "On"
                self.pluginModels = [
                    AIModel(name: "fastmcp: connected", subtitle: url.absoluteString, status: .ready)
                ]
            }
        }
    }
    
    func executeCommand(_ command: String) {
        let msg = ChatMessage(isUser: false, text: "Executed command: /\(command)\n\n(This is a mock response. Integration coming soon!)")
        messages.append(msg)
        
        if command == "themes" {
            // Mock theme switch action
            messages.append(ChatMessage(isUser: false, text: "Opening Theme Picker..."))
        }
    }
    
    func sendMessage(_ text: String) {
        // Handle command palette direct entry if user types it manually and hits enter
        if text.hasPrefix("/") {
            let cmd = String(text.dropFirst()).trimmingCharacters(in: .whitespacesAndNewlines)
            executeCommand(cmd)
            return
        }
        
        // Append user message
        messages.append(ChatMessage(isUser: true, text: text))
        
        if mcpClient.isConnected {
            // Call the fastmcp tool
            mcpClient.callTool(name: "sync_gemini_command", args: ["command": text]) { response in
                DispatchQueue.main.async {
                    self.messages.append(ChatMessage(isUser: false, text: response))
                }
            }
        } else {
            messages.append(ChatMessage(isUser: false, text: mcpStatus))
        }
    }
}
