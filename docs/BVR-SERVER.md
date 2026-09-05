# BVR-CLI Companion Server (Host-Side)

This document describes the process and architectural requirements for building the **Companion Server** inside the Go `bvr-cli`. This server acts as the bridge that the iOS/Android mobile companion apps connect to via WebSockets.

## 1. Architectural Overview

The mobile apps are designed to be "dumb terminals" or remote controls for the main `bvr-cli` running on your Mac/Linux machine. 

To achieve this, `bvr-cli` needs an embedded HTTP/WebSocket server that listens on your local network (e.g., `0.0.0.0:8080`), accepts connections from the companion app, and streams real-time data back and forth.

**Connection Flow:**
1. User starts the CLI bridge: `bvr-cli server --companion --port 8080` (or similar).
2. The Go CLI starts an HTTP listener on `0.0.0.0:8080`.
3. The iOS Companion app attempts to connect to `ws://<your-mac-ip>:8080/companion`.
4. The Go CLI upgrades the connection to a WebSocket.
5. Bidirectional JSON messages are exchanged to drive the mobile UI.

## 2. Go Implementation Details

### A. HTTP & WebSocket Server
You'll need a standard Go HTTP server and a WebSocket library (like `github.com/gorilla/websocket`).

```go
package companion

import (
    "net/http"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow local network connections
    },
}

func HandleCompanion(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    defer conn.Close()

    // Start read/write loop
    for {
        messageType, p, err := conn.ReadMessage()
        if err != nil {
            return
        }
        // Handle incoming commands from iOS...
    }
}
```

### B. Command Integration
You need to expose this server via Cobra commands in `main.go` or `internal/cmd`. 
You could add it to the existing `bvr-cli server` command by adding a `--companion` flag, or create a dedicated command like `bvr-cli companion serve`.

```go
// Example Cobra Command
var serveCompanionCmd = &cobra.Command{
    Use:   "companion-server",
    Short: "Starts the WebSocket bridge for mobile companions",
    Run: func(cmd *cobra.Command, args []string) {
        http.HandleFunc("/companion", companion.HandleCompanion)
        log.Println("Companion server listening on :8080...")
        http.ListenAndServe("0.0.0.0:8080", nil) // Note: 0.0.0.0 allows local network access
    },
}
```

## 3. Protocol & Data Model (JSON)

The iOS app expects structured data. You'll need to define a shared JSON schema for the WebSocket payloads. 

**Example Event Types:**
*   `SESSION_LIST`: Go sends the list of active BVR sessions to iOS.
*   `SESSION_CONNECT`: iOS tells Go to attach to a specific session ID.
*   `TERMINAL_OUTPUT`: Go streams raw terminal chunks to iOS.
*   `USER_PROMPT`: iOS sends a new prompt from the user to Go.

**Example Payload Structure (Go struct):**
```go
type CompanionMessage struct {
    Type    string          `json:"type"`    // e.g., "USER_PROMPT"
    Payload json.RawMessage `json:"payload"` // Variable data
}
```

## 4. Security & Local Network

*   **Host Binding:** Ensure the HTTP server binds to `0.0.0.0` or the specific LAN IP (`192.168.x.x`) so the iPhone can reach it. Binding to `127.0.0.1` will strictly prevent the iPhone from connecting (unless testing on an iOS Simulator, which shares the Mac's localhost).
*   **Authentication (Optional but Recommended):** In the future, you may want to generate a 4-digit PIN in the CLI that the iOS app must provide in the initial connection request to prevent unauthorized access on public WiFi.

## 5. Wiring It Up Checklist

1. [ ] Add `gorilla/websocket` dependency to `go.mod`.
2. [ ] Create `internal/companion/server.go` to handle WebSocket upgrading and the read/write pump.
3. [ ] Define the JSON structs for the protocol in `internal/companion/models.go`.
4. [ ] Wire the HTTP handler into a new or existing Cobra command in `internal/cmd`.
5. [ ] Launch the server and connect from the iOS Simulator!
