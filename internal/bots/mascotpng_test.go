package bots

import (
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func mascotNames() []string {
	return []string{"dingus", "brobro", "pillo", "luvah", "gnarf"}
}

func TestNGlyphIsTwoStemsAndOneDiagonal(t *testing.T) {
	t.Parallel()
	n := mascotFont['N']
	require.Equal(t, []string{
		"#..#",
		"##.#",
		"#.##",
		"#..#",
		"#..#",
	}, n)
	for _, row := range n {
		require.Equal(t, byte('#'), row[0])
		require.Equal(t, byte('#'), row[len(row)-1])
	}
	// The diagonal steps right once per row, then joins the right stem.
	require.Equal(t, byte('#'), n[1][1])
	require.Equal(t, byte('#'), n[2][2])
	require.Equal(t, byte('.'), n[2][1])
}

func loadAll(t *testing.T) []MascotGrid {
	t.Helper()
	var grids []MascotGrid
	for _, name := range mascotNames() {
		g, err := LoadMascotGrid(filepath.Join("sprites", name+".txt"))
		require.NoError(t, err)
		grids = append(grids, g)
	}
	return grids
}

// The oval is the hand-cleaned eye. Top is shifted one column right.
// Corners stay body color. The eye itself is the tilted oval.
var cleanOval = []string{
	"BEE",
	"EEE",
	"EEE",
	"EEB",
}

func TestMascotEyesAreCleanMatchingOvals(t *testing.T) {
	t.Parallel()
	for _, g := range loadAll(t) {
		t.Run(g.Name, func(t *testing.T) {
			t.Parallel()
			boxes := eyeBoxes(g.Rows)
			require.Len(t, boxes, 2, "want two eyes")
			require.Equal(t, boxes[0].w, boxes[1].w)
			require.Equal(t, boxes[0].h, boxes[1].h)
			require.Equal(t, 3, boxes[0].w)
			require.Equal(t, 4, boxes[0].h)
			require.Greater(t, boxes[1].x, boxes[0].x)
			require.Equal(t, boxes[0].y-1, boxes[1].y, "right eye sits one row higher")
			left := crop(g.Rows, boxes[0])
			right := crop(g.Rows, boxes[1])
			require.Equal(t, cleanOval, left)
			require.Equal(t, cleanOval, right)
		})
	}
}

func TestMascotGridGolden(t *testing.T) {
	t.Parallel()
	var b strings.Builder
	for _, g := range loadAll(t) {
		b.WriteString(g.Name)
		b.WriteByte('\n')
		for _, row := range g.Rows {
			b.WriteString(row)
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}
	path := filepath.Join("testdata", "mascot-grids.golden")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		require.NoError(t, os.WriteFile(path, []byte(b.String()), 0o644))
	}
	want, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, strings.ReplaceAll(string(want), "\r\n", "\n"), strings.ReplaceAll(b.String(), "\r\n", "\n"))
}

func TestPreviewLabelsIncludeEveryLetter(t *testing.T) {
	t.Parallel()
	grids := loadAll(t)
	img := RenderMascotPreview(grids)
	for _, g := range grids {
		for _, r := range strings.ToUpper(g.Name) {
			_, ok := mascotFont[r]
			require.Truef(t, ok, "%s missing font glyph %q", g.Name, r)
		}
		require.LessOrEqual(t, labelWidth(g.Name), img.Bounds().Dx())
		label := image.NewRGBA(image.Rect(0, 0, labelWidth(g.Name), 10))
		drawLabel(label, 0, 0, g.Name)
		require.Truef(t, imageContains(img, label), "%s label is clipped", g.Name)
	}
}

func imageContains(hay, needle *image.RGBA) bool {
	hb, nb := hay.Bounds(), needle.Bounds()
	nw, nh := nb.Dx(), nb.Dy()
	for y := hb.Min.Y; y+nh <= hb.Max.Y; y++ {
		for x := hb.Min.X; x+nw <= hb.Max.X; x++ {
			if labelAt(hay, needle, x, y) {
				return true
			}
		}
	}
	return false
}

func labelAt(hay, needle *image.RGBA, ox, oy int) bool {
	nb := needle.Bounds()
	any := false
	for y := nb.Min.Y; y < nb.Max.Y; y++ {
		for x := nb.Min.X; x < nb.Max.X; x++ {
			nr, ng, nbl, na := needle.At(x, y).RGBA()
			if na == 0 || (nr == 0 && ng == 0 && nbl == 0) {
				continue
			}
			any = true
			hr, hg, hb, _ := hay.At(ox+x, oy+y).RGBA()
			if hr != nr || hg != ng || hb != nbl {
				return false
			}
		}
	}
	return any
}

func TestWriteMascotMedia(t *testing.T) {
	if os.Getenv("UPDATE_GOLDEN") != "1" {
		t.Skip()
	}
	grids := loadAll(t)
	dir := filepath.Join("..", "..", "docs", "media", "bots")
	for _, g := range grids {
		require.NoError(t, WriteMascotPNG(filepath.Join(dir, g.Name+".png"), g))
	}
	require.NoError(t, writePNG(filepath.Join(dir, "preview.png"), RenderMascotPreview(grids)))
}

type eyeBox struct{ x, y, w, h int }

func eyeBoxes(rows []string) []eyeBox {
	h, w := len(rows), len(rows[0])
	seen := make([]bool, w*h)
	var boxes []eyeBox
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if rows[y][x] != 'E' || seen[y*w+x] {
				continue
			}
			minx, miny, maxx, maxy := x, y, x, y
			q := [][2]int{{x, y}}
			seen[y*w+x] = true
			for i := 0; i < len(q); i++ {
				p := q[i]
				if p[0] < minx {
					minx = p[0]
				}
				if p[1] < miny {
					miny = p[1]
				}
				if p[0] > maxx {
					maxx = p[0]
				}
				if p[1] > maxy {
					maxy = p[1]
				}
				for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					xx, yy := p[0]+d[0], p[1]+d[1]
					if xx < 0 || yy < 0 || xx >= w || yy >= h || seen[yy*w+xx] || rows[yy][xx] != 'E' {
						continue
					}
					seen[yy*w+xx] = true
					q = append(q, [2]int{xx, yy})
				}
			}
			boxes = append(boxes, eyeBox{minx, miny, maxx - minx + 1, maxy - miny + 1})
		}
	}
	if len(boxes) == 2 && boxes[0].x > boxes[1].x {
		boxes[0], boxes[1] = boxes[1], boxes[0]
	}
	return boxes
}

func crop(rows []string, b eyeBox) []string {
	out := make([]string, b.h)
	for y := 0; y < b.h; y++ {
		out[y] = rows[b.y+y][b.x : b.x+b.w]
	}
	return out
}
