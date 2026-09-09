package dialog

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/richavery/bvr-cli/internal/ui/common"
)

const SoundsID = "sounds"

type soundOption struct {
	Name   string
	Volume string
}

var soundOptions = []soundOption{
	{Name: "High (100%)", Volume: "high"},
	{Name: "Low (50%)", Volume: "low"},
	{Name: "Silent", Volume: "silent"},
}

type Sounds struct {
	com      *common.Common
	selected int
}

func NewSounds(com *common.Common, currentVolume string) *Sounds {
	selected := 0
	for i, opt := range soundOptions {
		if opt.Volume == currentVolume {
			selected = i
		}
	}
	return &Sounds{com: com, selected: selected}
}

func (s *Sounds) ID() string { return SoundsID }

func (s *Sounds) HandleMsg(msg tea.Msg) Action {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	switch key.String() {
	case "esc":
		return ActionClose{}
	case "up", "k":
		s.selected = (s.selected + len(soundOptions) - 1) % len(soundOptions)
	case "down", "j":
		s.selected = (s.selected + 1) % len(soundOptions)
	case "enter":
		return ActionSetAudioVolume{Volume: soundOptions[s.selected].Volume}
	}
	return nil
}

func (s *Sounds) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	lines := []string{"BVR SOUND OPTIONS", "", "↑/↓ choose  enter apply  esc close", ""}
	for i, opt := range soundOptions {
		marker := "○"
		if i == s.selected {
			marker = "●"
		}
		lines = append(lines, fmt.Sprintf("%s %s", marker, opt.Name))
	}
	view := s.com.Styles.Dialog.View.Width(min(56, max(1, area.Dx()-4))).Render(strings.Join(lines, "\n"))
	DrawCenter(scr, area, view)
	return nil
}
