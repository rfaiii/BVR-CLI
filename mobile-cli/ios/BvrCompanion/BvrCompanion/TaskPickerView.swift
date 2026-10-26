import SwiftUI

struct TaskPickerView: View {
    @Environment(MockStore.self) private var store
    
    var body: some View {
        HStack(spacing: 0) {
            Button(action: {
                store.selectedTaskType = 0
            }) {
                Text("Large Task")
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 8)
                    .background(store.selectedTaskType == 0 ? Theme.primary : Color.clear)
                    .foregroundColor(store.selectedTaskType == 0 ? .black : Theme.textPrimary)
            }
            
            Button(action: {
                store.selectedTaskType = 1
            }) {
                Text("Small Task")
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 8)
                    .background(store.selectedTaskType == 1 ? Theme.primary : Color.clear)
                    .foregroundColor(store.selectedTaskType == 1 ? .black : Theme.textPrimary)
            }
        }
        .background(Theme.surface)
        .cornerRadius(8)
        .overlay(
            RoundedRectangle(cornerRadius: 8)
                .stroke(Theme.primary, lineWidth: 1) // Outline is green
        )
        .padding(.horizontal)
    }
}
