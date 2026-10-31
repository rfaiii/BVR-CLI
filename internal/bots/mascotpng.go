package bots

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"
	"strings"
)

// MascotGrid is one extracted mascot, one art pixel per cell.
type MascotGrid struct {
	Name string
	Body color.RGBA
	Eye  color.RGBA
	Rows []string
}

// LoadMascotGrid reads a sprite text grid.
func LoadMascotGrid(path string) (MascotGrid, error) {
	f, err := os.Open(path)
	if err != nil {
		return MascotGrid{}, err
	}
	defer f.Close()
	var g MascotGrid
	var rows []string
	past := false
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "---" {
			past = true
			continue
		}
		if !past {
			switch {
			case strings.HasPrefix(line, "name:"):
				g.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			case strings.HasPrefix(line, "body:"):
				g.Body, err = parseHex(strings.TrimSpace(strings.TrimPrefix(line, "body:")))
			case strings.HasPrefix(line, "eye:"):
				g.Eye, err = parseHex(strings.TrimSpace(strings.TrimPrefix(line, "eye:")))
			}
			if err != nil {
				return MascotGrid{}, err
			}
			continue
		}
		if line != "" {
			rows = append(rows, line)
		}
	}
	if err := sc.Err(); err != nil {
		return MascotGrid{}, err
	}
	g.Rows = rows
	return g, nil
}

func parseHex(s string) (color.RGBA, error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return color.RGBA{}, fmt.Errorf("color %q", s)
	}
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return color.RGBA{}, err
	}
	return color.RGBA{R: uint8(n >> 16), G: uint8(n >> 8), B: uint8(n), A: 255}, nil
}

// WriteMascotPNG writes the grid at one file pixel per art pixel.
func WriteMascotPNG(path string, g MascotGrid) error {
	h, w := len(g.Rows), len(g.Rows[0])
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y, row := range g.Rows {
		for x, ch := range row {
			switch ch {
			case 'B':
				img.SetRGBA(x, y, g.Body)
			case 'E':
				img.SetRGBA(x, y, g.Eye)
			}
		}
	}
	return writePNG(path, img)
}

// RenderMascotPreview draws the five mascots, nearest-neighbor scaled,
// with a full label under each name.
func RenderMascotPreview(grids []MascotGrid) *image.RGBA {
	const (
		scale = 16
		gap   = 36
		pad   = 24
	)
	labelH := 16
	rowH := 0
	for _, g := range grids {
		if h := len(g.Rows) * scale; h > rowH {
			rowH = h
		}
	}
	slotW := func(g MascotGrid) int {
		sw := len(g.Rows[0]) * scale
		lw := labelWidth(g.Name)
		if lw > sw {
			return lw
		}
		return sw
	}
	rowWidth := func(ids []int) int {
		w := 0
		for _, i := range ids {
			w += slotW(grids[i]) + gap
		}
		return w - gap
	}
	width := pad*2 + rowWidth([]int{0, 1})
	if bottom := pad*2 + rowWidth([]int{2, 3, 4}); bottom > width {
		width = bottom
	}
	height := pad + rowH + labelH + gap + rowH + labelH + pad
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	bg := color.RGBA{8, 8, 10, 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, bg)
		}
	}
	place := func(ids []int, y int) {
		w := rowWidth(ids)
		x := (width - w) / 2
		for _, i := range ids {
			g := grids[i]
			sw := len(g.Rows[0]) * scale
			slot := slotW(g)
			ox := x + (slot-sw)/2
			for yy, row := range g.Rows {
				for xx, ch := range row {
					var c color.RGBA
					switch ch {
					case 'B':
						c = g.Body
					case 'E':
						c = g.Eye
					default:
						continue
					}
					for dy := 0; dy < scale; dy++ {
						for dx := 0; dx < scale; dx++ {
							img.SetRGBA(ox+xx*scale+dx, oy(y, yy, scale, dy), c)
						}
					}
				}
			}
			lw := labelWidth(g.Name)
			drawLabel(img, x+(slot-lw)/2, y+len(g.Rows)*scale+4, g.Name)
			x += slot + gap
		}
	}
	place([]int{0, 1}, pad)
	place([]int{2, 3, 4}, pad+rowH+labelH+gap/2)
	return img
}

func oy(base, row, scale, dy int) int {
	return base + row*scale + dy
}

func labelWidth(name string) int {
	w := 0
	n := 0
	for _, r := range strings.ToUpper(name) {
		glyph, ok := mascotFont[r]
		if !ok || len(glyph) == 0 {
			continue
		}
		if n > 0 {
			w += 2
		}
		w += len(glyph[0]) * 2
		n++
	}
	return w
}

var mascotFont = map[rune][]string{
	'A': {"###", "#.#", "###", "#.#", "#.#"},
	'B': {"##.", "#.#", "##.", "#.#", "##."},
	'C': {"###", "#..", "#..", "#..", "###"},
	'D': {"##.", "#.#", "#.#", "#.#", "##."},
	'E': {"###", "#..", "##.", "#..", "###"},
	'F': {"###", "#..", "##.", "#..", "#.."},
	'G': {"###", "#..", "#.#", "#.#", "###"},
	'H': {"#.#", "#.#", "###", "#.#", "#.#"},
	'I': {"###", ".#.", ".#.", ".#.", "###"},
	'K': {"#.#", "##.", "#..", "##.", "#.#"},
	'L': {"#..", "#..", "#..", "#..", "###"},
	'M': {"#.#", "###", "#.#", "#.#", "#.#"},
	// Two vertical strokes and one diagonal from top-left to bottom-right.
	'N': {"#..#", "##.#", "#.##", "#..#", "#..#"},
	'O': {".#.", "#.#", "#.#", "#.#", ".#."},
	'P': {"##.", "#.#", "##.", "#..", "#.."},
	'R': {"##.", "#.#", "##.", "#.#", "#.#"},
	'S': {"###", "#..", "###", "..#", "###"},
	'T': {"###", ".#.", ".#.", ".#.", ".#."},
	'U': {"#.#", "#.#", "#.#", "#.#", "###"},
	'V': {"#.#", "#.#", "#.#", ".#.", ".#."},
	'W': {"#.#", "#.#", "#.#", "###", "#.#"},
	'Y': {"#.#", "#.#", ".#.", ".#.", ".#."},
}

func drawLabel(img *image.RGBA, x, y int, name string) {
	fg := color.RGBA{236, 236, 240, 255}
	for _, r := range strings.ToUpper(name) {
		if r == ' ' {
			x += 8
			continue
		}
		glyph, ok := mascotFont[r]
		if !ok {
			continue
		}
		for gy, row := range glyph {
			for gx, ch := range row {
				if ch != '#' {
					continue
				}
				for dy := 0; dy < 2; dy++ {
					for dx := 0; dx < 2; dx++ {
						img.SetRGBA(x+gx*2+dx, y+gy*2+dy, fg)
					}
				}
			}
		}
		x += len(glyph[0])*2 + 2
	}
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
