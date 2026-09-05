import SwiftUI

struct HeaderView: View {
    var body: some View {
        HStack {
            Text("/// BVR-CLI ///")
                .font(.system(size: 32, weight: .bold, design: .monospaced))
                .foregroundColor(Theme.primary)
            Spacer()
        }
        .padding(.horizontal)
        .padding(.top, 10)
    }
}
