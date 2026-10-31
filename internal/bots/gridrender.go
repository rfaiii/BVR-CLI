package bots

import (
	"image/color"
	"strings"
)

// RenderGrid paints a letter grid. B is body, E is eye, anything else is
// empty. An odd height gains one transparent row at the bottom so
// half-blocks stay paired without shifting the baseline between frames.
func RenderGrid(rows []string, body, eye color.RGBA, mode ColorMode, selected bool) string {
	if len(rows) == 0 {
		return ""
	}
	h, w := len(rows), len(rows[0])
	if h%2 == 1 {
		h++
	}
	canvas := make([][]rgbaPx, h)
	for y := range canvas {
		canvas[y] = make([]rgbaPx, w)
	}
	tile, _ := PixTile.rgba()
	if selected {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if tileAt(x, y, w, h) {
					canvas[y][x] = rgbaPx{c: tile, on: true}
				}
			}
		}
	}
	for y, row := range rows {
		for x, ch := range row {
			switch ch {
			case 'B':
				canvas[y][x] = rgbaPx{c: body, on: true}
			case 'E':
				canvas[y][x] = rgbaPx{c: eye, on: true, eye: true}
			}
		}
	}
	var b strings.Builder
	for row := 0; row < h; row += 2 {
		if row > 0 {
			b.WriteByte('\n')
		}
		for x := 0; x < w; x++ {
			b.WriteString(paintRGBA(canvas[row][x], canvas[row+1][x], mode))
		}
	}
	return b.String()
}

func tileAt(x, y, w, h int) bool {
	if w < 3 || h < 3 {
		return true
	}
	if (x == 0 || x == w-1) && (y <= 1 || y >= h-2) {
		return false
	}
	if (y == 0 || y == h-1) && (x <= 1 || x >= w-2) {
		return false
	}
	return true
}

type rgbaPx struct {
	c   color.RGBA
	on  bool
	eye bool
}

func paintRGBA(top, bot rgbaPx, mode ColorMode) string {
	if mode == ColorNone {
		return monoRGBA(top, bot)
	}
	if !top.on && !bot.on {
		return " "
	}
	switch {
	case top.on && bot.on && top.c == bot.c:
		return fgOf(top.c, mode) + "█" + reset
	case top.on && bot.on:
		return fgOf(top.c, mode) + bgOf(bot.c, mode) + "▀" + reset
	case top.on:
		return fgOf(top.c, mode) + "▀" + reset
	default:
		return fgOf(bot.c, mode) + "▄" + reset
	}
}

func monoRGBA(top, bot rgbaPx) string {
	wt, wb := 0, 0
	if top.on {
		wt = 2
		if top.eye {
			wt = 3
		}
	}
	if bot.on {
		wb = 2
		if bot.eye {
			wb = 3
		}
	}
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
	case wt >= 2 && wb >= 2:
		return "▒"
	case wt >= 2:
		return "▀"
	case wb >= 2:
		return "▄"
	default:
		return "░"
	}
}
