import SwiftUI

enum Theme {
    static let primary = Color(hex: "39F66B")
    static let secondary = Color(hex: "B972FF")
    static let background = Color(hex: "121212")
    static let surface = Color(hex: "1E1E1E")
    static let border = Color(hex: "2C2C2C")
    static let textPrimary = Color(hex: "E0E0E0")
    static let textSecondary = Color(hex: "B0BEC5")
    static let success = Color(hex: "39F66B")
    static let error = Color(hex: "FF5252")
}

extension Color {
    init(hex: String) {
        let hex = hex.trimmingCharacters(in: CharacterSet.alphanumerics.inverted)
        var int: UInt64 = 0
        Scanner(string: hex).scanHexInt64(&int)
        let a, r, g, b: UInt64
        switch hex.count {
        case 3:
            a = 255
            r = (int >> 8) * 17
            g = (int >> 4 & 0xF) * 17
            b = (int & 0xF) * 17
        case 6:
            a = 255
            r = int >> 16
            g = int >> 8 & 0xFF
            b = int & 0xFF
        case 8:
            a = int >> 24
            r = int >> 16 & 0xFF
            g = int >> 8 & 0xFF
            b = int & 0xFF
        default:
            a = 255
            r = 0
            g = 0
            b = 0
        }
        self.init(
            .sRGB,
            red: Double(r) / 255,
            green: Double(g) / 255,
            blue: Double(b) / 255,
            opacity: Double(a) / 255
        )
    }
}
