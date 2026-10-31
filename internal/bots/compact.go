package bots

import "image/color"

// CompactW and CompactH are the small mascot canvas. Eleven cells is about
// two-thirds of the 16–17 cell mascots. Height is even so half-blocks pair.
const (
	CompactW = 11
	CompactH = 10
)

// CompactCardW is the picker slot. Fullwidth DINGUS and BROBRO are 12.
const CompactCardW = 12

// LargeCardW fits the 17-wide Pillo clip. The other large sprites are 16.
const LargeCardW = 17

// Small eye, identical on every compact bot. Right eye sits one column
// gap to the right and one row higher. The corners stay body or empty
// so the oval still reads after the sprite is shrunk.
//
//	.E
//	EE
//	E.
var compactEye = [][2]int{{1, 0}, {0, 1}, {1, 1}, {0, 2}}

// Compact is the hand-tuned small grid for a roster id.
func Compact(id string) (rows []string, body, eye color.RGBA, ok bool) {
	switch id {
	case "dingus":
		return dingusSmall, dingusBody, dingusEye, true
	case "brobro":
		return brobroSmall, brobroBody, brobroEye, true
	case "pillo":
		return pilloSmallRest(), PilloBody, PilloEye, true
	case "luvah":
		return luvahSmall, luvahBody, luvahEye, true
	case "gnarf":
		return gnarfSmall, gnarfBody, gnarfEye, true
	default:
		return nil, color.RGBA{}, color.RGBA{}, false
	}
}

// PilloClipSmall is the six-frame pill on the compact canvas. Sequences
// match the large clip, so the dialog can swap size without new timings.
func PilloClipSmall() *Clip {
	return pilloSmallClip
}

var (
	dingusBody = color.RGBA{R: 0xFE, G: 0xFD, B: 0xFE, A: 0xFF}
	dingusEye  = color.RGBA{R: 0x07, G: 0x06, B: 0x06, A: 0xFF}
	brobroBody = color.RGBA{R: 0x2B, G: 0xE2, B: 0x6E, A: 0xFF}
	brobroEye  = color.RGBA{R: 0x04, G: 0x04, B: 0x04, A: 0xFF}
	luvahBody  = color.RGBA{R: 0xFD, G: 0x3B, B: 0xAC, A: 0xFF}
	luvahEye   = color.RGBA{R: 0x04, G: 0x04, B: 0x04, A: 0xFF}
	gnarfBody  = color.RGBA{R: 0x91, G: 0x59, B: 0xFD, A: 0xFF}
	gnarfEye   = color.RGBA{R: 0x05, G: 0x05, B: 0x05, A: 0xFF}
)

var dingusSmall = placeEyes(compactRows(
	"...BBBBB...",
	"..BBBBBBB..",
	".BBBBBBBBB.",
	"BBBBBBBBBBB",
	"BBBBBBBBBBB",
	"BBBBBBBBBBB",
	".BBBBBBBBB.",
	"..BBBBBBB..",
	"...BBBBB...",
	"....BBB....",
), 5, 2, 9, 1)

var brobroSmall = placeEyes(compactRows(
	".....B.....",
	"....BBB....",
	"...BBBBB...",
	"..BBBBBBB..",
	".BBBBBBBBB.",
	"BBBBBBBBBBB",
	"BBBBBBBBBBB",
	".BBBBBBBBB.",
	"..BBBBBBB..",
	"...BBBBB...",
), 4, 4, 7, 3)

var luvahSmall = placeEyes(compactRows(
	"...........",
	"...........",
	"..BB...BB..",
	".BBBB.BBBB.",
	"BBBBBBBBBBB",
	".BBBBBBBBB.",
	"..BBBBBBB..",
	"...BBBBB...",
	"....BBB....",
	".....B.....",
), 6, 4, 9, 3)

var gnarfSmall = placeEyes(compactRows(
	"....BBB....",
	"...BBBBB...",
	"..BBBBBBB..",
	".BBBBBBBBB.",
	"BBBBBBBBBBB",
	"BBBBBBBBBBB",
	".BBBBBBBBB.",
	"..BBBBBBB..",
	"...BBBBB...",
	"....BBB....",
), 4, 5, 7, 4)

// Rest pill. The other small frames share this baseline.
func pilloRestBody() []string {
	return compactRows(
		"...........",
		"...........",
		".....BBBBB.",
		"...BBBBBBBB",
		"..BBBBBBBBB",
		".BBBBBBBBBB",
		"BBBBBBBBBBB",
		"BBBBBBBBBB.",
		".BBBBBBBB..",
		"..BBBBBB...",
	)
}

func pilloSmallRest() []string {
	return placeEyes(pilloRestBody(), 6, 4, 9, 3)
}

var pilloSmallClip = &Clip{
	W: CompactW, H: CompactH,
	Seqs: map[string]Sequence{
		"rest": {Name: "rest", Steps: []Step{{Frame: 0, Ticks: 1}}},
		"idle": {Name: "idle", Steps: []Step{
			{Frame: 0, Ticks: 8},
			{Frame: 2, Ticks: 2},
			{Frame: 0, Ticks: 6},
			{Frame: 3, Ticks: 2},
		}},
		"skeptical": {Name: "skeptical", Steps: []Step{
			{Frame: 1, Ticks: 3},
			{Frame: 0, Ticks: 2},
		}},
		"talk": {Name: "talk", Steps: []Step{
			{Frame: 2, Ticks: 2},
			{Frame: 0, Ticks: 2},
			{Frame: 3, Ticks: 2},
			{Frame: 0, Ticks: 2},
		}},
		"select": {Name: "select", Once: true, Steps: []Step{
			{Frame: 4, Ticks: 2},
			{Frame: 5, Ticks: 2},
			{Frame: 0, Ticks: 2},
		}},
	},
}

func init() {
	rest := pilloRestBody()
	squish := compactRows(
		"...........",
		"...........",
		"...........",
		"...........",
		"....BBBBBB.",
		"..BBBBBBBBB",
		".BBBBBBBBBB",
		"BBBBBBBBBBB",
		".BBBBBBBBB.",
		"..BBBBBB...",
	)
	hop := compactRows(
		"....BBB....",
		"...BBBBB...",
		"..BBBBBBB..",
		".BBBBBBBBB.",
		"BBBBBBBBBBB",
		"BBBBBBBBBBB",
		".BBBBBBBBB.",
		"..BBBBBBB..",
		"...BBBBB...",
		"....BBB....",
	)
	pilloSmallClip.Frames = [][]string{
		placeEyes(rest, 6, 4, 9, 3),   // rest, eyes on the right
		placeEyes(rest, 6, 4, 9, 2),   // skeptical: right eye up
		placeEyes(rest, 6, 5, 9, 2),   // glance: right high, left lower
		placeEyes(rest, 5, 4, 8, 3),   // both eyes shifted left
		placeEyes(squish, 5, 6, 8, 5), // squish, eyes lower, same baseline
		placeEyes(hop, 5, 5, 8, 4),    // hop: taller, narrower
	}
}

func compactRows(rows ...string) []string {
	out := make([]string, len(rows))
	copy(out, rows)
	return out
}

// placeEyes copies rows and paints two matching eyes. lx,ly and rx,ry
// are the top-left of each 2×3 oval.
func placeEyes(rows []string, lx, ly, rx, ry int) []string {
	g := make([][]byte, len(rows))
	for i, row := range rows {
		g[i] = []byte(row)
	}
	place := func(ox, oy int) {
		for _, c := range compactEye {
			g[oy+c[1]][ox+c[0]] = 'E'
		}
	}
	place(lx, ly)
	place(rx, ry)
	out := make([]string, len(g))
	for i := range g {
		out[i] = string(g[i])
	}
	return out
}

// Squint drops a protruding top pixel and leaves the oval connected.
// A pixel goes only when it has an eye below and no eye beside or above.
func Squint(rows []string) []string {
	out := make([]string, len(rows))
	for y, row := range rows {
		b := []byte(row)
		for x := range b {
			if b[x] != 'E' {
				continue
			}
			above := y > 0 && rows[y-1][x] == 'E'
			below := y+1 < len(rows) && rows[y+1][x] == 'E'
			left := x > 0 && row[x-1] == 'E'
			right := x+1 < len(row) && row[x+1] == 'E'
			if !above && !left && !right && below {
				b[x] = 'B'
			}
		}
		out[y] = string(b)
	}
	return out
}
