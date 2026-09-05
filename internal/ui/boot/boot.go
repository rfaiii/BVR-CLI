// Package boot renders the animated BVR-CLI beaver intro sequence before the
// Bubble Tea main loop takes over the terminal.
package boot

import (
	"fmt"
	"time"

	"charm.land/lipgloss/v2"
)

// BeaverFramesDenseAlpha are the dense, character-filled mascot frames for the
// normal (0_0 eyes) state. Filler characters (A, V, @@, H, AW, WX) shift behind
// the facial features to keep the 13-cell monospaced bounding box stable across
// frames while the mascot rotates.
var BeaverFramesDenseAlpha = map[string]string{
	"center": `  (\.---./)  
 /A  0_0  V\ 
| @@-(v)- H |
 \AW [=] WX/ 
  '-------'  `,
	"left": `  (\.---./)  
 / 0_0  A V\ 
| -(v)- @@H |
 \ [=] AWWX/ 
  '-------'  `,
	"right": `  (\.---./)  
 /A V  0_0 \ 
| @@H -(v)- |
 \AWWX [=] / 
  '-------'  `,
}

// BeaverFramesDenseBeta is the x-ray ("dead") mascot variant (X_X eyes), shown
// when the agent errors. Same 13x5 bounding box; only the fillers change (to
// @R/JU/HI/YH) so the face stays textured.
var BeaverFramesDenseBeta = map[string]string{
	"center": `  (\.---./)  
 /A  X_X AX\ 
| @R-(v)-JU |
 \HI [=] YH/ 
  '-------'  `,
	"left": `  (\.---./)  
 / X_X A AX\ 
| -(v)-@RJU |
 \ [=] HIYH/ 
  '-------'  `,
	"right": `  (\.---./)  
 /A AX X_X \ 
| @RJU-(v)- |
 \HIYH [=] / 
  '-------'  `,
}

// BootSequence is the frame order for the intro animation.
var BootSequence = []string{"center", "left", "center", "right", "center"}

// MascotColor is the BVR neon green (#39f66b) used for the beaver so it pops
// vibrantly against the dark terminal background.
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
