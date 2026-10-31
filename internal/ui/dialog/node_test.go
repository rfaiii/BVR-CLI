package dialog

import (
	"image"
	"testing"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/richavery/bvr-cli/internal/node"
	"github.com/richavery/bvr-cli/internal/ui/common"
	"github.com/richavery/bvr-cli/internal/ui/styles"
	"github.com/stretchr/testify/require"
)

func TestNodeSettingsDrawShowsDeviceNamesAndStatuses(t *testing.T) {
	theme := styles.DarkBvrTheme()
	dialog := NewNodeSettings(&common.Common{Styles: &theme})
	dialog.HandleMsg(NodeSettingsUpdateMsg{Devices: []node.Device{
		{ID: "offline", Name: "Grey Node", Status: node.DeviceStatusOffline},
		{ID: "error", Name: "Red Node", Status: node.DeviceStatusError},
		{ID: "online", Name: "Green Node", Status: node.DeviceStatusOnline},
	}})
	screen := uv.NewScreenBuffer(80, 20)
	dialog.Draw(screen, image.Rect(0, 0, 80, 20))
	output := screen.Render()
	for _, name := range []string{"Grey Node", "Red Node", "Green Node", "NODE CONNECTIONS"} {
		require.Contains(t, output, name)
	}
	require.Contains(t, output, "●")
	require.Contains(t, output, "○")
	require.Contains(t, output, "host")
}

func TestNodeSettingsEnterReturnsConnectCommand(t *testing.T) {
	theme := styles.DarkBvrTheme()
	settings := NewNodeSettings(&common.Common{Styles: &theme})
	settings.host = ""
	require.Nil(t, settings.HandleMsg(tea.KeyPressMsg{Code: tea.KeyEnter}))
	require.Contains(t, settings.status, "Host is required")

	settings.host = "127.0.0.1:7777"
	settings.focus = 1
	settings.HandleMsg(tea.KeyPressMsg{Text: "secret"})
	require.Equal(t, "secret", settings.token)
	action := settings.HandleMsg(tea.KeyPressMsg{Code: tea.KeyEnter})
	_, ok := action.(ActionCmd)
	require.True(t, ok)

	action = settings.HandleMsg(tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl})
	_, ok = action.(ActionCmd)
	require.True(t, ok, "ctrl+r probes saved NODEs")
}
