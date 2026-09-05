package dialog

import (
	_ "embed"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/richavery/bvr-cli/internal/ui/common"
)

//go:embed ollama_how_to.md
var ollamaHowTo string

const OllamaHowToID = "ollama-how-to"

type OllamaHowTo struct {
	com             *common.Common
	scroll          int
	up, down, close key.Binding
}

func NewOllamaHowTo(com *common.Common) *OllamaHowTo {
	return &OllamaHowTo{com: com, up: key.NewBinding(key.WithKeys("up", "k")), down: key.NewBinding(key.WithKeys("down", "j")), close: CloseKey}
}

func (o *OllamaHowTo) ID() string { return OllamaHowToID }

func (o *OllamaHowTo) HandleMsg(msg tea.Msg) Action {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	switch {
	case key.Matches(keyMsg, o.close):
		return ActionClose{}
	case key.Matches(keyMsg, o.up):
		o.scroll = max(0, o.scroll-1)
	case key.Matches(keyMsg, o.down):
		o.scroll++
	}
	return nil
}

func (o *OllamaHowTo) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	width := max(1, area.Dx()-2)
	height := max(1, area.Dy()-2)
	lines := strings.Split(ollamaHowTo, "\n")
	visible := max(1, height-3)
	start := min(o.scroll, max(0, len(lines)-visible))
	content := fixedLines(lines[start:], width, visible)
	content = append(content, "", "↑/↓ scroll  esc close")
	text := strings.Join(content, "\n")
	view := o.com.Styles.Dialog.View.Width(width).Height(height).Render(text)
	DrawCenter(scr, area, view)
	return nil
}
