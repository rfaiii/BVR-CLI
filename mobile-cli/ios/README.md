# BVR-CLI Companion — iOS

This directory contains the iOS companion app for BVR-CLI.

## Setup

1. Open Xcode 15+.
2. Create a new iOS app project named `BvrCompanion`.
3. Set the interface to SwiftUI and the language to Swift 5.9+.
4. Copy the source files from this directory into the Xcode project.
5. Enable the following capabilities:
   - Network
   - Local Network (`Bonjour` / `NSLocalNetworkUsageDescription`)
6. Add the companion host URL scheme / local network discovery config in `Info.plist`.

## Theme

The app uses a dark ultra-contrast material theme to match the desktop CLI:

- Primary: `#39F66B`
- Secondary: `#B972FF`
- Background: `#121212`
- Surface: `#1E1E1E`
- Border: `#2C2C2C`
- Text primary: `#E0E0E0`
- Text secondary: `#B0BEC5`

## Requirements

- iOS 17.0+
- SwiftUI
- WebSocket-capable network layer
- Optional: React Native / Flutter alternative branch if native SwiftUI is not desired.
