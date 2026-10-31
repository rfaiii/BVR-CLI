package model

import (
	"image"
	"testing"

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
	require.Contains(t, plain, "CHAT W/ GROK BOT — ctrl+b")
	require.NotContains(t, plain, "IMAGE RECOGNITION")
	require.Contains(t, plain, "\uf075")
}
