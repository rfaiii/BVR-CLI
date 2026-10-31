package bots

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// Fake is the demo Backend. It streams canned, personality-flavored lines
// and never leaves the process.
type Fake struct {
	instant bool
}

// NewFake returns the demo backend with each bot's own typing delays.
func NewFake() *Fake {
	return &Fake{}
}

// Instant returns a backend whose delays are zero. Tests use it.
func (f *Fake) Instant() *Fake {
	f.instant = true
	return f
}

// List implements Backend.
func (f *Fake) List() []Bot {
	return Catalog()
}

// Send implements Backend.
func (f *Fake) Send(ctx context.Context, botID, msg string) (<-chan Reply, error) {
	bot, ok := ByID(botID)
	if !ok {
		return nil, fmt.Errorf("unknown bot %q", botID)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	text := renderLine(bot, msg)
	var emote Emote
	if n := len(bot.Voice.Emotes); n > 0 {
		emote = bot.Voice.Emotes[indexOf(n, msg)]
	}
	chunks := chunksOf(text)
	ch := make(chan Reply)
	lead := time.Duration(bot.Voice.Lead) * time.Millisecond
	gap := time.Duration(bot.Voice.Gap) * time.Millisecond
	if f.instant {
		lead, gap = 0, 0
	}
	boop := bot.Cues.Boop != ""
	go func() {
		defer close(ch)
		if !sleep(ctx, lead) {
			return
		}
		for i, chunk := range chunks {
			if i > 0 && !sleep(ctx, gap) {
				return
			}
			reply := Reply{Text: chunk}
			if i == 0 {
				reply.Emote = emote.Glyph
				reply.Boop = boop
			}
			select {
			case <-ctx.Done():
				return
			case ch <- reply:
			}
		}
	}()
	return ch, nil
}

func renderLine(bot Bot, msg string) string {
	lines := bot.Voice.Lines
	if len(lines) == 0 {
		return "Demo."
	}
	line := lines[indexOf(len(lines), msg)]
	quoted := oneLine(msg)
	if bot.ID == "gnarf" {
		quoted = strings.ToUpper(quoted)
	}
	return strings.ReplaceAll(line, "{msg}", quoted)
}

func oneLine(msg string) string {
	msg = strings.Join(strings.Fields(msg), " ")
	if msg == "" {
		return "..."
	}
	if utf8.RuneCountInString(msg) <= 80 {
		return msg
	}
	rs := []rune(msg)
	return string(rs[:80])
}

func chunksOf(text string) []string {
	parts := strings.Split(text, " ")
	out := make([]string, 0, len(parts))
	for i, p := range parts {
		if p == "" {
			continue
		}
		if i == 0 {
			out = append(out, p)
			continue
		}
		out = append(out, " "+p)
	}
	if len(out) == 0 {
		return []string{text}
	}
	return out
}

func indexOf(n int, msg string) int {
	if n <= 1 {
		return 0
	}
	h := 2166136261
	for _, r := range msg {
		h ^= int(r)
		h *= 16777619
	}
	if h < 0 {
		h = -h
	}
	return h % n
}

func sleep(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		select {
		case <-ctx.Done():
			return false
		default:
			return true
		}
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
