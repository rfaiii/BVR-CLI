package anim

import (
	"strings"
	"testing"

	"github.com/richavery/bvr-cli/internal/ui/boot"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

// TestStartSupersedesPreviousChain verifies that calling Start() twice
// does not result in two concurrent tick chains advancing the same Anim.
// The second Start() bumps the generation, so ticks from the first chain
// (carrying the old generation) are dropped by Animate().
func TestStartSupersedesPreviousChain(t *testing.T) {
	t.Parallel()

	a := New(Settings{ID: "test", Size: 5})

	// First chain.
	cmd1 := a.Start()
	gen1 := a.gen.Load()
	require.Equal(t, int64(1), gen1)

	// Second chain supersedes the first.
	cmd2 := a.Start()
	gen2 := a.gen.Load()
	require.Equal(t, int64(2), gen2)

	// Execute both commands to get their StepMsgs.
	msg1 := cmd1().(StepMsg)
	msg2 := cmd2().(StepMsg)

	require.Equal(t, gen1, msg1.Gen, "first chain carries old generation")
	require.Equal(t, gen2, msg2.Gen, "second chain carries new generation")

	// The old-generation tick must be dropped.
	framesBefore := a.framesSinceStart.Load()
	next := a.Animate(msg1)
	require.Nil(t, next, "old-generation tick must not schedule another step")
	require.Equal(t, framesBefore, a.framesSinceStart.Load(),
		"old-generation tick must not advance the frame")

	// The new-generation tick must advance.
	next = a.Animate(msg2)
	require.NotNil(t, next, "current-generation tick must schedule another step")
	require.Equal(t, framesBefore+1, a.framesSinceStart.Load(),
		"current-generation tick must advance the frame")
}

// TestBeaverFrameEveryKeyResolves guards against a pose key that silently
// falls back to the resting face. The Alpha and Beta frame maps hold disjoint
// poses (Alpha: center/left-dir/right-dir/what; Beta: error/confused), so
// looking in only one map made "confused" unreachable unless errored was also
// set, and the agent-busy expression never rendered.
func TestBeaverFrameEveryKeyResolves(t *testing.T) {
	t.Parallel()

	all := []BeaverFrameKey{
		FrameCenter, FrameLeftDir, FrameRightDir,
		FrameWhat, FrameError, FrameConfused,
	}

	for _, key := range all {
		for _, errored := range []bool{false, true} {
			got := stripANSI(BeaverFrame(key, errored))
			want := stripANSI(BeaverFrame(FrameCenter, false))
			if got == want && key != FrameCenter {
				t.Errorf("%q (errored=%v) fell back to the resting pose", key, errored)
			}
			if lipgloss.Width(stripANSI(BeaverFrame(key, errored))) != DenseFrameWidth {
				t.Errorf("%q width = %d, want %d",
					key, lipgloss.Width(stripANSI(BeaverFrame(key, errored))), DenseFrameWidth)
			}
		}
	}
}

// TestBeaverFrameConfusedRendersWhenNotErrored pins the specific regression:
// the agent-busy pose must render without the error flag set.
func TestBeaverFrameConfusedRendersWhenNotErrored(t *testing.T) {
	t.Parallel()

	confused := stripANSI(BeaverFrame(FrameConfused, false))
	center := stripANSI(BeaverFrame(FrameCenter, false))
	if confused == center {
		t.Fatal("confused pose must differ from the resting pose when errored is false")
	}
	if !strings.Contains(confused, "9_g") {
		t.Errorf("confused pose did not render the Beta expression, got:\n%s", confused)
	}
}

// TestBeaverFrameErrorTakesPriorityWhenErrored verifies the x-ray map wins
// while the error state is active.
func TestBeaverFrameErrorTakesPriorityWhenErrored(t *testing.T) {
	t.Parallel()

	got := stripANSI(BeaverFrame(FrameError, true))
	if got == stripANSI(BeaverFrame(FrameCenter, true)) {
		t.Fatal("error pose must differ from the resting pose when errored is true")
	}
}

// TestBeaverFrameKeepsDenseFootprint verifies every pose keeps the 13x5 grid so
// the layout and click hitbox never shift between frames.
func TestBeaverFrameKeepsDenseFootprint(t *testing.T) {
	t.Parallel()

	for _, key := range []BeaverFrameKey{
		FrameCenter, FrameLeftDir, FrameRightDir, FrameWhat, FrameError, FrameConfused,
	} {
		for _, errored := range []bool{false, true} {
			frame := stripANSI(BeaverFrame(key, errored))
			if got := lipgloss.Width(frame); got != DenseFrameWidth {
				t.Errorf("%q width = %d, want %d", key, got, DenseFrameWidth)
			}
			if got := lipgloss.Height(frame); got != DenseFrameHeight {
				t.Errorf("%q height = %d, want %d", key, got, DenseFrameHeight)
			}
		}
	}
}

func stripANSI(s string) string {
	return ansi.Strip(s)
}

func TestBeaverHeroFrameKeepsStableWidthAndChangesPose(t *testing.T) {
	t.Parallel()
	a := BeaverHeroFrame(0, false)
	b := BeaverHeroFrame(6, false)
	if lipgloss.Width(a) != lipgloss.Width(b) {
		t.Fatalf("hero frame width changed: %d vs %d", lipgloss.Width(a), lipgloss.Width(b))
	}
	if a == b {
		t.Fatal("hero animation did not change pose")
	}
}

func TestMascotFramesKeepCanonicalFootprints(t *testing.T) {
	t.Parallel()
	states := []MascotState{
		StateCenter, StateLookLeft, StateLookRight, StateLookUp, StateLookDown,
		StateLaugh, StateShocked, StateDanceA, StateDanceB, StateChompA,
		StateChompB, StateClickBoop,
	}
	for _, state := range states {
		large := LargeMascotFrame(state, 0, false)
		mini := MiniMascotFrame(state, 0, false)
		if got := lipgloss.Width(large); got != LargeMascotWidth {
			t.Errorf("%v large width = %d, want %d", state, got, LargeMascotWidth)
		}
		if got := lipgloss.Height(large); got != LargeMascotHeight {
			t.Errorf("%v large height = %d, want %d", state, got, LargeMascotHeight)
		}
		if got := lipgloss.Width(mini); got != MiniMascotWidth {
			t.Errorf("%v mini width = %d, want %d", state, got, MiniMascotWidth)
		}
		if got := lipgloss.Height(mini); got != MiniMascotHeight {
			t.Errorf("%v mini height = %d, want %d", state, got, MiniMascotHeight)
		}
		if got := len(strings.Split(large, "\n")); got != LargeMascotHeight {
			t.Errorf("%v large rows = %d, want %d", state, got, LargeMascotHeight)
		}
	}
}

func TestEyeDirectionUsesStableBuckets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		dx, dy int
		want   MascotState
	}{
		{0, 0, StateCenter},
		{-10, 0, StateLookLeft},
		{10, 0, StateLookRight},
		{0, -10, StateLookUp},
		{0, 10, StateLookDown},
		{2, 1, StateCenter},
	}
	for _, tt := range tests {
		if got := EyeDirection(tt.dx, tt.dy); got != tt.want {
			t.Errorf("EyeDirection(%d, %d) = %v, want %v", tt.dx, tt.dy, got, tt.want)
		}
	}
}

// TestStopKillsChain verifies that Stop() bumps the generation so any
// in-flight tick chain is terminated.
func TestStopKillsChain(t *testing.T) {
	t.Parallel()

	a := New(Settings{ID: "test", Size: 5})
	cmd := a.Start()
	msg := cmd().(StepMsg)

	a.Stop()
	require.NotEqual(t, msg.Gen, a.gen.Load(), "Stop must bump the generation")

	// The in-flight tick must be dropped.
	framesBefore := a.framesSinceStart.Load()
	next := a.Animate(msg)
	require.Nil(t, next, "tick after Stop must not schedule another step")
	require.Equal(t, framesBefore, a.framesSinceStart.Load(),
		"tick after Stop must not advance the frame")
}

// TestForeignIDStillDropped verifies that the existing ID gate still works
// alongside the generation gate.
func TestForeignIDStillDropped(t *testing.T) {
	t.Parallel()

	a := New(Settings{ID: "test", Size: 5})
	cmd := a.Start()
	msg := cmd().(StepMsg)

	framesBefore := a.framesSinceStart.Load()
	next := a.Animate(StepMsg{ID: "other", Gen: msg.Gen})
	require.Nil(t, next)
	require.Equal(t, framesBefore, a.framesSinceStart.Load(),
		"foreign ID must not advance the frame")
}

// TestStepMsgCarriesGeneration verifies that Step() stamps the current
// generation into the emitted StepMsg.
func TestStepMsgCarriesGeneration(t *testing.T) {
	t.Parallel()

	a := New(Settings{ID: "test", Size: 5})
	require.Equal(t, int64(0), a.gen.Load(), "fresh Anim starts at generation 0")

	cmd := a.Step()
	msg := cmd().(StepMsg)
	require.Equal(t, int64(0), msg.Gen, "Step before Start carries gen 0")

	a.Start()
	cmd = a.Step()
	msg = cmd().(StepMsg)
	require.Equal(t, int64(1), msg.Gen, "Step after Start carries gen 1")
}

// TestAnimateSchedulesNextStep verifies the normal happy path: a matching
// tick advances the frame and schedules the next step.
func TestAnimateSchedulesNextStep(t *testing.T) {
	t.Parallel()

	a := New(Settings{ID: "test", Size: 5})
	cmd := a.Start()
	msg := cmd().(StepMsg)

	next := a.Animate(msg)
	require.NotNil(t, next, "matching tick must schedule the next step")

	// The next command should also produce a StepMsg with the same gen.
	nextMsg := next().(StepMsg)
	require.Equal(t, msg.Gen, nextMsg.Gen, "chained tick must carry the same generation")
	require.Equal(t, msg.ID, nextMsg.ID)
}

// TestSingleChainAdvances verifies that a single Start() produces a
// working chain that advances frames on each tick.
func TestSingleChainAdvances(t *testing.T) {
	t.Parallel()

	a := New(Settings{ID: "test", Size: 5})
	cmd := a.Start()
	require.NotNil(t, cmd)

	msg := cmd().(StepMsg)
	require.Equal(t, "test", msg.ID)
	require.Equal(t, int64(1), msg.Gen)

	// Advance a few frames.
	for range 5 {
		next := a.Animate(msg)
		require.NotNil(t, next)
		msg = next().(StepMsg)
	}
	require.Equal(t, int64(5), a.framesSinceStart.Load())
}

// TestConcurrentStartDoesNotDoubleAdvance simulates the bug scenario:
// two Start() calls produce two chains, and both chains' ticks arrive.
// Only the second chain's ticks should advance the frame.
func TestConcurrentStartDoesNotDoubleAdvance(t *testing.T) {
	t.Parallel()

	a := New(Settings{ID: "test", Size: 5})

	cmd1 := a.Start()
	cmd2 := a.Start()

	msg1 := cmd1().(StepMsg)
	msg2 := cmd2().(StepMsg)

	// Interleave ticks from both chains.
	frames := int64(0)
	for range 10 {
		if next := a.Animate(msg1); next != nil {
			frames++
			msg1 = next().(StepMsg)
		}
		if next := a.Animate(msg2); next != nil {
			frames++
			msg2 = next().(StepMsg)
		}
	}

	// Only chain 2 should have advanced. Chain 1's ticks are all dropped
	// after the first Start() supersedes it.
	require.Equal(t, int64(10), a.framesSinceStart.Load(),
		"only the latest chain should advance the frame")
	require.Equal(t, frames, a.framesSinceStart.Load())
}

// TestMultipleAnimsIndependent verifies that two Anim instances with
// different IDs don't interfere with each other.
func TestMultipleAnimsIndependent(t *testing.T) {
	t.Parallel()

	a1 := New(Settings{ID: "a1", Size: 5})
	a2 := New(Settings{ID: "a2", Size: 5})

	cmd1 := a1.Start()
	cmd2 := a2.Start()

	msg1 := cmd1().(StepMsg)
	msg2 := cmd2().(StepMsg)

	// Advance a1 only.
	a1.Animate(msg1)
	require.Equal(t, int64(1), a1.framesSinceStart.Load())
	require.Equal(t, int64(0), a2.framesSinceStart.Load())

	// Advance a2 only.
	a2.Animate(msg2)
	require.Equal(t, int64(1), a1.framesSinceStart.Load())
	require.Equal(t, int64(1), a2.framesSinceStart.Load())

	// Cross-talk: a1's tick must not advance a2.
	a2.Animate(msg1)
	require.Equal(t, int64(1), a2.framesSinceStart.Load())
}

// TestStopThenStart verifies that Stop() followed by Start() produces a
// working chain with a fresh generation.
func TestStopThenStart(t *testing.T) {
	t.Parallel()

	a := New(Settings{ID: "test", Size: 5})
	cmd := a.Start()
	msg := cmd().(StepMsg)

	a.Stop()
	genAfterStop := a.gen.Load()

	cmd = a.Start()
	msg2 := cmd().(StepMsg)
	require.Equal(t, genAfterStop+1, msg2.Gen)

	// Old tick must be dropped.
	require.Nil(t, a.Animate(msg))

	// New tick must work.
	require.NotNil(t, a.Animate(msg2))
}

// TestAnimateWithoutStart verifies that Animate works on a fresh Anim
// whose generation is still zero, matching a zero-gen StepMsg.
func TestAnimateWithoutStart(t *testing.T) {
	t.Parallel()

	a := New(Settings{ID: "test", Size: 5})
	msg := StepMsg{ID: "test", Gen: 0}
	next := a.Animate(msg)
	require.NotNil(t, next, "matching gen-0 tick must advance a fresh Anim")
}

// TestDenseBeaverFramesKeepCanonicalFootprint verifies that every dense beaver
// frame (all keys in both Alpha and Beta) renders at exactly 13x5 cells so the
// homescreen hero and its click hitbox never jump between poses.
func TestDenseBeaverFramesKeepCanonicalFootprint(t *testing.T) {
	t.Parallel()

	t.Run("alpha keys", func(t *testing.T) {
		t.Parallel()
		for key := range boot.BeaverFramesDenseAlpha {
			rendered := BeaverFrame(BeaverFrameKey(key), false)
			if got := lipgloss.Width(rendered); got != 13 {
				t.Errorf("frame %q width = %d, want 13", key, got)
			}
			if got := lipgloss.Height(rendered); got != 5 {
				t.Errorf("frame %q height = %d, want 5", key, got)
			}
		}
	})

	t.Run("beta keys", func(t *testing.T) {
		t.Parallel()
		for key := range boot.BeaverFramesDenseBeta {
			rendered := BeaverFrame(BeaverFrameKey(key), true)
			if got := lipgloss.Width(rendered); got != 13 {
				t.Errorf("frame %q width = %d, want 13", key, got)
			}
			if got := lipgloss.Height(rendered); got != 5 {
				t.Errorf("frame %q height = %d, want 5", key, got)
			}
		}
	})

	// BootSequence only references keys that actually exist (either map).
	for _, key := range boot.BootSequence {
		if _, ok := boot.BeaverFramesDenseAlpha[key]; !ok {
			if _, ok := boot.BeaverFramesDenseBeta[key]; !ok {
				t.Errorf("BootSequence references unknown key %q", key)
			}
		}
	}
}

// TestBeaverFrameKeySelection verifies key lookup in the correct map and the
// unknown-key fallback.
func TestBeaverFrameKeySelection(t *testing.T) {
	t.Parallel()

	// Alpha: normal-expression frames render without the errored flag.
	for _, key := range []string{
		string(BeaverFrameKey("center")), string(BeaverFrameKey("left-dir")),
		string(BeaverFrameKey("right-dir")), string(BeaverFrameKey("what")),
	} {
		rendered := BeaverFrame(BeaverFrameKey(key), false)
		if got := lipgloss.Width(rendered); got != 13 {
			t.Errorf("Alpha key %q width = %d, want 13", key, got)
		}
		if rendered == "" {
			t.Errorf("Alpha key %q rendered empty", key)
		}
	}

	// Beta: x-ray frames only render with the errored flag.
	for _, key := range []string{
		string(BeaverFrameKey("error")), string(BeaverFrameKey("confused")),
	} {
		rendered := BeaverFrame(BeaverFrameKey(key), true)
		if got := lipgloss.Width(rendered); got != 13 {
			t.Errorf("Beta key %q width = %d, want 13", key, got)
		}
		if rendered == "" {
			t.Errorf("Beta key %q rendered empty", key)
		}
	}

	// Unknown keys fall back to the Alpha "center" pose.
	fallback := BeaverFrame(BeaverFrameKey("nonexistent"), false)
	want := BeaverFrame(BeaverFrameKey("center"), false)
	if fallback != want {
		t.Errorf("unknown key fell back to %q, want %q", fallback, want)
	}
}

// TestFrameFromDir verifies the horizontal-direction helper used to select
// homescreen beaver direction frames from a -1/0/+1 direction value.
func TestFrameFromDir(t *testing.T) {
	t.Parallel()

	require.Equal(t, FrameLeftDir, FrameFromDir(-1))
	require.Equal(t, FrameRightDir, FrameFromDir(1))
	require.Equal(t, FrameCenter, FrameFromDir(0))
	require.Equal(t, FrameCenter, FrameFromDir(2))
	require.Equal(t, FrameCenter, FrameFromDir(-42))
}
