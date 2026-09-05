package model

import (
	"slices"

	"charm.land/lipgloss/v2"
	"github.com/richavery/bvr-cli/internal/node"
	"github.com/richavery/bvr-cli/internal/ui/common"
)

func (m *UI) nodeInfo(width, maxItems int) string {
	t := m.com.Styles
	devices := node.Devices()
	if len(devices) == 0 {
		return common.Section(t, t.Resource.Heading.Render("NODE")+"\n"+t.Resource.AdditionalText.Render("No devices"), width)
	}
	items := make([]string, 0, len(devices))
	for _, device := range devices {
		icon := t.Resource.OfflineIcon.Render("○")
		switch device.Status {
		case node.DeviceStatusOnline:
			icon = t.Resource.OnlineIcon.Render("●")
		case node.DeviceStatusError:
			icon = t.Resource.ErrorIcon.Render("●")
		}
		items = append(items, icon+" "+t.Resource.Name.Render(device.DisplayName())+" "+t.Resource.AdditionalText.Render(device.ConnectionLabel()))
	}
	if maxItems > 0 && len(items) > maxItems {
		items = append(slices.Delete(items, maxItems-1, len(items)), t.Resource.AdditionalText.Render("…"))
	}
	return common.Section(t, t.Resource.Heading.Render("NODE")+"\n"+lipgloss.JoinVertical(lipgloss.Left, items...), width)
}
