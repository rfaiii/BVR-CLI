package bots

import (
	"image/color"
	"strings"
)

// SpriteW and SpriteH are the pixel size of every bot. Half-block cells
// fold two vertical pixels into one row, so a sprite is SpriteW columns
// by SpriteH/2 terminal rows.
const (
	SpriteW = 16
	SpriteH = 16
	CardGap = 2
)

// Frame is one SpriteH by SpriteW pixel grid, row-major.
type Frame [][]Pix

// Sprite is the open pose and the blink pose.
type Sprite struct {
	Open  Frame
	Blink Frame
}

// Render paints a sprite. Bob shifts the character down by that many
// pixels inside the tile. Selected draws the dark rounded tile behind
// any transparent pixel.
func Render(s Sprite, blink bool, mode ColorMode, bob int, selected bool) string {
	frame := s.Open
	if blink && len(s.Blink) == SpriteH {
		frame = s.Blink
	}
	canvas := make(Frame, SpriteH)
	for y := range canvas {
		canvas[y] = make([]Pix, SpriteW)
	}
	if selected {
		fillTile(canvas)
	}
	if bob < 0 {
		bob = 0
	}
	for y := 0; y < SpriteH; y++ {
		src := y - bob
		if src < 0 || src >= len(frame) {
			continue
		}
		for x := 0; x < SpriteW && x < len(frame[src]); x++ {
			if frame[src][x] != PixClear {
				canvas[y][x] = frame[src][x]
			}
		}
	}
	return paintFrame(canvas, mode)
}

func fillTile(dst Frame) {
	for y := 0; y < SpriteH; y++ {
		for x := 0; x < SpriteW; x++ {
			if tilePixel(x, y) {
				dst[y][x] = PixTile
			}
		}
	}
}

// tilePixel is a chunky rounded rectangle, corners knocked out.
func tilePixel(x, y int) bool {
	if (x == 0 || x == SpriteW-1) && (y <= 1 || y >= SpriteH-2) {
		return false
	}
	if (y == 0 || y == SpriteH-1) && (x <= 1 || x >= SpriteW-2) {
		return false
	}
	return true
}

func paintFrame(frame Frame, mode ColorMode) string {
	var b strings.Builder
	rows := len(frame) / 2
	for row := range rows {
		if row > 0 {
			b.WriteByte('\n')
		}
		y := row * 2
		for x := 0; x < SpriteW; x++ {
			b.WriteString(paintCell(frame[y][x], frame[y+1][x], mode))
		}
	}
	return b.String()
}

func paintCell(top, bot Pix, mode ColorMode) string {
	if mode == ColorNone {
		return monoCell(top, bot)
	}
	if top == PixClear && bot == PixClear {
		return " "
	}
	tc, tok := top.rgba()
	bc, bok := bot.rgba()
	switch {
	case tok && bok && tc == bc:
		return fgOf(tc, mode) + "█" + reset
	case tok && bok:
		return fgOf(tc, mode) + bgOf(bc, mode) + "▀" + reset
	case tok:
		return fgOf(tc, mode) + "▀" + reset
	case bok:
		return fgOf(bc, mode) + "▄" + reset
	default:
		return " "
	}
}

func fgOf(c color.RGBA, mode ColorMode) string {
	if mode == ColorANSI256 {
		return fg256(nearest256(c))
	}
	return fgTrue(c)
}

func bgOf(c color.RGBA, mode ColorMode) string {
	if mode == ColorANSI256 {
		return bg256(nearest256(c))
	}
	return bgTrue(c)
}

// monoCell keeps the silhouette, and the eyes, readable without color.
// Eyes are heavy blocks. Bodies are medium shades. The badge star is a
// literal asterisk so it survives a monochrome dump.
func monoCell(top, bot Pix) string {
	if top == PixStar || bot == PixStar {
		return "*"
	}
	wt := monoWeight(top)
	wb := monoWeight(bot)
	switch {
	case wt == 0 && wb == 0:
		return " "
	case wt >= 3 && wb >= 3:
		return "█"
	case wt >= 3 && wb == 0:
		return "▀"
	case wt == 0 && wb >= 3:
		return "▄"
	case wt >= 3:
		return "▀"
	case wb >= 3:
		return "▄"
	case top == PixOrange || bot == PixOrange:
		return "▓"
	case wt >= 2 && wb >= 2:
		return "▒"
	case wt >= 2 && wb == 0:
		return "▀"
	case wt == 0 && wb >= 2:
		return "▄"
	case wt >= 2 || wb >= 2:
		return "▒"
	case wt == 1 || wb == 1:
		return "░"
	default:
		return " "
	}
}

func monoWeight(p Pix) int {
	switch p {
	case PixClear:
		return 0
	case PixTile:
		return 1
	case PixEye:
		return 3
	default:
		return 2
	}
}
