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
	LargeMascotHeight = 12
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
		"    /|___|___|__|\\     ",
		"   /_/  /     \\  \\_   ",
		"  //// /_/   \\_\\ //// ",
		"========================",
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
		"    /|___|___|__|\\     ",
		"   /_/  /     \\  \\_   ",
		"  //// /_/   \\_\\ //// ",
		"========================",
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
		"    /|___|___|__|\\     ",
		"   /_/  /     \\  \\_   ",
		"  //// /_/   \\_\\ //// ",
		"========================",
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
		"    /|___|___|__|\\     ",
		"   /_/  /     \\  \\_   ",
		"  //// /_/   \\_\\ //// ",
		"========================",
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
		"    /|___|___|__|\\     ",
		"   /_/  /     \\  \\_   ",
		"  //// /_/   \\_\\ //// ",
		"========================",
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
		"    /|___|___|__|\\     ",
		"   /_/  /     \\  \\_   ",
		"  //// /_/   \\_\\ //// ",
		"========================",
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
		"    /|___|___|__|\\     ",
		"   /_/  /     \\  \\_   ",
		"  //// /_/   \\_\\ //// ",
		"========================",
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
		"    /|___|___|__|\\     ",
		"   /_/ /     \\  \\_   ",
		"  ////_/     \\_\\ //// ",
		"========================",
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
		"    /|___|___|__|\\     ",
		"   /_/  /     \\  \\_   ",
		"  //// /_/   \\_\\ //// ",
		"========================",
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
		"    /|___|___|__|\\     ",
		"   /_/  /     \\  \\_   ",
		"  //// /_/   \\_\\ //// ",
		"========================",
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
		"    /|___|___|__|\\     ",
		"   /_/  /     \\  \\_   ",
		"  //// /_/   \\_\\ //// ",
		"========================",
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
		"    /|___|___|__|\\     ",
		"   /_/  /     \\  \\_   ",
		"  //// /_/   \\_\\ //// ",
		"========================",
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

// BeaverFrame returns the homescreen beaver mascot, facing the given direction.
//
// facing: -1 = look left, 0 = center (resting), +1 = look right.
// errored: render the x-ray Beta variant (X_X eyes) when the agent has errored.
// resting: render the neutral center-facing "rest" pose between direction changes.
func BeaverFrame(facing int, errored, resting bool) string {
	frames := boot.BeaverFramesDenseAlpha
	if errored {
		frames = boot.BeaverFramesDenseBeta
	}

	if resting {
		return mascotStyle.Render(frames["center"])
	}

	// Face the cursor/prompt based on the debounced facing direction.
	// -1 = left, 0 = center, +1 = right.
	facingStr := "right"
	switch facing {
	case -1:
		facingStr = "left"
	case 0:
		facingStr = "center"
	case 1:
		facingStr = "right"
	}
	frame, ok := frames[facingStr]
	if !ok {
		frame = frames["center"]
	}
	return mascotStyle.Render(frame)
}
