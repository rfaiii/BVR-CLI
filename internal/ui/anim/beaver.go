// Package anim provides small terminal animation helpers for BVR-CLI.
// In addition to the cursor trail, it hosts the idle beaver mascot that
// lives on the homescreen and reuses the boot splash frames.
package anim

import (
	"charm.land/lipgloss/v2"

	"github.com/richavery/bvr-cli/internal/ui/boot"
)

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
