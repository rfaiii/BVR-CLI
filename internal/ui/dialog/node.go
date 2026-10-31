package dialog

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/richavery/bvr-cli/internal/node"
	"github.com/richavery/bvr-cli/internal/ui/common"
)

const NodeSettingsID = "node-settings"

// NodeSettingsUpdateMsg refreshes the device list after a probe or a
// manual connect.
type NodeSettingsUpdateMsg struct {
	Devices []node.Device
	Status  string
	Err     string
}

// NodeSettings is the NODE connection dialog. Host and token are typed
// here; refresh probes saved NODE rows rather than local dev ports.
type NodeSettings struct {
	com     *common.Common
	devices []node.Device
	host    string
	token   string
	focus   int
	status  string
}

func NewNodeSettings(com *common.Common) *NodeSettings {
	return &NodeSettings{com: com, devices: node.Devices(), host: "127.0.0.1:7777"}
}

func (n *NodeSettings) ID() string { return NodeSettingsID }

func (n *NodeSettings) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case NodeSettingsUpdateMsg:
		n.devices = msg.Devices
		switch {
		case msg.Err != "":
			n.status = msg.Err
		case msg.Status != "":
			n.status = msg.Status
		}
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			return ActionClose{}
		case "tab", "shift+tab":
			n.focus = (n.focus + 1) % 2
		case "enter":
			return n.connect()
		case "backspace":
			n.backspace()
		case "ctrl+r":
			n.status = "Probing saved NODEs…"
			return ActionCmd{Cmd: probeNodesCmd()}
		default:
			if msg.Text != "" && msg.Mod == 0 {
				n.insert(msg.Text)
			}
		}
	}
	return nil
}

func (n *NodeSettings) connect() Action {
	host := strings.TrimSpace(n.host)
	if host == "" {
		n.status = "Host is required"
		return nil
	}
	n.status = "Connecting…"
	token := n.token
	return ActionCmd{Cmd: connectNodeCmd(host, token)}
}

func (n *NodeSettings) insert(text string) {
	if n.focus == 0 {
		if len(n.host)+len(text) > 512 {
			return
		}
		n.host += text
		return
	}
	if len(n.token)+len(text) > 512 {
		return
	}
	n.token += text
}

func (n *NodeSettings) backspace() {
	if n.focus == 0 {
		n.host = dropLastRune(n.host)
		return
	}
	n.token = dropLastRune(n.token)
}

func dropLastRune(value string) string {
	if value == "" {
		return ""
	}
	_, size := utf8.DecodeLastRuneInString(value)
	return value[:len(value)-size]
}

func connectNodeCmd(host, token string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		device, err := node.Connect(ctx, host, token, "")
		update := NodeSettingsUpdateMsg{Devices: node.Devices()}
		if err != nil {
			update.Err = err.Error()
			return update
		}
		update.Status = "Connected to " + device.DisplayName()
		return update
	}
}

func probeNodesCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		node.ProbeSavedNodes(ctx)
		return NodeSettingsUpdateMsg{Devices: node.Devices(), Status: "Probed saved NODEs"}
	}
}

func (n *NodeSettings) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	t := n.com.Styles
	hostMark, tokenMark := " ", " "
	if n.focus == 0 {
		hostMark = ">"
	} else {
		tokenMark = ">"
	}
	tokenView := strings.Repeat("•", utf8.RuneCountInString(n.token))
	if tokenView == "" {
		tokenView = "(none)"
	}
	lines := []string{
		"NODE CONNECTIONS",
		"",
		"enter connect   ctrl+r refresh   tab field   esc close",
		fmt.Sprintf("%s host   %s", hostMark, n.host),
		fmt.Sprintf("%s token  %s", tokenMark, tokenView),
	}
	if n.status != "" {
		lines = append(lines, "", n.status)
	}
	lines = append(lines, "")
	for _, device := range n.devices {
		status := t.Resource.OfflineIcon.Render("○")
		switch device.Status {
		case node.DeviceStatusOnline:
			status = t.Resource.OnlineIcon.Render("●")
		case node.DeviceStatusError:
			status = t.Resource.ErrorIcon.Render("●")
		}
		endpoint := device.Address
		if device.TransportURL != "" {
			endpoint = device.TransportURL
		}
		lines = append(lines, fmt.Sprintf("%s %s  %s  %s", status, device.DisplayName(), device.ConnectionLabel(), endpoint))
	}
	if len(n.devices) == 0 {
		lines = append(lines, t.Resource.AdditionalText.Render("No devices discovered"))
	}
	view := t.Dialog.View.Width(min(72, max(1, area.Dx()-4))).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	DrawCenter(scr, area, view)
	return nil
}
