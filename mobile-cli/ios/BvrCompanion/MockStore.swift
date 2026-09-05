import SwiftUI

enum ModelStatus: String {
    case ready
    case folder
    case alert
    case smallDot
    
    var iconName: String {
        switch self {
        case .ready: return "checkmark.circle.fill"
        case .folder: return "folder"
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
    let name: String
    let subtitle: String?
    let status: ModelStatus
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
        AIModel(name: "ornith:latest", subtitle: "(Ollama - Local)", status: .ready),
        AIModel(name: "qwen2.5-coder:3b-instruct", subtitle: nil, status: .ready),
        AIModel(name: "phi3:3.8b", subtitle: nil, status: .folder),
        AIModel(name: "ph14-mini:3.8b", subtitle: nil, status: .folder),
        AIModel(name: "ph14-mini: latest", subtitle: nil, status: .alert),
        AIModel(name: "ph14-mini: latest", subtitle: nil, status: .alert)
    ]
    
    var pluginModels: [AIModel] = [
        AIModel(name: "audio-engineering-plugin-lab", subtitle: nil, status: .smallDot),
        AIModel(name: "azure-aigateway", subtitle: nil, status: .smallDot)
    ]
}
