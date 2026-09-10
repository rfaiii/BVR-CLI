package ggwave

import "math"

// Mode controls the visual language of the homescreen waveform.
type Mode uint8

const (
	ModeIdle Mode = iota
	ModeListening
	ModeTransmitting
	ModePaired
)

// Waveform renders a deterministic, low-allocation terminal waveform. It is
// presentation-only: no audio is captured or emitted by this function.
func Waveform(width int, frame uint64, mode Mode) string {
	if width < 1 {
		return ""
	}
	levels := []rune("▁▂▃▄▅▆▇█")
	result := make([]rune, 0, width)
	for x := 0; x < width; x++ {
		phase := float64(x)*0.52 + float64(frame)*0.31
		amplitude := 0.20 + 0.13*math.Sin(float64(frame)*0.11)
		switch mode {
		case ModeListening:
			amplitude += 0.18
		case ModeTransmitting:
			amplitude += 0.28 + 0.12*math.Sin(float64(frame)*0.23)
		case ModePaired:
			amplitude += 0.08
		}
		value := 0.5 + amplitude*math.Sin(phase) + 0.08*math.Sin(phase*2.7)
		level := int(math.Round(value * float64(len(levels)-1)))
		if level < 0 {
			level = 0
		} else if level >= len(levels) {
			level = len(levels) - 1
		}
		result = append(result, levels[level])
	}
	return string(result)
}
