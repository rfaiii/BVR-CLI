// Package bots is the data and demo backend for the /bots picker.
//
// Sprites, voices, and sound cue names live here so a real transport can
// replace Fake later without the TUI growing a second copy of the roster.
package bots

import (
	"image/color"
	"os"
	"strings"

	"github.com/charmbracelet/colorprofile"
)

// ColorMode is how a sprite is painted.
type ColorMode int

const (
	// ColorTrue emits 24-bit ANSI colors.
	ColorTrue ColorMode = iota
	// ColorANSI256 emits the nearest xterm 256-color index.
	ColorANSI256
	// ColorNone emits no color sequences. Shapes stay readable via glyph
	// weight: eyes are solid blocks, bodies are lighter shades.
	ColorNone
)

// Pix is one palette index. Zero is transparent.
type Pix byte

const (
	PixClear Pix = iota
	PixWhite
	PixEye
	PixOrange
	PixStar
	PixGreen
	PixGray
	PixPink
	PixPurple
	PixTile
)

// ModeFromProfile picks a color mode from the terminal profile. NO_COLOR
// and TERM=dumb always win, including over a truecolor report.
func ModeFromProfile(p colorprofile.Profile) ColorMode {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return ColorNone
	}
	switch p {
	case colorprofile.TrueColor:
		return ColorTrue
	case colorprofile.ANSI256, colorprofile.ANSI:
		return ColorANSI256
	case colorprofile.Ascii, colorprofile.NoTTY:
		return ColorNone
	default:
		return modeFromEnv()
	}
}

func modeFromEnv() ColorMode {
	ct := strings.ToLower(os.Getenv("COLORTERM"))
	if strings.Contains(ct, "truecolor") || strings.Contains(ct, "24bit") {
		return ColorTrue
	}
	term := os.Getenv("TERM")
	if strings.Contains(term, "256color") {
		return ColorANSI256
	}
	if term == "" || term == "dumb" {
		return ColorNone
	}
	return ColorANSI256
}

// Audible reports whether bot cues should play. Silent volume and a
// non-empty NO_SOUND environment variable both mute.
func Audible(volume string) bool {
	if os.Getenv("NO_SOUND") != "" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(volume)) {
	case "silent", "off", "mute", "muted":
		return false
	default:
		return true
	}
}

func (p Pix) rgba() (color.RGBA, bool) {
	switch p {
	case PixWhite, PixStar:
		return color.RGBA{R: 0xFC, G: 0xFC, B: 0xFC, A: 0xFF}, true
	case PixEye:
		return color.RGBA{R: 0x14, G: 0x14, B: 0x1C, A: 0xFF}, true
	case PixOrange:
		return color.RGBA{R: 0xF4, G: 0xA2, B: 0x3C, A: 0xFF}, true
	case PixGreen:
		return color.RGBA{R: 0x3D, G: 0xDC, B: 0x6A, A: 0xFF}, true
	case PixGray:
		return color.RGBA{R: 0xA0, G: 0xA0, B: 0xA8, A: 0xFF}, true
	case PixPink:
		return color.RGBA{R: 0xFF, G: 0x4D, B: 0x8D, A: 0xFF}, true
	case PixPurple:
		return color.RGBA{R: 0x7C, G: 0x5C, B: 0xFF, A: 0xFF}, true
	case PixTile:
		return color.RGBA{R: 0x3A, G: 0x3A, B: 0x46, A: 0xFF}, true
	default:
		return color.RGBA{}, false
	}
}

// Paint colors a plain string with a palette entry.
func Paint(text string, p Pix, mode ColorMode) string {
	if text == "" || mode == ColorNone {
		return text
	}
	c, ok := p.rgba()
	if !ok {
		return text
	}
	switch mode {
	case ColorTrue:
		return fgTrue(c) + text + reset
	case ColorANSI256:
		return fg256(nearest256(c)) + text + reset
	default:
		return text
	}
}

const reset = "\x1b[0m"

func fgTrue(c color.RGBA) string {
	return ansiCode(38, 2, int(c.R), int(c.G), int(c.B))
}

func bgTrue(c color.RGBA) string {
	return ansiCode(48, 2, int(c.R), int(c.G), int(c.B))
}

func fg256(n int) string {
	return ansiCode(38, 5, n)
}

func bg256(n int) string {
	return ansiCode(48, 5, n)
}

func ansiCode(parts ...int) string {
	var b strings.Builder
	b.WriteString("\x1b[")
	for i, p := range parts {
		if i > 0 {
			b.WriteByte(';')
		}
		b.WriteString(itoa(p))
	}
	b.WriteByte('m')
	return b.String()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [8]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// nearest256 maps an RGB color onto the xterm 256-color cube or gray ramp.
func nearest256(c color.RGBA) int {
	cr, cg, cb := cubeIndex(c.R), cubeIndex(c.G), cubeIndex(c.B)
	cube := color.RGBA{R: cubeValue(cr), G: cubeValue(cg), B: cubeValue(cb), A: 0xFF}
	gi := grayIndex(c)
	gv := grayValue(gi)
	gray := color.RGBA{R: gv, G: gv, B: gv, A: 0xFF}
	if dist2(c, gray) < dist2(c, cube) {
		return 232 + gi
	}
	return 16 + 36*cr + 6*cg + cb
}

func cubeIndex(v uint8) int {
	best, bestD := 0, 1<<30
	for i, val := range []int{0, 95, 135, 175, 215, 255} {
		d := abs(int(v) - val)
		if d < bestD {
			best, bestD = i, d
		}
	}
	return best
}

func cubeValue(i int) uint8 {
	return [...]uint8{0, 95, 135, 175, 215, 255}[i]
}

func grayIndex(c color.RGBA) int {
	y := (int(c.R)*30 + int(c.G)*59 + int(c.B)*11) / 100
	idx := (y - 8) / 10
	if idx < 0 {
		return 0
	}
	if idx > 23 {
		return 23
	}
	return idx
}

func grayValue(i int) uint8 {
	return uint8(8 + i*10)
}

func dist2(a, b color.RGBA) int {
	dr := int(a.R) - int(b.R)
	dg := int(a.G) - int(b.G)
	db := int(a.B) - int(b.B)
	return dr*dr + dg*dg + db*db
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
