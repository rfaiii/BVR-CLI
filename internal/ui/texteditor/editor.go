// Package texteditor provides text editing capabilities for BVR-CLI.
//
// This package implements two editor modes:
//   - Embedded editor: Bubble Tea textarea component for quick notes and prompt adjustments
//   - Standalone editor: Micro editor integration for larger external documents
//
// The embedded editor uses Charmbracelet's bubbles.Textarea component with full
// Bubble Tea integration, while the standalone editor uses Go's os/exec to spawn
// the micro editor with proper TTY handling.
package texteditor

import (
	"os"
	"os/exec"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	charm "charm.land/lipgloss/v2"
)

// EditorState represents the different editor modes and states.
type EditorState int

const (
	EmbeddedState EditorState = iota
	StandaloneState
	ClosedState
)

// EditorMessage is a message sent to the Bubble Tea program to update editor state.
type EditorMessage struct {
	Operation string // "save", "close", "error"
	Content   string // Text content for embedded editor
	Filename  string // File path for standalone editor
}

// Editor represents the main editor model with both embedded and standalone modes.
type Editor struct {
	State       EditorState
	Embedded    textarea.Model
	Text        string
	ErrorMsg    string
	IsOpen      bool
	ThemeColors struct {
		Background string
		Border     string
		Cursor     string
		Success    string
		CloseBtn   string
	}
}

// NewEditor creates a new editor instance with the default embedded textarea.
func NewEditor() *Editor {
	embedded := textarea.New()
	embedded.SetWidth(60)
	embedded.SetHeight(20)
	embedded.Focus()

	return &Editor{
		State:    EmbeddedState,
		Embedded: embedded,
		IsOpen:   true,
		ThemeColors: struct {
			Background string
			Border     string
			Cursor     string
			Success    string
			CloseBtn   string
		}{
			Background: "#1e1e1e",
			Border:     "#b972ff",
			Cursor:     "#b972ff",
			Success:    "#39f66b",
			CloseBtn:   "#ff5caa",
		},
	}
}

// Init initializes the editor.
func (e *Editor) Init() tea.Cmd {
	return e.Embedded.Focus()
}

// Update handles editor messages and key events.
func (e *Editor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case EditorMessage:
		return e.handleEditorMessage(msg)
	case tea.KeyMsg:
		return e.handleKeyEvent(msg)
	}

	// Forward any other messages to the embedded textarea
	e.Embedded, cmd = e.Embedded.Update(msg)
	return e, cmd
}

// handleEditorMessage processes custom editor messages.
func (e *Editor) handleEditorMessage(msg EditorMessage) (tea.Model, tea.Cmd) {
	switch msg.Operation {
	case "save":
		e.Text = e.Embedded.View()
		return e, tea.Printf("✓ Saved successfully")
	case "close":
		e.IsOpen = false
		return e, tea.Quit
	case "error":
		e.ErrorMsg = msg.Content
		return e, tea.Printf("Error: %s", e.ErrorMsg)
	}
	return e, nil
}

// handleKeyEvent processes key events for the editor.
func (e *Editor) handleKeyEvent(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+s":
		e.Text = e.Embedded.View()
		return e, tea.Printf("✓ Saved successfully")
	case "esc":
		return e, tea.Quit
	}

	// Forward other keys to the embedded textarea
	e.Embedded, _ = e.Embedded.Update(msg)
	return e, nil
}

// View renders the editor UI.
func (e *Editor) View() tea.View {
	if !e.IsOpen {
		return tea.NewView("")
	}

	var s string

	// Render close button with pink color
	s += charm.NewStyle().Foreground(charm.Color(e.ThemeColors.CloseBtn)).Render("[ X ] Close")
	s += "\n\n"

	// Render embedded editor with purple border
	title := "BVR-CLI Embedded Editor"
	titleStyle := charm.NewStyle().
		Foreground(charm.Color(e.ThemeColors.Border)).
		Bold(true)
	s += titleStyle.Render(title)
	s += "\n"

	style := charm.NewStyle().
		Border(charm.NormalBorder()).
		BorderForeground(charm.Color(e.ThemeColors.Border)).
		Padding(1, 2)

	s += style.Render(e.Embedded.View())

	// Render controls
	s += "\n\n" + e.renderControls()

	// Render error message if any
	if e.ErrorMsg != "" {
		s += "\n" + charm.NewStyle().Foreground(charm.Color("red")).Render("Error: "+e.ErrorMsg)
	}

	return tea.NewView(s)
}

// renderControls displays the editor controls.
func (e *Editor) renderControls() string {
	controls := "Controls:"
	controls += "\n"
	controls += "  Type to enter text"
	controls += "\n"
	controls += "  Enter for new line"
	controls += "\n"
	controls += "  Ctrl+S to Save & Apply"
	controls += "\n"
	controls += "  Esc to Cancel / Close"
	controls += "\n"
	controls += "  Click [ X ] to close (if mouse enabled)"

	return charm.NewStyle().
		Foreground(charm.Color("#888888")).
		Render(controls)
}

// OpenInMicro opens a file in the micro editor.
func OpenInMicro(filename string) tea.Cmd {
	cmd := exec.Command("micro", filename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return tea.ExecProcess(cmd, nil)
}
