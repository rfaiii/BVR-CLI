package model

import (
	"image"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/catwalk/pkg/catwalk"
	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/richavery/bvr-cli/internal/config"
	"github.com/richavery/bvr-cli/internal/csync"
	"github.com/richavery/bvr-cli/internal/ui/anim"
	attachments "github.com/richavery/bvr-cli/internal/ui/attachments"
	"github.com/richavery/bvr-cli/internal/ui/common"
	dialog "github.com/richavery/bvr-cli/internal/ui/dialog"
	"github.com/richavery/bvr-cli/internal/workspace"
	"github.com/stretchr/testify/require"
)

func TestCurrentModelSupportsImages(t *testing.T) {
	t.Parallel()

	t.Run("returns false when config is nil", func(t *testing.T) {
		t.Parallel()

		ui := newTestUIWithConfig(t, nil)
		require.False(t, ui.currentModelSupportsImages())
	})

	t.Run("returns false when coder agent is missing", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			Providers: csync.NewMap[string, config.ProviderConfig](),
			Agents:    map[string]config.Agent{},
		}
		ui := newTestUIWithConfig(t, cfg)
		require.False(t, ui.currentModelSupportsImages())
	})

	t.Run("returns false when model is not found", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			Providers: csync.NewMap[string, config.ProviderConfig](),
			Agents: map[string]config.Agent{
				config.AgentCoder: {Model: config.SelectedModelTypeLarge},
			},
		}
		ui := newTestUIWithConfig(t, cfg)
		require.False(t, ui.currentModelSupportsImages())
	})

	t.Run("returns true when current model supports images", func(t *testing.T) {
		t.Parallel()

		providers := csync.NewMap[string, config.ProviderConfig]()
		providers.Set("test-provider", config.ProviderConfig{
			ID: "test-provider",
			Models: []catwalk.Model{
				{ID: "test-model", SupportsImages: true},
			},
		})

		cfg := &config.Config{
			Models: map[config.SelectedModelType]config.SelectedModel{
				config.SelectedModelTypeLarge: {
					Provider: "test-provider",
					Model:    "test-model",
				},
			},
			Providers: providers,
			Agents: map[string]config.Agent{
				config.AgentCoder: {Model: config.SelectedModelTypeLarge},
			},
		}

		ui := newTestUIWithConfig(t, cfg)
		require.True(t, ui.currentModelSupportsImages())
	})
}

func newTestUIWithConfig(t *testing.T, cfg *config.Config) *UI {
	t.Helper()

	return &UI{
		com: &common.Common{
			Workspace: &testWorkspace{cfg: cfg},
		},
	}
}

// testWorkspace is a minimal [workspace.Workspace] stub for unit tests.
type testWorkspace struct {
	workspace.Workspace
	cfg *config.Config
}

func (w *testWorkspace) Config() *config.Config {
	return w.cfg
}

// newTestUIForBeaver builds a minimal but functional UI on the landing page,
// sufficient to exercise mascot click / mouse-gaze handling without a full
// session. The beaver rect is fixed at (5,5)-(18,10): 13-wide by 5-tall.
func newTestUIForBeaver(t *testing.T, state uiState, focus uiFocusState) *UI {
	t.Helper()

	cfg := &config.Config{
		Providers: csync.NewMap[string, config.ProviderConfig](),
		Agents:    map[string]config.Agent{},
	}
	com := common.DefaultCommon(&testWorkspace{cfg: cfg})

	u := &UI{
		com:         com,
		dialog:      dialog.NewOverlay(),
		attachments: &attachments.Attachments{},
		state:       state,
		focus:       focus,
		beaverRect:  image.Rect(5, 5, 18, 10), // 13-wide, 5-tall beaver at (5,5)
	}
	return u
}

// TestBeaverClickBoop verifies that clicking inside the mascot rect triggers
// the "what" reaction (beaverBoopUntil is set), which makes the homescreen
// hero render the what frame.
func TestBeaverClickBoop(t *testing.T) {
	t.Parallel()

	// Mouse clicks outside the rect do not boop the mascot.
	m, _ := newTestUIForBeaver(t, uiLanding, uiFocusEditor).Update(tea.MouseClickMsg{
		Button: uv.MouseLeft,
		X:      50,
		Y:      50,
	})
	ui := m.(*UI)
	require.False(t, time.Now().Before(ui.beaverBoopUntil), "click outside rect must not boop")

	// A left click inside the rect boops the mascot.
	m, _ = newTestUIForBeaver(t, uiLanding, uiFocusEditor).Update(tea.MouseClickMsg{
		Button: uv.MouseLeft,
		X:      10,
		Y:      7,
	})
	ui = m.(*UI)
	require.True(t, time.Now().Before(ui.beaverBoopUntil), "click inside rect must boop")

	// The boop window is short (~650ms).
	require.True(t, ui.beaverBoopUntil.After(time.Now()), "boop must be in the future")
}

// TestBeaverMouseDir verifies the mouse-gaze-to-direction mapping used by the
// homescreen mascot: gaze left/right of the mascot centers the beaver's
// direction faces; center gaze falls back to the keyboard cursor position.
func TestBeaverMouseDir(t *testing.T) {
	t.Parallel()

	m := newTestUIForBeaver(t, uiLanding, uiFocusEditor)

	// Default gaze is center → no mouse direction override.
	require.Equal(t, 0, m.beaverMouseDir(), "default gaze must be center")

	// Simulate mouse clearly left of the mascot.
	m.beaverGaze = anim.StateLookLeft
	require.Equal(t, -1, m.beaverMouseDir(), "look-left gaze must map to -1")

	// Simulate mouse clearly right of the mascot.
	m.beaverGaze = anim.StateLookRight
	require.Equal(t, 1, m.beaverMouseDir(), "look-right gaze must map to 1")

	// Other gaze states (up/down/center) fall back to center.
	for _, s := range []anim.MascotState{anim.StateCenter, anim.StateLookUp, anim.StateLookDown} {
		m.beaverGaze = s
		require.Equal(t, 0, m.beaverMouseDir(), "gaze %v must map to 0", s)
	}
}

// TestBeaverMouseGazeFromMotion verifies that mouse motion on the landing
// page updates the mascot's gaze (beaverGaze), which the homescreen hero uses
// to decide whether to look left or right.
func TestBeaverMouseGazeFromMotion(t *testing.T) {
	t.Parallel()

	u := newTestUIForBeaver(t, uiLanding, uiFocusEditor)

	// Before any motion, gaze is the default (center).
	require.Equal(t, anim.StateCenter, u.beaverGaze, "initial gaze must be center")

	// Mouse clearly left of the mascot center.
	leftOf := u.beaverRect.Min.X - 10
	m, _ := u.Update(tea.MouseMotionMsg{
		Button: uv.MouseNone,
		X:      leftOf,
		Y:      u.beaverRect.Min.Y,
	})
	um := m.(*UI)
	require.Equal(t, anim.StateLookLeft, um.beaverGaze, "motion left of mascot must set look-left gaze")
	require.Equal(t, -1, um.beaverMouseDir(), "look-left gaze must produce left direction")

	// Mouse clearly right of the mascot center.
	rightOf := u.beaverRect.Max.X + 10
	m, _ = u.Update(tea.MouseMotionMsg{
		Button: uv.MouseNone,
		X:      rightOf,
		Y:      u.beaverRect.Min.Y,
	})
	um = m.(*UI)
	require.Equal(t, anim.StateLookRight, um.beaverGaze, "motion right of mascot must set look-right gaze")
	require.Equal(t, 1, um.beaverMouseDir(), "look-right gaze must produce right direction")

	// Motion over the mascot center stays centered.
	centerX := u.beaverRect.Min.X + u.beaverRect.Dx()/2
	m, _ = u.Update(tea.MouseMotionMsg{
		Button: uv.MouseNone,
		X:      centerX,
		Y:      u.beaverRect.Min.Y,
	})
	um = m.(*UI)
	require.Equal(t, anim.StateCenter, um.beaverGaze, "motion over mascot center must stay centered")
}

// TestBeaverFrameSharedAcrossScreens verifies the homescreen and the
// working-screen sidebar render the exact same BVR mascot. Both must resolve
// through beaverFrameKey, so any state must produce identical frames.
func TestBeaverFrameSharedAcrossScreens(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		setup func(*UI)
	}{
		{"default", func(*UI) {}},
		{"errored", func(m *UI) { m.beaverErrored = true }},
		{"boop", func(m *UI) { m.beaverBoopUntil = time.Now().Add(time.Minute) }},
		{"busy", func(m *UI) { m.bangCancel = func() {} }},
		{"look left", func(m *UI) { m.beaverGaze = anim.StateLookLeft }},
		{"look right", func(m *UI) { m.beaverGaze = anim.StateLookRight }},
		{"facing left", func(m *UI) { m.beaverFacing = -1 }},
		{"facing right", func(m *UI) { m.beaverFacing = 1 }},
		{"resting", func(m *UI) { m.beaverResting = true }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// The homescreen hero is built by drawLanding.
			home := newTestUIForBeaver(t, uiLanding, uiFocusEditor)
			tc.setup(home)
			homeHero := stripANSI(home.beaverFrame())

			// The sidebar mascot is built by cacheSidebarLogo.
			work := newTestUIForBeaver(t, uiChat, uiFocusEditor)
			tc.setup(work)
			work.cacheSidebarLogo(24)

			require.Equal(t, homeHero, stripANSI(work.sidebarBeaverFrame),
				"homescreen and working screen must show the same mascot for state %q", tc.name)
			require.NotEmpty(t, homeHero, "mascot must not be empty")
		})
	}
}

// TestSidebarBeaverFrameUsesDenseMascot guards the migration off the old
// 12x7 mini mascot. The sidebar must render the same dense 13x5 frame as the
// homescreen hero.
func TestSidebarBeaverFrameUsesDenseMascot(t *testing.T) {
	t.Parallel()

	u := newTestUIForBeaver(t, uiChat, uiFocusEditor)
	u.cacheSidebarLogo(24)

	frame := stripANSI(u.sidebarBeaverFrame)
	require.Equal(t, 5, lipgloss.Height(frame),
		"sidebar mascot must be the dense 5-row frame, not the 7-row mini mascot")
	require.Equal(t, 13, lipgloss.Width(frame),
		"sidebar mascot must be the dense 13-cell frame, not the 12-cell mini mascot")

	// The homescreen hero and the sidebar frame must be byte-identical.
	require.Equal(t, stripANSI(u.beaverFrame()), frame,
		"sidebar mascot must be identical to the homescreen hero")
}

// TestVisibleBeaverRectFollowsState verifies mouse gaze and clicks resolve to
// the mascot actually on screen: the landing rect on the homescreen and the
// sidebar rect on the working screen.
func TestVisibleBeaverRectFollowsState(t *testing.T) {
	t.Parallel()

	home := newTestUIForBeaver(t, uiLanding, uiFocusEditor)
	require.Equal(t, home.beaverRect, home.visibleBeaverRect(), "homescreen must use the landing rect")
	require.False(t, home.visibleBeaverRect().Empty(), "homescreen rect must be non-empty")

	work := newTestUIForBeaver(t, uiChat, uiFocusEditor)
	work.sidebarBeaverRect = image.Rect(40, 2, 53, 7)
	require.Equal(t, work.sidebarBeaverRect, work.visibleBeaverRect(), "working screen must use the sidebar rect")

	// A click inside the sidebar rect boops the mascot on the working screen.
	// chat is populated because a miss would fall through to the chat handler.
	boops := newTestUIForBeaver(t, uiChat, uiFocusEditor)
	boops.chat = NewChat(boops.com, config.ScrollbarDefault)
	boops.sidebarBeaverRect = image.Rect(40, 2, 53, 7)
	m, _ := boops.Update(tea.MouseClickMsg{Button: uv.MouseLeft, X: 45, Y: 4})
	require.True(t, time.Now().Before(m.(*UI).beaverBoopUntil), "sidebar click must boop the mascot")

	// When the sidebar mascot is not drawn the hitbox is empty, so clicks in
	// the empty logo area cannot boop an invisible mascot.
	hidden := newTestUIForBeaver(t, uiChat, uiFocusEditor)
	require.True(t, hidden.visibleBeaverRect().Empty(), "hidden mascot must expose an empty hitbox")
	require.False(t, image.Pt(45, 4).In(hidden.visibleBeaverRect()), "empty hitbox must not contain any point")
}
