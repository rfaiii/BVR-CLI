package bots

import "image/color"

// Pillo canvas is the max of Richard's six frames, bottom-anchored, so a
// squish or a hop does not slide him across the cell.
const (
	pilloW = 17
	pilloH = 13
)

// PilloBody and PilloEye are the flat colors of the grey pill.
var (
	PilloBody = color.RGBA{R: 0x9A, G: 0x98, B: 0x98, A: 0xFF}
	PilloEye  = color.RGBA{R: 0x05, G: 0x04, B: 0x04, A: 0xFF}
)

// PilloClip is the reusable frame stack for the grey pill. Other bots can
// grow a Clip of the same shape later.
func PilloClip() *Clip {
	return pilloClip
}

var pilloClip = &Clip{
	W: pilloW, H: pilloH,
	Frames: [][]string{pillo01, pillo02, pillo03, pillo04, pillo05, pillo06},
	Seqs: map[string]Sequence{
		// Reduce-motion and the finished select both land here.
		"rest": {Name: "rest", Steps: []Step{{Frame: 0, Ticks: 1}}},
		// Mostly the rest pose, with two short glances.
		"idle": {Name: "idle", Steps: []Step{
			{Frame: 0, Ticks: 8},
			{Frame: 2, Ticks: 2},
			{Frame: 0, Ticks: 6},
			{Frame: 3, Ticks: 2},
		}},
		// Raised right eye. Snappy, then back to rest, while he thinks.
		"skeptical": {Name: "skeptical", Steps: []Step{
			{Frame: 1, Ticks: 3},
			{Frame: 0, Ticks: 2},
		}},
		// Glance up-right, rest, glance left, rest.
		"talk": {Name: "talk", Steps: []Step{
			{Frame: 2, Ticks: 2},
			{Frame: 0, Ticks: 2},
			{Frame: 3, Ticks: 2},
			{Frame: 0, Ticks: 2},
		}},
		// Squish, hop, settle. Plays once.
		"select": {Name: "select", Once: true, Steps: []Step{
			{Frame: 4, Ticks: 2},
			{Frame: 5, Ticks: 2},
			{Frame: 0, Ticks: 2},
		}},
	},
}

// Rest body, eyes up and to the right. One blank row on top of the
// 12-row pill so the baseline matches the taller hop.
var pillo01 = []string{
	".................",
	"..........BBBBB..",
	"....BBBBBBBBBBBB.",
	".BBBBBBBBBBBBEEB.",
	".BBBBBBBBEEBEEEBB",
	"BBBBBBBBEEEBEEEBB",
	"BBBBBBBBEEEBEEBBB",
	"BBBBBBBBEEBBBBBBB",
	"BBBBBBBBBBBBBBBBB",
	"BBBBBBBBBBBBBBBB.",
	".BBBBBBBBBBBBBB..",
	"..BBBBBBBBBBBB...",
	"..BBBBBBBB.......",
}

// Right eye raised: the skeptical beat.
var pillo02 = []string{
	".................",
	"..........BBBBB..",
	"....BBBBBBBBBEEB.",
	".BBBBBBBBBBBEEEB.",
	".BBBBBBBBEEBEEEBB",
	"BBBBBBBBEEEBEEBBB",
	"BBBBBBBBEEEBBBBBB",
	"BBBBBBBBEEBBBBBBB",
	"BBBBBBBBBBBBBBBBB",
	"BBBBBBBBBBBBBBBB.",
	".BBBBBBBBBBBBBB..",
	"..BBBBBBBBBBBB...",
	"..BBBBBBBB.......",
}

// Glance up-right: right eye high, left eye dropped.
var pillo03 = []string{
	".................",
	"..........BBBBB..",
	"....BBBBBBBBBEEB.",
	".BBBBBBBBBBBEEEB.",
	".BBBBBBBBBBBEEEBB",
	"BBBBBBBBBBBBEEBBB",
	"BBBBBBBBBEEBBBBBB",
	"BBBBBBBBEEEBBBBBB",
	"BBBBBBBBEEEBBBBBB",
	"BBBBBBBBEEBBBBBB.",
	".BBBBBBBBBBBBBB..",
	"..BBBBBBBBBBBB...",
	"..BBBBBBBB.......",
}

// Glance left and up. Same ovals, shifted toward the left edge.
var pillo04 = []string{
	".................",
	"..........BBBBB..",
	"....BBBBBEEBBBBB.",
	".BBBBEEBEEEBBBBB.",
	".BBBEEEBEEEBBBBBB",
	"BBBBEEEBEEBBBBBBB",
	"BBBBEEBBBBBBBBBBB",
	"BBBBBBBBBBBBBBBBB",
	"BBBBBBBBBBBBBBBBB",
	"BBBBBBBBBBBBBBBB.",
	".BBBBBBBBBBBBBB..",
	"..BBBBBBBBBBBB...",
	"..BBBBBBBB.......",
}

// Squish. Narrower pill, eyes lower, still on the same baseline.
var pillo05 = []string{
	".................",
	".................",
	"..........BBBB...",
	"......BBBBBBBBB..",
	"....BBBBBBBBBBBB.",
	"...BBBBBBBBBBBBB.",
	"..BBBBBBBBBEEBBB.",
	"..BBBBBEEBEEEBBB.",
	"...BBBEEEBEEEBBB.",
	"....BBEEEBEEBBBB.",
	".....BEEBBBBBB...",
	"......BBBBBB.....",
	".......BBBB......",
}

// Stretch / hop. Taller, narrower, small matching squint-ovals.
var pillo06 = []string{
	"......BBBBB......",
	"....BBBBBBBBB....",
	"...BBBBBBBBBBB...",
	"..BBBBBBBBBBBBB..",
	"..BBBBBBBBBBBBB..",
	"..BBBBBBBBBBBBB..",
	"...BBBBBBBBBBBB..",
	"...BBBBBBBBEBBB..",
	"....BBBEBBEEBB...",
	"....BBEEBBEBBB...",
	".....BEBBBBBB....",
	"......BBBBBB.....",
	".......BBBB......",
}
