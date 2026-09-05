# texteditor

Package `texteditor` provides text editing capabilities for BVR-CLI.

## Overview

This package implements two editor modes for the BVR-CLI application:

- **Embedded Editor**: A Bubble Tea `textarea` component for quick notes, prompt adjustments, and internal configuration changes. Keeps the user within the application context.
- **Standalone Editor**: Integration with the `micro` editor for larger external documents where the user expects a full-screen experience.

## Features

### Embedded Editor (Bubble Tea)

- **Theme Integration**: Dark terminal background with consistent color scheme
- **Active Borders/Cursor**: Purple (`#b972ff`) for focused state
- **Save/Success Highlights**: Neon green (`#39f66b`) for save confirmations
- **Close Button**: Pink (`#ff5caa`) `[ X ]` close button rendered in the header
- **Word Wrap**: Enabled by default for comfortable text entry
- **Key Controls**:
  - `Type` to enter text
  - `Enter` for new line
  - `Ctrl+S` to Save & Apply
  - `Esc` to Cancel / Close
  - Click `[ X ]` (if terminal mouse reporting is enabled) to close

### Standalone Editor (Micro)

- Spawns `micro <filename>` via Go's `os/exec`
- Proper TTY handling with `tea.ExecProcess` and explicit stdin/stdout/stderr mapping
- Bubble Tea suspends its event loop while `micro` is running
- Controls:
  - `Mouse` to click and position cursor
  - `Ctrl+S` to Save
  - `Ctrl+Q` to Quit
  - `Ctrl+C` / `Ctrl+V` to Copy and Paste

## Usage

```go
import "github.com/richavery/bvr-cli/internal/ui/texteditor"

// Create a new editor instance
editor := texteditor.NewEditor()

// Use within a Bubble Tea program
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle editor messages
    // ...
}

// Open a file in the standalone micro editor
cmd := texteditor.OpenInMicro("config.toml")
```

## Architecture

- `Editor` struct implements `tea.Model` for Bubble Tea integration
- `EditorMessage` provides custom message types for save/close/error operations
- `EditorState` tracks whether the editor is in embedded, standalone, or closed mode
- `ThemeColors` struct holds the color configuration for the editor UI

## Color Scheme

| Color | Hex Code | Usage |
|-------|----------|-------|
| Purple | `#b972ff` | Active border, cursor, focus indicator |
| Neon Green | `#39f66b` | Save confirmations, success highlights |
| Pink | `#ff5caa` | Close button, interrupts |
| Dark | `#1e1e1e` | Background |

## Dependencies

- `charm.land/bubbles/v2/textarea` — Embedded textarea component
- `charm.land/lipgloss/v2` — Styling and borders
- `charm.land/bubbletea/v2` — Terminal UI framework
- `os/exec` — Standalone editor process spawning

## Integration with BVR-CLI

The text editor is accessible from the BVR-CLI main interface. Users can trigger the editor through commands like `:edit` or `:open` to open files for text editing. The editor integrates with the existing UI theme and color scheme.

See the [BVR-CLI Editor Specification](../../bvr-cli-editor-spec.md) for full details on the design and implementation requirements.
