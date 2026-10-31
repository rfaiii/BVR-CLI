package model

import (
	"image"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/richavery/bvr-cli/internal/config"
	"github.com/richavery/bvr-cli/internal/csync"
	"github.com/richavery/bvr-cli/internal/ui/common"
	"github.com/stretchr/testify/require"
)

func TestLandingButtonTopAccountsForCWD(t *testing.T) {
	t.Parallel()
	// Buttons appear after the CWD line + 1 blank separator.
	// mainY=10, cwdHeight=1 → 10 + 1 + 1 = 12
	if got := landingButtonTop(10, 1); got != 12 {
		t.Fatalf("landing button top = %d, want 12", got)
	}
}

func TestLandingOffersGrokBotInsteadOfImageRecognition(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{
		Providers: csync.NewMap[string, config.ProviderConfig](),
		Agents:    map[string]config.Agent{},
		Options:   &config.Options{},
	}
	ui := &UI{
		com:   common.DefaultCommon(&testWorkspace{cfg: cfg}),
		state: uiLanding,
	}
	ui.layout.main = image.Rect(0, 0, 80, 40)
	plain := ansi.Strip(ui.landingView())
	require.Contains(t, plain, "CHAT W/ GROK BOT — ctrl+shift+g")
	require.NotContains(t, plain, "IMAGE RECOGNITION")
	require.Contains(t, plain, "\uf075")
	require.NotContains(t, plain, "— ctrl+b")
	require.NotContains(t, plain, "— ctrl+g")
}

func TestGrokBotChordDoesNotStealHelp(t *testing.T) {
	t.Parallel()
	km := DefaultKeyMap()
	require.Equal(t, []string{"ctrl+shift+g"}, km.Bots.Keys())
	require.Equal(t, []string{"ctrl+g"}, km.Help.Keys())
	require.Equal(t, "grok bots", km.Bots.Help().Desc)

	helpKey := tea.KeyPressMsg{Code: 'g', Mod: tea.ModCtrl}
	botsKey := tea.KeyPressMsg{Code: 'g', Mod: tea.ModCtrl | tea.ModShift}
	backKey := tea.KeyPressMsg{Code: 'b', Mod: tea.ModCtrl}
	require.True(t, key.Matches(helpKey, km.Help))
	require.False(t, key.Matches(helpKey, km.Bots))
	require.True(t, key.Matches(botsKey, km.Bots))
	require.False(t, key.Matches(botsKey, km.Help))
	require.False(t, key.Matches(backKey, km.Bots))
	require.False(t, key.Matches(backKey, km.Help))
}
