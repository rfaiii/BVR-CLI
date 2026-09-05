package model

import (
	"image"
	"image/color"
	"math/rand/v2"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"

	"github.com/richavery/bvr-cli/internal/version"
)

// Boot-time version banner.
//
// A one-line bar on the bottom of the UI that plays a short brand sequence at
// startup, scrambling each element's letters into place, holding it for two
// seconds, then advancing to the next:
//
//	1. "BVR"
//	2. "v1.1.6:beta_v3"   (the app version)
//	3. "OH BEAV!"
//	4. "created by RICHARD AIZEN AVERY III"
//
// The bar wipes in at boot and fades out at the end. The bar uses the theme's
// ACCENT color for its background and the app's BACKGROUND color for its text
// so the brand "pops". It is drawn last so it overlays the status bar, and it
// only runs once per boot.

const (
	versionBannerFPS         = 20
	versionBannerTick        = time.Second / versionBannerFPS
	versionBannerInFrames    = 10 // ~0.5s wipe-in
	versionBannerOutFrames   = 10 // ~0.5s fade-out
	versionBannerRevealChunk = 2  // frames between each revealed char
	versionBannerHoldFrames  = 40 // 2s hold at 20fps
)

// bannerGlyphs is the scramble alphabet, matching the anim package's runes.
var bannerGlyphs = []rune("0123456789abcdefABCDEF~!@#$%^&*()+=_")

type versionBannerPhase int

const (
	bannerPhaseHidden versionBannerPhase = iota
	bannerPhaseIn
	bannerPhaseReveal
	bannerPhaseHold
	bannerPhaseOut
	bannerPhaseDone
)

// versionBannerTickMsg drives the banner animation loop.
type versionBannerTickMsg struct{}

type versionBanner struct {
	phase    versionBannerPhase
	frame    int
	elements []string
	idx      int
	random   *rand.Rand
}

// bannerElements returns the boot banner sequence. Element 0 is strictly the
// brand wordmark "BVR"; the version is the current build, an attribution
// line credits the author, and "OH BEAV!" is the BVR refrain.
func bannerElements() []string {
	return []string{
		"BVR",
		"v" + version.ShortVersion(),
		"OH BEAV!",
		"created by RICHARD AIZEN AVERY III",
	}
}

func newVersionBanner() *versionBanner {
	return &versionBanner{
		phase:    bannerPhaseHidden,
		elements: bannerElements(),
		random:   rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0xB56CFF)),
	}
}

// start begins the boot banner sequence. It is a no-op after the first run.
func (b *versionBanner) start() tea.Cmd {
	if b.phase != bannerPhaseHidden {
		return nil
	}
	b.phase = bannerPhaseIn
	b.frame = 0
	return bannerTick()
}

func bannerTick() tea.Cmd {
	return tea.Tick(versionBannerTick, func(time.Time) tea.Msg {
		return versionBannerTickMsg{}
	})
}

// advance moves the state machine forward one frame and returns the next
// tick command, or nil once the banner has finished.
func (b *versionBanner) advance() tea.Cmd {
	switch b.phase {
	case bannerPhaseIn:
		b.frame++
		if b.frame >= versionBannerInFrames {
			b.phase = bannerPhaseReveal
			b.frame = 0
		}
	case bannerPhaseReveal:
		b.frame++
		if b.revealed() >= len(b.elements[b.idx]) {
			b.phase = bannerPhaseHold
			b.frame = 0
		}
	case bannerPhaseHold:
		b.frame++
		if b.frame >= versionBannerHoldFrames {
			if b.idx < len(b.elements)-1 {
				b.idx++
				b.phase = bannerPhaseReveal
				b.frame = 0
			} else {
				b.phase = bannerPhaseOut
				b.frame = 0
			}
		}
	case bannerPhaseOut:
		b.frame++
		if b.frame >= versionBannerOutFrames {
			b.phase = bannerPhaseDone
			return nil
		}
	default:
		return nil
	}
	return bannerTick()
}

// revealed returns how many leading characters are settled (non-cycling).
func (b *versionBanner) revealed() int {
	// Reveal every character of the current element before advancing to the
	// hold phase. A previous fixed cap (versionBannerRevealMax) stopped the
	// reveal short of any string longer than 24 runes — e.g. the author
	// attribution — so those elements kept scrambling forever and the banner
	// never faded. The reveal pace is already governed by
	// versionBannerRevealChunk, so no separate cap is needed here.
	return min(b.frame/versionBannerRevealChunk, len(b.elements[b.idx]))
}

// active reports whether the banner is currently on screen.
func (b *versionBanner) active() bool {
	switch b.phase {
	case bannerPhaseIn, bannerPhaseReveal, bannerPhaseHold, bannerPhaseOut:
		return true
	default:
		return false
	}
}

// progress returns the bar width progress (0..1) for the in/out phases.
func (b *versionBanner) progress() float64 {
	switch b.phase {
	case bannerPhaseIn:
		return float64(b.frame) / float64(versionBannerInFrames)
	case bannerPhaseOut:
		return 1 - float64(b.frame)/float64(versionBannerOutFrames)
	default:
		return 1
	}
}

// renderText returns the text shown inside the bar for the current frame.
func (b *versionBanner) renderText() string {
	switch b.phase {
	case bannerPhaseReveal:
		revealed := b.revealed()
		targetRunes := []rune(b.elements[b.idx])
		var sb strings.Builder
		for i, r := range targetRunes {
			switch {
			case r == ' ':
				sb.WriteRune(' ')
			case i < revealed:
				sb.WriteRune(r)
			default:
				sb.WriteRune(bannerGlyphs[b.random.IntN(len(bannerGlyphs))])
			}
		}
		return sb.String()
	case bannerPhaseHold:
		return b.elements[b.idx]
	default:
		return ""
	}
}

// draw overlays the banner on the bottom row of the screen. The bar uses the
// theme accent color for its background and the app background color for its
// text so the brand "pops".
func (b *versionBanner) draw(scr uv.Screen, area uv.Rectangle, accent, bg color.Color) {
	if b == nil || !b.active() || area.Dx() <= 0 {
		return
	}

	width := area.Dx()
	barWidth := max(1, int(b.progress()*float64(width)))
	xoff := (width - barWidth) / 2

	// During text phases the bar spans full width (grows from center on in/out).
	if b.phase == bannerPhaseReveal || b.phase == bannerPhaseHold {
		barWidth = width
		xoff = 0
	}

	text := b.renderText()
	inner := max(0, barWidth-2) // 1-cell padding each side
	if runes := []rune(text); len(runes) > inner {
		text = string(runes[:inner])
	}

	style := lipgloss.NewStyle().
		Background(accent).
		Foreground(bg).
		Bold(true)

	var sb strings.Builder
	sb.WriteString(strings.Repeat(" ", xoff))
	sb.WriteString(style.Render(" " + text + strings.Repeat(" ", max(0, inner-len([]rune(text)))) + " "))
	sb.WriteString(strings.Repeat(" ", max(0, width-xoff-barWidth)))

	rect := image.Rectangle{
		Min: image.Pt(area.Min.X, area.Max.Y-1),
		Max: image.Pt(area.Max.X, area.Max.Y),
	}
	uv.NewStyledString(sb.String()).Draw(scr, rect)
}
