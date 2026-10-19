package dialog

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/richavery/bvr-cli/internal/ggwave"
	"github.com/richavery/bvr-cli/internal/ui/common"
)

// PairingConfirmMsg is sent when the user confirms or rejects a GGWave pairing.
type PairingConfirmMsg struct {
	Envelope *ggwave.PairingEnvelope
	Accepted bool
}

// PairingConfirmComponent shows a GGWave pairing confirmation dialog.
type PairingConfirmComponent struct {
	com        *common.Common
	envelope   *ggwave.PairingEnvelope
	confirmed  bool
	keyLeft    key.Binding
	keyRight   key.Binding
	keyYes     key.Binding
	keyNo      key.Binding
	keyEnter   key.Binding
	keyClose   key.Binding
	focused    bool
	lastWidth  int
}

// NewPairingConfirmComponent creates a new pairing confirmation dialog.
func NewPairingConfirmComponent(com *common.Common, envelope *ggwave.PairingEnvelope) *PairingConfirmComponent {
	return &PairingConfirmComponent{
		com:      com,
		envelope: envelope,
		keyLeft:  key.NewBinding(key.WithKeys("left"), key.WithHelp("←/→", "switch")),
		keyRight: key.NewBinding(key.WithKeys("right"), key.WithHelp("←/→", "switch")),
		keyYes:   key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "accept")),
		keyNo:    key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "reject")),
		keyEnter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		keyClose: CloseKey,
	}
}

func (p *PairingConfirmComponent) ID() string { return "ggwave-pairing-confirm" }

func (p *PairingConfirmComponent) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, p.keyLeft), key.Matches(msg, p.keyRight):
			p.confirmed = !p.confirmed
			return nil
		case key.Matches(msg, p.keyYes):
			p.confirmed = true
			return p.accept()
		case key.Matches(msg, p.keyNo):
			p.confirmed = false
			return p.reject()
		case key.Matches(msg, p.keyEnter):
			if p.confirmed {
				return p.accept()
			}
			return p.reject()
		case key.Matches(msg, p.keyClose):
			return p.reject()
		}
	}
	return nil
}

func (p *PairingConfirmComponent) accept() Action {
	return ActionCmd{Cmd: func() tea.Msg {
		return PairingConfirmMsg{Envelope: p.envelope, Accepted: true}
	}}
}

func (p *PairingConfirmComponent) reject() Action {
	return ActionCmd{Cmd: func() tea.Msg {
		return PairingConfirmMsg{Envelope: p.envelope, Accepted: false}
	}}
}

func (p *PairingConfirmComponent) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	t := p.com.Styles
	w := area.Dx()
	if w <= 0 {
		w = p.lastWidth
	}
	if w <= 0 {
		w = 70
	}
	p.lastWidth = w

	e := p.envelope
	expiresAt := time.Unix(e.ExpiresAt, 0)
	expiresIn := time.Until(expiresAt)
	expiresStr := fmt.Sprintf("in %s", expiresIn.Round(time.Second))
	if expiresIn <= 0 {
		expiresStr = "expired"
	}

	// Use Resource.AdditionalText for subtle/muted text
	subtle := t.Resource.AdditionalText

	lines := []string{
		t.Dialog.Title.Render("🔗  GGWAVE NODE PAIRING REQUEST"),
		"",
		subtle.Render(fmt.Sprintf("Node ID:   %s", e.NodeID)),
		subtle.Render(fmt.Sprintf("Endpoint:  %s", e.Endpoint)),
		subtle.Render(fmt.Sprintf("Token:     %s", e.Token)),
		subtle.Render(fmt.Sprintf("Expires:   %s (%s)", expiresAt.Format("15:04:05"), expiresStr)),
		"",
	}

	if p.confirmed {
		lines = append(lines, t.Resource.OnlineIcon.Render("  [ ✓ ]  ACCEPT  ") + "  " + t.Resource.OfflineIcon.Render("  [   ]  REJECT  "))
	} else {
		lines = append(lines, t.Resource.OnlineIcon.Render("  [   ]  ACCEPT  ") + "  " + t.Resource.OfflineIcon.Render("  [ ✓ ]  REJECT  "))
	}

	view := t.Dialog.View.Width(min(76, max(1, w-4))).Render(
		lipgloss.JoinVertical(lipgloss.Left, lines...),
	)
	DrawCenter(scr, area, view)
	return nil
}

// OpenPairingConfirm opens the GGWave pairing confirmation dialog.
func OpenPairingConfirm(com *common.Common, envelope *ggwave.PairingEnvelope) Action {
	return ActionOpenDialog{
		DialogID: "ggwave-pairing-confirm",
	}
}