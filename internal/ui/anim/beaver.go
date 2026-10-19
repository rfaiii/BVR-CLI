// Package anim provides small terminal animation helpers for BVR-CLI.
// In addition to the cursor trail, it hosts the idle beaver mascot that
// lives on the homescreen and reuses the boot splash frames.
package anim

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/richavery/bvr-cli/internal/ui/boot"
)

// MascotState is the shared state vocabulary for the large and mini Beaver.
// Passive gaze states are followed by transient expression states so the UI
// can add reactions without replacing the character renderer.
type MascotState int

const (
	StateCenter MascotState = iota
	StateLookLeft
	StateLookRight
	StateLookUp
	StateLookDown
	StateLaugh
	StateShocked
	StateDanceA
	StateDanceB
	StateChompA
	StateChompB
	StateClickBoop
	StateMusicA
	StateMusicB
)

const (
	LargeMascotWidth  = 24
	LargeMascotHeight = 8
	MiniMascotWidth   = 12
	MiniMascotHeight  = 7
)

// EyeDirection maps mascot-center-relative pointer coordinates into the five
// stable gaze buckets described by the mascot handoff. The dead zone prevents
// eye twitching when the pointer is close to the mascot's face.
func EyeDirection(dx, dy int) MascotState {
	const deadX = 3
	const deadY = 2
	if abs(dx) <= deadX && abs(dy) <= deadY {
		return StateCenter
	}
	if abs(dx) > abs(dy) {
		if dx < 0 {
			return StateLookLeft
		}
		return StateLookRight
	}
	if dy < 0 {
		return StateLookUp
	}
	return StateLookDown
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

var largeMascotFrames = map[MascotState][]string{
	StateCenter: {
		"       .-^^^^-.        ",
		"     .'  .--.  '.      ",
		"    /   /X  X\\   \\",
		"   |      v       |==  ",
		"   |    .---.     |==  ",
		"   |    | | |     |    ",
		"    \\    \\_/    /     ",
		"     '._       _.'      ",
	},
	StateLookLeft: {
		"       .-^^^^-.        ",
		"     .'  .--.  '.      ",
		"    /   /o  o\\   \\",
		"   |      v       |==  ",
		"   |    .---.     |==  ",
		"   |    | | |     |    ",
		"    \\    \\_/    /     ",
		"     '._       _.'      ",
	},
	StateLookRight: {
		"       .-^^^^-.        ",
		"     .'  .--.  '.      ",
		"    /   /  o  o\\   \\",
		"   |      v       |==  ",
		"   |    .---.     |==  ",
		"   |    | | |     |    ",
		"    \\    \\_/    /     ",
		"     '._       _.'      ",
	},
	StateLookUp: {
		"       .-^^^^-.        ",
		"     .'  .--.  '.      ",
		"    /   /^  ^\\   \\",
		"   |      v       |==  ",
		"   |    .---.     |==  ",
		"   |    | | |     |    ",
		"    \\    \\_/    /     ",
		"     '._       _.'      ",
	},
	StateLookDown: {
		"       .-^^^^-.        ",
		"     .'  .--.  '.      ",
		"    /   /-  -\\   \\",
		"   |      v       |==  ",
		"   |    .---.     |==  ",
		"   |    | | |     |    ",
		"    \\    \\_/    /     ",
		"     '._       _.'      ",
	},
	StateLaugh: {
		"       .-^^^^-.        ",
		"     .'  .--.  '.      ",
		"    /   /^  ^\\   \\",
		"   |      v       |==  ",
		"   |    .---.     |==  ",
		"   |   / HA! \\    |    ",
		"    \\  \\___/    /     ",
		"     '._       _.'      ",
	},
	StateShocked: {
		"       .-^^^^-.        ",
		"     .'  .--.  '.      ",
		"    /   /O  O\\   \\",
		"   |      !       |==  ",
		"   |     ___      |==  ",
		"   |    / O \\     |    ",
		"    \\   \\___/   /     ",
		"     '._       _.'      ",
	},
	StateDanceA: {
		"       .-^^^^-.        ",
		"     .'  .--.  '.      ",
		"    /   /X  X\\   \\",
		"   |      v       |==  ",
		"   |    .---.     |==  ",
		"   |    | | |     |    ",
		"    \\   /_\\    /     ",
		"     '._       _.'      ",
	},
	StateDanceB: {
		"       .-^^^^-.        ",
		"     .'  .--.  '.      ",
		"    /   /X  X\\   \\",
		"   |      v       |==  ",
		"   |    .---.     |==  ",
		"   |    | | |     |    ",
		"    \\  \\_/     /     ",
		"     '._       _.'      ",
	},
	StateChompA: {
		"       .-^^^^-.        ",
		"     .'  .--.  '.      ",
		"    /   />  <\\   \\",
		"   |      v       |==  ",
		"   |   .-WWW-.    |==  ",
		"   |   |___  |    |==  ",
		"    \\  \\_/ /   / WOOD",
		"     '._   _.'         ",
	},
	StateChompB: {
		"       .-^^^^-.        ",
		"     .'  .--.  '.      ",
		"    /   />  <\\   \\",
		"   |      v       |==  ",
		"   |   .-VVV-.    |==  ",
		"   |   |___  |    |==  ",
		"    \\  \\_/ /   / WOO ",
		"     '._   _.'         ",
	},
	StateClickBoop: {
		"       .-^^^^-.        ",
		"     .'  .--.  '.      ",
		"    /   /@  @\\   \\",
		"   |      v       |==  ",
		"   |   .-------.   |== ",
		"   |   | BOOP! |   |    ",
		"    \\  '-----'   /     ",
		"     '._       _.'      ",
	},
}

var miniMascotFrames = map[MascotState][]string{
	StateCenter:    {"  .--.     ", " /X  X\\  = ", "|  v  |  ==", "| | | |    ", " \\_/_/     ", "/|_|_|\\    ", "  ====     "},
	StateLookLeft:  {"  .--.     ", " /o  o\\  = ", "|  v  |  ==", "| | | |    ", " \\_/_/     ", "/|_|_|\\    ", "  ====     "},
	StateLookRight: {"  .--.     ", " /  o o\\ = ", "|  v  |  ==", "| | | |    ", " \\_/_/     ", "/|_|_|\\    ", "  ====     "},
	StateLookUp:    {"  .--.     ", " /^  ^\\  = ", "|  v  |  ==", "| | | |    ", " \\_/_/     ", "/|_|_|\\    ", "  ====     "},
	StateLookDown:  {"  .--.     ", " /-  -\\  = ", "|  v  |  ==", "| | | |    ", " \\_/_/     ", "/|_|_|\\    ", "  ====     "},
	StateLaugh:     {"  .--.     ", " /^  ^\\  = ", "|  v  |  ==", "| HA! |    ", " \\___/     ", "/|_|_|\\    ", "  ====     "},
	StateShocked:   {"  .--.     ", " /O  O\\  = ", "|  !  |  ==", "|  O  |    ", " \\___/     ", "/|_|_|\\    ", "  ====     "},
	StateDanceA:    {"  .--.     ", " /X  X\\  = ", "|  v  |  ==", "| | | |    ", " \\_/_/     ", "/|_|_|\\    ", " /====\\    "},
	StateDanceB:    {"  .--.     ", " /X  X\\  = ", "|  v  |  ==", "| | | |    ", " \\_/_/     ", "/|_|_|\\    ", " \\====/    "},
	StateChompA:    {"  .--.     ", " />  <\\  = ", "|  v  |  ==", "| WWW |    ", " \\_/_/WOOD ", "/|_|_|\\    ", "  ====     "},
	StateChompB:    {"  .--.     ", " />  <\\  = ", "|  v  |  ==", "| VVV |    ", " \\_/_/WOO  ", "/|_|_|\\    ", "  ====     "},
	StateClickBoop: {"  .--.     ", " /@  @\\  = ", "| BOOP!| ==", "|  v  |    ", " \\_/_/     ", "/|_|_|\\    ", "  ====     "},
}

// MascotFrame returns a fixed-size ASCII frame for the requested mascot.
// Unknown or music states fall back to the nearest stable expression.
func MascotFrame(state MascotState, frame int, mini, errored bool) string {
	frames := largeMascotFrames
	width, height := LargeMascotWidth, LargeMascotHeight
	if mini {
		frames = miniMascotFrames
		width, height = MiniMascotWidth, MiniMascotHeight
	}
	if state == StateMusicA || state == StateMusicB {
		state = StateCenter
	}
	poses, ok := frames[state]
	if !ok {
		poses = frames[StateCenter]
	}
	// Each state is stored as one complete fixed-size frame: the slice holds
	// rows, not alternate poses. Direction and expression changes select the
	// state; the central animation tick still drives the surrounding UI.
	_ = frame
	pose := strings.Join(poses, "\n")
	if errored {
		pose = strings.ReplaceAll(pose, "X", "!")
	}
	return renderMascotLines(normalizeMascotFrame(pose, width, height))
}

// renderMascotLines applies terminal styling one row at a time. The screen
// buffer treats a styled multiline string as one write in some terminals,
// which can collapse the mascot into a single garbled line. Keeping the ANSI
// span local to each row preserves the fixed frame height everywhere.
func renderMascotLines(frame string) string {
	lines := strings.Split(frame, "\n")
	for i, line := range lines {
		lines[i] = mascotStyle.Render(line)
	}
	return strings.Join(lines, "\n")
}

func normalizeMascotFrame(pose string, width, height int) string {
	lines := strings.Split(pose, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	for i, line := range lines {
		if lipgloss.Width(line) > width {
			line = truncateCells(line, width)
		}
		lines[i] = line + strings.Repeat(" ", width-lipgloss.Width(line))
	}
	return strings.Join(lines, "\n")
}

func truncateCells(s string, width int) string {
	for lipgloss.Width(s) > width {
		runes := []rune(s)
		s = string(runes[:len(runes)-1])
	}
	return s
}

// LargeMascotFrame and MiniMascotFrame are named helpers for call sites that
// want to make the reserved footprint explicit.
func LargeMascotFrame(state MascotState, frame int, errored bool) string {
	return MascotFrame(state, frame, false, errored)
}

func MiniMascotFrame(state MascotState, frame int, errored bool) string {
	return MascotFrame(state, frame, true, errored)
}

// BeaverHeroFrames are the larger, Cline-inspired mascot poses used in the
// landing page and chat sidebar. The fixed-width frames make the animation
// feel intentional instead of making the layout jump as the character moves.
var BeaverHeroFrames = []string{
	`          .-^^^^-.
       .-'  0  0  '-.
      /       \\_/     \\
     |    .--------.    |====
     |   /   BVR    \\   |====
     |   \\_________/   |====
      \\      V       _/
       '._        _.-'
          /  /\\  \\`,
	`          .-^^^^-.
       .-'  -  -  '-.
      /       \\_/     \\
     |    .--------.    |====
     |   /   BVR    \\   |====
     |   \\_________/   |====
      \\      V       _/
       '._        _.-'
          /  /\\  \\`,
	`          .-^^^^-.
       .-'  0  0  '-.
      /      _\\_/     \\
     |    .--------.    |====
     |   /  WORKING  \\   |====
     |   \\_________/   |====
      \\      V       _/
       '._        _.-'
          /  /\\  \\`,
	`           .-^^^^-.
        .-'  0  0  '-.
       /       \\_/    \\
      |    .--------.   |====
      |   /   BVR    \\  |====
      |   \\_________/  |====
       \\      V      _/
        '._       _.-'
           / /\\ \\`,
}

// BeaverHeroFrame returns a stable-width hero pose. Holding each pose for a
// few animation ticks gives the mascot the deliberate rhythm seen in modern
// terminal AI interfaces while remaining almost free to render.
func BeaverHeroFrame(frame int, errored bool) string {
	if len(BeaverHeroFrames) == 0 {
		return ""
	}
	pose := BeaverHeroFrames[(frame/3)%len(BeaverHeroFrames)]
	if errored {
		pose = strings.ReplaceAll(pose, "0  0", "X  X")
	}
	return renderMascotLines(normalizeHeroFrame(pose))
}

func normalizeHeroFrame(pose string) string {
	lines := strings.Split(pose, "\n")
	const heroWidth = 31
	width := heroWidth
	for _, line := range lines {
		width = max(width, lipgloss.Width(line))
	}
	for i, line := range lines {
		lines[i] = line + strings.Repeat(" ", width-lipgloss.Width(line))
	}
	return strings.Join(lines, "\n")
}

// mascotStyle renders the beaver mascot in BVR neon green (#39f66b) with bold
// weight, matching the boot splash treatment so the homescreen mascot looks
// identical to the intro.
var mascotStyle = lipgloss.NewStyle().Foreground(boot.MascotColor).Bold(true)

// beaverFrameKey selects the dense beaver pose to render.
//
//   - "center": resting, neutral 0_0 face.
//   - "left-dir": keys or selected text are to the LEFT of center.
//   - "right-dir": keys or selected text are to the RIGHT of center.
//   - "what": the user clicked the beaver mascot.
//   - "error": the agent errored (X_X x-ray eyes).
//   - "confused": the agent is loading or thinking.
//
// Every dense frame is 13×5 cells: a 7-cell head (\\.-./), 13-wide body rows,
// and a centered 9-cell base. normalizeDenseFrame pads short rows to exactly
// 13 so the grid never collapses while the mascot reacts.
type BeaverFrameKey string

const (
	FrameCenter   BeaverFrameKey = "center"
	FrameLeftDir  BeaverFrameKey = "left-dir"
	FrameRightDir BeaverFrameKey = "right-dir"
	FrameWhat     BeaverFrameKey = "what"
	FrameError    BeaverFrameKey = "error"
	FrameConfused BeaverFrameKey = "confused"
)

// The dense BVR mascot is a fixed 13x5 grid: a 5-cell head, 13-wide body rows,
// and a 5-row base. Every pose is padded to this footprint so the layout and
// the click hitbox stay identical across the whole animation.
const (
	DenseFrameWidth  = 13
	DenseFrameHeight = 5
)

// FrameFromDir maps a horizontal direction into a dense mascot frame key: -1
// for left-dir, 1 for right-dir, and center otherwise.
func FrameFromDir(dir int) BeaverFrameKey {
	switch dir {
	case -1:
		return FrameLeftDir
	case 1:
		return FrameRightDir
	default:
		return FrameCenter
	}
}

// BeaverFrame returns the BVR mascot for the given frame key.
//
// key: the expression to show (see beaverFrameKey constants).
// errored: when true, the Beta (x-ray) map is preferred so error and confused
// poses win; otherwise the Alpha (normal) map is preferred.
//
// The two maps hold different poses: Alpha carries center/left-dir/right-dir/
// what, Beta carries error/confused. Looking in a single map made "confused"
// unreachable whenever errored was false, so the agent-busy expression silently
// degraded to the resting face. Both maps are consulted, in a priority order,
// before falling back to the resting pose.
func BeaverFrame(key BeaverFrameKey, errored bool) string {
	if frame := lookupBeaverFrame(string(key), errored); frame != "" {
		return mascotStyle.Render(normalizeDenseFrame(frame))
	}
	return mascotStyle.Render(normalizeDenseFrame(boot.BeaverFramesDenseAlpha[string(FrameCenter)]))
}

// lookupBeaverFrame resolves a pose name across the Alpha and Beta frame maps.
// When errored is set the Beta (x-ray) poses take precedence, otherwise the
// Alpha poses do. Each map is still consulted as a fallback so a pose that only
// exists in one map stays reachable from either state.
func lookupBeaverFrame(name string, errored bool) string {
	if errored {
		if frame := boot.BeaverFramesDenseBeta[name]; frame != "" {
			return frame
		}
		return boot.BeaverFramesDenseAlpha[name]
	}
	if frame := boot.BeaverFramesDenseAlpha[name]; frame != "" {
		return frame
	}
	return boot.BeaverFramesDenseBeta[name]
}

// normalizeDenseFrame pads every row of a dense beaver frame to the 13-cell
// monospaced grid so the rendered mascot footprint is always 13x5 regardless
// of the exact spacing in the source strings. The stable footprint keeps the
// surrounding TUI layout (click hitbox, hero position) unchanged across poses.
//
// Every frame in BeaverFramesDenseAlpha/Beta is designed so no row exceeds
// DenseFrameWidth cells; shorter rows are right-padded with spaces.
func normalizeDenseFrame(frame string) string {
	lines := strings.Split(frame, "\n")
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		lw := lipgloss.Width(l)
		if lw < DenseFrameWidth {
			l = l + strings.Repeat(" ", DenseFrameWidth-lw)
		}
		out = append(out, l)
	}
	return strings.Join(out, "\n")
}
