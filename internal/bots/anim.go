package bots

import "os"

// Step is one frame in an animation, held for Ticks of the 100ms clock.
type Step struct {
	Frame int
	Ticks int
}

// Sequence is a named animation. Once sequences stop on the last step.
type Sequence struct {
	Name  string
	Steps []Step
	Once  bool
}

// Clip is a fixed-size frame stack plus named sequences. Every frame is
// W by H so playback does not shift the baseline.
type Clip struct {
	W, H   int
	Body   string
	Eye    string
	Frames [][]string
	Seqs   map[string]Sequence
}

// Player walks a clip on the dialog's 100ms tick.
type Player struct {
	clip *Clip
	name string
	step int
	tick int
	done bool
}

// NewPlayer starts on the rest sequence, which must be a one-frame loop
// of the base pose.
func NewPlayer(c *Clip) *Player {
	p := &Player{clip: c, name: "rest"}
	if _, ok := c.Seqs["idle"]; ok {
		p.name = "idle"
	}
	return p
}

// Name is the sequence currently playing.
func (p *Player) Name() string {
	if p == nil {
		return ""
	}
	return p.name
}

// Done reports a once-sequence that has finished its last hold.
func (p *Player) Done() bool {
	return p != nil && p.done
}

// Play switches sequence. Repeating the active sequence does nothing, so
// a loop is not restarted every tick.
func (p *Player) Play(name string) {
	if p == nil || p.name == name && !p.done {
		return
	}
	if _, ok := p.clip.Seqs[name]; !ok {
		return
	}
	p.name = name
	p.step = 0
	p.tick = 0
	p.done = false
}

// Advance consumes one 100ms tick. Reduce-motion holds frame 0.
func (p *Player) Advance() {
	if p == nil {
		return
	}
	if ReduceMotion() {
		p.name = "rest"
		p.step = 0
		p.tick = 0
		p.done = false
		return
	}
	seq, ok := p.clip.Seqs[p.name]
	if !ok || len(seq.Steps) == 0 || p.done {
		return
	}
	hold := seq.Steps[p.step].Ticks
	if hold < 1 {
		hold = 1
	}
	p.tick++
	if p.tick < hold {
		return
	}
	p.tick = 0
	if p.step+1 < len(seq.Steps) {
		p.step++
		return
	}
	if seq.Once {
		p.done = true
		return
	}
	p.step = 0
}

// Rows is the current frame. Reduce-motion is always the first frame.
func (p *Player) Rows() []string {
	if p == nil || len(p.clip.Frames) == 0 {
		return nil
	}
	if ReduceMotion() {
		return p.clip.Frames[0]
	}
	seq := p.clip.Seqs[p.name]
	if len(seq.Steps) == 0 {
		return p.clip.Frames[0]
	}
	idx := seq.Steps[p.step].Frame
	if idx < 0 || idx >= len(p.clip.Frames) {
		return p.clip.Frames[0]
	}
	return p.clip.Frames[idx]
}

// ReduceMotion is set when the user asked for less animation.
func ReduceMotion() bool {
	return os.Getenv("BVR_REDUCE_MOTION") != "" || os.Getenv("PREFERS_REDUCED_MOTION") != ""
}
