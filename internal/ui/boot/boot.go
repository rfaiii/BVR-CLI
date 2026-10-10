// Package boot renders the animated BVR-CLI beaver intro sequence before the
// Bubble Tea main loop takes over the terminal.
package boot

import (
	"fmt"
	"time"

	"charm.land/lipgloss/v2"
)

// BeaverFramesDenseAlpha are the dense mascot frames for the normal state
// (0_0 / -0_0- eyes), reused by the boot splash and the homescreen hero. Each
// frame is exactly 13 cells wide and 5 rows tall so the terminal grid never
// collapses while the mascot reacts:
//
//   - center: resting, neutral 0_0 face.
//   - left-dir: keys or selected text are to the LEFT of center.
//   - right-dir: keys or selected text are to the RIGHT of center.
//   - what: the user clicked the beaver mascot.
var BeaverFramesDenseAlpha = map[string]string{
	"center": `  (\.-./)  
 /   0_0   \ 
|   -(v)-   |
 \   [=]   / 
  '-------'  `,
	"left-dir": `  (\.-./)  
 / X_X     \ 
| \X(v)/    |
 \ [X]     / 
  '-------'  `,
	"right-dir": `  (\.-./)  
 /     X_X \ 
|     \X(v)/|
 \     [X] / 
  '-------'  `,
	"what": `  (\.-./)  
 /  -0_0-  \ 
|  \\(o)//  |
 \  ((0))   /
  '-------'  `,
}

var BeaverFramesDenseBeta = map[string]string{
	"error": `  (\.-./)  
 /   X_X   \ 
|   \X(v)/  |
 \   [=]   / 
  '-------'  `,
	"confused": `  (\.-./)  
 /  -9_g-  \ 
|   >(%)<   |
 \   [*}   / 
  '-------'  `,
}

// BootSequence is the frame order for the intro animation.
var BootSequence = []string{"center", "left-dir", "center", "right-dir", "center"}

var MascotColor = lipgloss.Color("#39f66b")

// LoadingColor is a secondary accent used for the loading line beneath the
// mascot.
var LoadingColor = lipgloss.Color("#b972ff")

var mascotStyle = lipgloss.NewStyle().Foreground(MascotColor).Bold(true)
var loadingStyle = lipgloss.NewStyle().Foreground(LoadingColor)

// Run plays the beaver boot animation. It clears the terminal first, renders
// each mascot frame in neon green with a secondary-accent loading line beneath,
// then returns so the caller can start the Bubble Tea main event loop.
//
// Run is a no-op guard: callers should only invoke it from an interactive
// terminal (see IsInteractive) and should clear any prior program output first.
func Run() {
	// Clear the screen and home the cursor so the mascot starts top-left.
	fmt.Print("\033[H\033[2J")
	for _, frame := range BootSequence {
		fmt.Print("\033[0;0H") // cursor to top-left
		fmt.Println(mascotStyle.Render(BeaverFramesDenseAlpha[frame]))
		fmt.Println()
		fmt.Println(loadingStyle.Render("Initializing bvr-cli components..."))
		time.Sleep(400 * time.Millisecond)
	}
}
