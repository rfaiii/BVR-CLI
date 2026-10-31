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

// Hand-drawn. Both eyes are the same 2×3 oval, the right one one row
// higher, with a column of body between them and a body pixel on every
// side. No eye sits on '.' .
var dingusSmall = []string{
	"...BBBBB...",
	"..BBBBBBB..",
	".BBBBBBBEB.",
	"BBBBBEBEEBB",
	"BBBBEEBEBBB",
	"BBBBEBBBBBB",
	".BBBBBBBBB.",
	"..BBBBBBB..",
	"...BBBBB...",
	"....BBB....",
}

var brobroSmall = []string{
	".....B.....",
	"....BBB....",
	"...BBBBB...",
	"..BBBBBBB..",
	".BBBBBBEBB.",
	"BBBBEBEEBBB",
	"BBBEEBEBBBB",
	".BBEBBBBBB.",
	"..BBBBBBB..",
	"...BBBBB...",
}

var luvahSmall = []string{
	"...........",
	"..BBB.BBB..",
	".BBBBBBBEB.",
	"BBBBBEBEEBB",
	"BBBBEEBEBBB",
	".BBBEBBBBB.",
	"..BBBBBBB..",
	"...BBBBB...",
	"....BBB....",
	".....B.....",
}

var gnarfSmall = []string{
	"....BBB....",
	"...BBBBB...",
	"..BBBBBBB..",
	".BBBBBBBBB.",
	"BBBBBBBBBBB",
	"BBBBBBBEBBB",
	"BBBBEBEEBBB",
	".BBEEBEBBB.",
	"..BEBBBBB..",
	"...BBBBB...",
}

func pilloSmallRest() []string {
	return pilloSmall01
}

var pilloSmallClip = &Clip{
	W: CompactW, H: CompactH,
	Frames: [][]string{
		pilloSmall01, pilloSmall02, pilloSmall03,
		pilloSmall04, pilloSmall05, pilloSmall06,
	},
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

var (
	// Rest. Eyes on the right, feet on the last row.
	pilloSmall01 = []string{
		".BBBBBBBBB.",
		"BBBBBBBBBBB",
		"BBBBBBBBEBB",
		"BBBBBEBEEBB",
		"BBBBEEBEBBB",
		"BBBBEBBBBBB",
		"BBBBBBBBBBB",
		".BBBBBBBBB.",
		"..BBBBBBB..",
		"...BBBBB...",
	}
	// Skeptical: right eye raised one row.
	pilloSmall02 = []string{
		".BBBBBBBBB.",
		"BBBBBBBBEBB",
		"BBBBBBBEEBB",
		"BBBBBEBEBBB",
		"BBBBEEBBBBB",
		"BBBBEBBBBBB",
		"BBBBBBBBBBB",
		".BBBBBBBBB.",
		"..BBBBBBB..",
		"...BBBBB...",
	}
	// Glance: right eye high, left eye dropped.
	pilloSmall03 = []string{
		".BBBBBBBBB.",
		"BBBBBBBBEBB",
		"BBBBBBBEEBB",
		"BBBBBBBEBBB",
		"BBBBBEBBBBB",
		"BBBBEEBBBBB",
		"BBBBEBBBBBB",
		".BBBBBBBBB.",
		"..BBBBBBB..",
		"...BBBBB...",
	}
	// Both eyes shifted left.
	pilloSmall04 = []string{
		".BBBBBBBBB.",
		"BBBBBBBBBBB",
		"BBBBBBBEBBB",
		"BBBBEBEEBBB",
		"BBBEEBEBBBB",
		"BBBEBBBBBBB",
		"BBBBBBBBBBB",
		".BBBBBBBBB.",
		"..BBBBBBB..",
		"...BBBBB...",
	}
	// Squish. Shorter pill, eyes lower, same feet as rest.
	pilloSmall05 = []string{
		"...........",
		"...........",
		"...........",
		".BBBBBBBBB.",
		"BBBBBBBBEBB",
		"BBBBBEBEEBB",
		"BBBBEEBEBBB",
		".BBBEBBBBB.",
		"..BBBBBBB..",
		"...BBBBB...",
	}
	// Hop. Taller and narrower, same oval, still inside the body.
	pilloSmall06 = []string{
		"....BBB....",
		"...BBBBB...",
		"..BBBBBBB..",
		".BBBBBBEBB.",
		"BBBBEBEEBBB",
		"BBBEEBEBBBB",
		".BBEBBBBBB.",
		"..BBBBBBB..",
		"...BBBBB...",
		"....BBB....",
	}
)

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
