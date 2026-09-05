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
    var nodeName: String = "Local Device"
    var mcps: String = "None"
    var isReady: Bool = true
    var cpuUsage: Double = 1.0 // 100%
    var ramUsage: Double = 0.64 // 64%
    
    var selectedTaskType: Int = 0 // 0 for Large, 1 for Small
    var currentModelName: String = "OpenAI: GPT-5.6 Luna Pro"
    
    var models: [AIModel] = [
        AIModel(name: "ornith:latest", subtitle: "(Ollama - Local)", status: .ready)
    ]
    
    var pluginModels: [AIModel] = [
        AIModel(name: "fastmcp: disconnected", subtitle: nil, status: .alert)
    ]
    
    var messages: [ChatMessage] = []
    
    let mcpClient = MCPClient()
    
    var availableCommands: [AppCommand] = [
        AppCommand(command: "themes", description: "Change application theme", icon: "paintbrush.fill"),
        AppCommand(command: "ollama", description: "Manage local Ollama models", icon: "cpu"),
        AppCommand(command: "node", description: "Manage desktop NODE connections", icon: "network"),
        AppCommand(command: "finder", description: "Open file browser", icon: "folder.fill"),
        AppCommand(command: "cd", description: "Change working directory", icon: "arrow.uturn.right"),
        AppCommand(command: "mcp", description: "Manage MCP servers", icon: "server.rack")
    ]
    
    init() {
        // Automatically try to connect to MCP on launch
        mcpClient.connect {
            self.pluginModels = [
                AIModel(name: "fastmcp: connected", subtitle: nil, status: .ready),
                AIModel(name: "tool: sync_gemini_command", subtitle: nil, status: .smallDot)
            ]
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
            messages.append(ChatMessage(isUser: false, text: "MCP Server not connected. Run `uv run mcp-server` first!"))
        }
    }
}
