import SwiftUI
import Combine

struct GlitchText: View {
    let text: String
    
    @State private var offset: CGSize = .zero
    @State private var opacity: Double = 1.0
    
    let timer = Timer.publish(every: 0.15, on: .main, in: .common).autoconnect()
    
    var body: some View {
        Text(text)
            .font(.custom("JetBrainsMonoNerdFontMono-Regular", size: 10))
            .foregroundColor(Theme.primary)
            .offset(offset)
            .opacity(opacity)
            .animation(.interactiveSpring(response: 0.1, dampingFraction: 0.5, blendDuration: 0.1), value: offset)
            .onReceive(timer) { _ in
                // 40% chance to glitch on each tick to save battery while still looking cool
                if Double.random(in: 0...1) > 0.6 {
                    offset = CGSize(width: CGFloat.random(in: -2...2), height: CGFloat.random(in: -1...1))
                    opacity = Double.random(in: 0.5...1.0)
                    
                    // Reset quickly to create a stutter effect
                    DispatchQueue.main.asyncAfter(deadline: .now() + 0.05) {
                        offset = .zero
                        opacity = 1.0
                    }
                }
            }
    }
}
