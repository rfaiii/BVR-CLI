package model

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/richavery/bvr-cli/internal/ui/common"
	"github.com/richavery/bvr-cli/internal/ui/dialog"
	fimage "github.com/richavery/bvr-cli/internal/ui/image"
	uistyles "github.com/richavery/bvr-cli/internal/ui/styles"
	"github.com/stretchr/testify/require"
)

func newTourModel(step onboardingStep, caps *common.Capabilities) *onboardingModel {
	st := uistyles.CharmtonePantera()
	o := &onboardingModel{com: &common.Common{Styles: &st}, step: step}
	o.SetImageCapabilities(caps)
	return o
}

func TestShotCells(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name                             string
		termCols, termRows, cellW, cellH int
		wantCols, wantRows               int
	}{
		{"wide terminal leaves room for the text", 150, 46, 12, 24, 70, 22},
		// A terminal that never answered the cell-size query still gets a box
		// sized for the 2:1 cell every modern one has.
		{"unknown cell size assumes 2:1", 150, 46, 0, 0, 70, 22},
		{"mid-size terminal scales down", 120, 40, 12, 24, 51, 16},
		{"too short for a picture", 80, 24, 12, 24, 0, 0},
		{"narrow and short", 60, 20, 12, 24, 0, 0},
		{"no size at all", 0, 0, 12, 24, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cols, rows := shotCells(tc.termCols, tc.termRows, tc.cellW, tc.cellH, 960, 600)
			require.Equal(t, tc.wantCols, cols)
			require.Equal(t, tc.wantRows, rows)
		})
	}
}

// TestTourShotsAreEmbedded guards the mapping between steps and assets: a step
// naming a file the generator never wrote would silently fall back to text.
func TestTourShotsAreEmbedded(t *testing.T) {
	t.Parallel()

	seen := map[string]bool{}
	for step := onboardingStepWelcome; step <= onboardingStepComplete; step++ {
		shot, ok := shotForStep(step)
		if !ok {
			continue
		}
		require.False(t, seen[shot.id], "step %d reuses %s", step, shot.id)
		seen[shot.id] = true
		require.NotEmpty(t, shot.caption)

		width, height, err := shotPixelSize(shot.id)
		require.NoErrorf(t, err, "shots/%s.png must be embedded", shot.id)
		require.Positive(t, width)
		require.Positive(t, height)
	}

	// Notifications has no capture in resources/screenshots, so the step stays
	// text rather than borrowing someone else's screenshot.
	_, ok := shotForStep(onboardingStepNotifications)
	require.False(t, ok)
}

func TestSendShotNeedsGraphics(t *testing.T) {
	t.Parallel()

	o := newTourModel(onboardingStepHome, nil)
	require.Nil(t, o.sendShot())
	require.Equal(t, "fallback", o.renderShot("fallback"))

	// 16 colours and no kitty protocol: the half-block painter would have to
	// approximate a screenshot with sixteen swatches, so the text wins. This is
	// also the common case under tmux, which reports 256 colors — those do get
	// a picture.
	o = newTourModel(onboardingStepHome, &common.Capabilities{
		Profile: colorprofile.ANSI, Columns: 150, Rows: 46,
	})
	require.Nil(t, o.sendShot())
	require.Equal(t, "fallback", o.renderShot("fallback"))
}

func TestSendShotPaintsOnATmuxProfile(t *testing.T) {
	fimage.ResetCache()
	t.Cleanup(fimage.ResetCache)

	// tmux reports 256 colors even in front of a truecolor terminal, and the
	// tour still shows pictures there.
	o := newTourModel(onboardingStepHome, &common.Capabilities{
		Profile: colorprofile.ANSI256,
		Columns: 150, Rows: 46,
		PixelX: 150 * 16, PixelY: 46 * 32,
	})
	cmd := o.sendShot()
	require.NotNil(t, cmd)
	require.IsType(t, fimage.TransmittedMsg{}, cmd())
	require.NotContains(t, o.renderShot("fallback"), "fallback")
}

func TestSendShotPutsThePictureOnTheTerminal(t *testing.T) {
	fimage.ResetCache()
	t.Cleanup(fimage.ResetCache)

	caps := &common.Capabilities{
		Profile: colorprofile.TrueColor,
		Columns: 150, Rows: 46,
		PixelX: 150 * 12, PixelY: 46 * 24,
	}
	o := newTourModel(onboardingStepHome, caps)

	cmd := o.sendShot()
	require.NotNil(t, cmd)
	require.IsType(t, fimage.TransmittedMsg{}, cmd())
	require.True(t, fimage.HasTransmitted("home", o.shotCols, o.shotRows))

	view := o.renderShot("fallback")
	require.NotContains(t, view, "fallback")
	require.Contains(t, view, "Home: project")

	// Re-entering a step at the same size must not decode it again.
	require.Nil(t, o.sendShot())

	// With the kitty protocol answering, the picture goes out as raw escape
	// data instead of blocks.
	caps.KittyGraphics = true
	o = newTourModel(onboardingStepModels, caps)
	cmd = o.sendShot()
	require.NotNil(t, cmd)
	require.IsType(t, tea.RawMsg{}, cmd())
	require.NotContains(t, o.renderShot("fallback"), "fallback")
}

// TestStepChangeRequestsTheNextPicture: the tour only advances on Enter, and
// each step has its own asset, so the key handler has to hand back the transmit
// command rather than dropping it.
func TestStepChangeRequestsTheNextPicture(t *testing.T) {
	fimage.ResetCache()
	t.Cleanup(fimage.ResetCache)

	o := newTourModel(onboardingStepWelcome, &common.Capabilities{
		Profile: colorprofile.TrueColor,
		Columns: 150, Rows: 46,
		PixelX: 150 * 12, PixelY: 46 * 24,
	})
	require.Equal(t, onboardingStepWelcome, o.step)

	action := o.handleKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, action, "advancing to a step with a picture must schedule it")
	cmd, ok := action.(dialog.ActionCmd)
	require.True(t, ok)
	require.NotNil(t, cmd.Cmd)
	require.Equal(t, onboardingStepHome, o.step)
}
