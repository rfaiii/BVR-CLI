package dialog

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/richavery/bvr-cli/internal/node"
	"github.com/richavery/bvr-cli/internal/ui/common"
)

const NodeSettingsID = "node-settings"

type NodeSettingsUpdateMsg struct{ Devices []node.Device }

type NodeSettings struct {
	com     *common.Common
	devices []node.Device
}

func NewNodeSettings(com *common.Common) *NodeSettings {
	return &NodeSettings{com: com, devices: node.Devices()}
}

func (n *NodeSettings) ID() string { return NodeSettingsID }

func (n *NodeSettings) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case NodeSettingsUpdateMsg:
		n.devices = msg.Devices
	case tea.KeyPressMsg:
		switch {
		case strings.EqualFold(msg.String(), "esc"):
			return ActionClose{}
		case strings.EqualFold(msg.String(), "r"):
			return ActionCmd{Cmd: discoverNodesCmd()}
		}
	}
	return nil
}

func discoverNodesCmd() tea.Cmd {
	return func() tea.Msg {
		// Discovery is intentionally performed off the Bubble Tea update path.
		node.DiscoverDevices()
		return NodeSettingsUpdateMsg{Devices: node.Devices()}
	}
}

func (n *NodeSettings) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	t := n.com.Styles
	lines := []string{"NODE CONNECTIONS", "", "r refresh  esc close", ""}
	for _, device := range n.devices {
		status := t.Resource.OfflineIcon.Render("○")
		if device.Status == node.DeviceStatusOnline {
			status = t.Resource.OnlineIcon.Render("●")
		} else if device.Status == node.DeviceStatusError {
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
