package dialog

import (
	"image"
	"os"
	"path/filepath"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/richavery/bvr-cli/internal/ui/common"
)

// CreateFileID identifies the create file dialog.
const CreateFileID = "create-file"

// CreateFile is a dialog that allows users to create a new file.
type CreateFile struct {
	com       *common.Common
	input     textinput.Model
	help      help.Model
	keyMap    struct{ Close, Enter key.Binding }
	closeRect image.Rectangle
	panelRect image.Rectangle
	dir       string
}

var _ Dialog = (*CreateFile)(nil)

// NewCreateFile creates a new create file dialog.
func NewCreateFile(com *common.Common, dir string) (*CreateFile, tea.Cmd) {
	c := &CreateFile{com: com, dir: dir}
	c.help = help.New()
	c.help.Styles = com.Styles.DialogHelpStyles()
	c.input = textinput.New()
	c.input.SetVirtualCursor(false)
	c.input.SetStyles(com.Styles.TextInput)
	c.input.Placeholder = "Enter filename..."
	c.input.Focus()
	c.keyMap.Close = CloseKey
	c.keyMap.Close.SetHelp("esc", "close")
	c.keyMap.Enter = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "create file"))
	return c, textinput.Blink
}

// ID implements Dialog.
func (c *CreateFile) ID() string { return CreateFileID }

// HandleMsg implements Dialog.
func (c *CreateFile) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, c.keyMap.Close) {
			return ActionClose{}
		}
		var cmd tea.Cmd
		c.input, cmd = c.input.Update(msg)
		if msg.String() == "enter" {
			filename := c.input.Value()
			if filename != "" {
				path := filepath.Join(c.dir, filename)
				// Create the file if it doesn't exist
				if _, err := os.Stat(path); os.IsNotExist(err) {
					// ensure directory exists
					if err := os.MkdirAll(filepath.Dir(path), 0755); err == nil {
						if f, err := os.Create(path); err == nil {
							f.Close()
						}
					}
				}
				// Return the action to open the file in external editor
				return ActionFileBrowserOpenExternal{Path: path}
			}
			return nil
		}
		if cmd != nil {
			return ActionCmd{Cmd: cmd}
		}
	case tea.MouseClickMsg:
		if msg.Button == uv.MouseLeft && image.Pt(msg.X, msg.Y).In(c.closeRect) {
			return ActionClose{}
		}
		if msg.Button == uv.MouseLeft && !c.panelRect.Empty() && !image.Pt(msg.X, msg.Y).In(c.panelRect) {
			return ActionClose{}
		}
	}
	return nil
}

// Draw implements Dialog.
func (c *CreateFile) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	t := c.com.Styles
	dialogStyle := t.Dialog.View
	dw := min(area.Dx(), 60)
	title := common.DialogTitle(t, "Create File", dw, t.Dialog.TitleGradFromColor, t.Dialog.TitleGradToColor)
	inputArea := lipgloss.JoinHorizontal(lipgloss.Top, t.Dialog.InputPrompt.Render("Name "), c.input.View())
	helpView := renderDialogHelp(t, &c.help, c, dw)
	view := lipgloss.JoinVertical(lipgloss.Left, t.Dialog.Title.Render(title), dialogStyle.Render(inputArea), helpView)
	dialog := dialogStyle.Width(dw).Render(view)
	center := common.CenterRect(area, lipgloss.Width(dialog), lipgloss.Height(dialog))
	uv.NewStyledString(dialog).Draw(scr, center)

	closeText := closeLabel(dw)
	closeW := ansi.StringWidth(closeText)
	closeX := center.Min.X + dw - closeW - 1
	c.closeRect = image.Rect(closeX, center.Min.Y, closeX+closeW, center.Min.Y+1)
	c.panelRect = center
	closeStyle := t.Dialog.FileBrowser.Close
	uv.NewStyledString(closeStyle.Render(closeText)).Draw(scr, c.closeRect)
	return &tea.Cursor{Position: tea.Position{X: center.Min.X + 1, Y: center.Min.Y}}
}

// ShortHelp implements help.KeyMap.
func (c *CreateFile) ShortHelp() []key.Binding {
	return []key.Binding{c.keyMap.Enter, c.keyMap.Close}
}

// FullHelp implements help.KeyMap.
func (c *CreateFile) FullHelp() [][]key.Binding {
	return [][]key.Binding{{c.keyMap.Enter}, {c.keyMap.Close}}
}
