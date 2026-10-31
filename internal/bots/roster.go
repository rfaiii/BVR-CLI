package bots

import "fmt"

// Emote is a small mark shown beside a chat line. Squint asks the header
// sprite to hold the blink pose so the expression reads on the character.
type Emote struct {
	Glyph  string
	Squint bool
}

// Cues are existing audio event names. The UI maps them through the
// current audio backend onto audio/sounds/*.wav. Empty Boop means the bot
// does not add a notify accent.
type Cues struct {
	Chat   []string
	Errors []string
	Select string
	Boop   string
}

// Voice is the demo personality: canned lines, emotes, and pacing.
// Lines may contain {msg}, replaced with the user's text.
type Voice struct {
	Lead       int // milliseconds before the first streamed chunk
	Gap        int // milliseconds between chunks
	TypeMillis int // typewriter tick
	TypeStep   int // runes revealed per tick
	IdleMillis int // idle animation period
	Bob        bool
	Blink      bool
	Lines      []string
	Emotes     []Emote
	Greeting   string
	GreetEmote Emote
}

// Bot is one roster entry. Add another by appending to the catalog.
type Bot struct {
	ID      string
	Name    string
	Tagline string
	Color   Pix
	Sprite  Sprite
	Voice   Voice
	Cues    Cues
}

// Catalog is Dingus, BroBro, Pillo, Luvah, and Gnarf, in picker order.
func Catalog() []Bot {
	return []Bot{dingus(), brobro(), pillo(), luvah(), gnarf()}
}

// ByID returns a catalog bot.
func ByID(id string) (Bot, bool) {
	for _, b := range Catalog() {
		if b.ID == id {
			return b, true
		}
	}
	return Bot{}, false
}

// ChunkyName renders a name in fullwidth letters so it sits under a
// sprite like a chunky NES label.
func ChunkyName(name string) string {
	var b []rune
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b = append(b, r-'a'+'Ａ')
		case r >= 'A' && r <= 'Z':
			b = append(b, r-'A'+'Ａ')
		case r == ' ':
			b = append(b, '　')
		default:
			b = append(b, r)
		}
	}
	return string(b)
}

// ChatCue picks a chat sound for this message. n varies per send.
func (b Bot) ChatCue(n int) string {
	return pick(b.Cues.Chat, n)
}

// ErrorCue picks an error sound.
func (b Bot) ErrorCue(n int) string {
	return pick(b.Cues.Errors, n)
}

func pick(items []string, n int) string {
	if len(items) == 0 {
		return ""
	}
	if n < 0 {
		n = -n
	}
	return items[n%len(items)]
}

func dingus() Bot {
	body := mustBody("dingus", []string{
		"................",
		"...WWWWWWWWW....",
		"..WWWWWWWWWWWW..",
		".WWWWWWWWWWWWWW.",
		".WWWWWWWWWWWWWW.",
		".WWWWWWWWWWWWWW.",
		".WWWWWWWWWWWWWW.",
		".WWWWWWWWWWWWWW.",
		".WWWWWWWWWWWWWW.",
		"..WWWWWWWWWWWW..",
		"...WWWWWWWWWW...",
		"....WWWWWWWW....",
		".....WWWWWW.....",
		"......WWWW......",
		"................",
		"................",
	}, 'W')
	return Bot{
		ID:      "dingus",
		Name:    "Dingus",
		Tagline: "calm",
		Color:   PixWhite,
		Sprite:  mustSprite(body, 7, 2, 10, 11),
		Voice: Voice{
			Lead: 280, Gap: 70, TypeMillis: 34, TypeStep: 1,
			IdleMillis: 640, Blink: true,
			Lines: []string{
				"Understood. {msg}.",
				"That tracks. {msg}.",
				"Okay. {msg}.",
			},
			Emotes: []Emote{
				{Glyph: "•ᴗ•", Squint: true},
				{Glyph: "–"},
			},
			Greeting:   "Hey. I am Dingus. No rush.",
			GreetEmote: Emote{Glyph: "•ᴗ•", Squint: true},
		},
		Cues: Cues{
			Chat:   []string{"chat-01", "chat-02"},
			Errors: []string{"error-01"},
			Select: "quick-notify-01",
		},
	}
}

func brobro() Bot {
	body := mustBody("brobro", []string{
		"................",
		".......GG.......",
		"......GGGG......",
		".....GGGGGG.....",
		"....GGGGGGGG....",
		"...GGGGGGGGGG...",
		"..GGGGGGGGGGGG..",
		".GGGGGGGGGGGGGG.",
		".GGGGGGGGGGGGGG.",
		"GGGGGGGGGGGGGGGG",
		"GGGGGGGGGGGGGGGG",
		".GGGGGGGGGGGGGG.",
		"..GGGGGGGGGGGG..",
		"...GGGGGGGGGG...",
		"................",
		"................",
	}, 'G')
	return Bot{
		ID:      "brobro",
		Name:    "BroBro",
		Tagline: "hyped",
		Color:   PixGreen,
		Sprite:  mustSprite(body, 4, 6, -1, -1),
		Voice: Voice{
			Lead: 40, Gap: 12, TypeMillis: 12, TypeStep: 3,
			IdleMillis: 180, Blink: true, Bob: true,
			Lines: []string{
				"YES!! {msg}!!",
				"BRO!! {msg}!!",
				"WOOO!! {msg}!!",
			},
			Emotes: []Emote{
				{Glyph: "!!"},
				{Glyph: "↗"},
				{Glyph: "✧"},
			},
			Greeting:   "BROBRO IN!! LET'S GO!!",
			GreetEmote: Emote{Glyph: "!!"},
		},
		Cues: Cues{
			Chat:   []string{"chat-01", "chat-02", "chat-03", "chat-04"},
			Errors: []string{"error-02"},
			Select: "quick-notify-02",
		},
	}
}

func pillo() Bot {
	body := mustBody("pillo", []string{
		"................",
		"................",
		"................",
		"..AAAAAAAAAAAA..",
		".AAAAAAAAAAAAAA.",
		"AAAAAAAAAAAAAAAA",
		"AAAAAAAAAAAAAAAA",
		"AAAAAAAAAAAAAAAA",
		"AAAAAAAAAAAAAAAA",
		"AAAAAAAAAAAAAAAA",
		".AAAAAAAAAAAAAA.",
		"..AAAAAAAAAAAA..",
		"................",
		"................",
		"................",
		"................",
	}, 'A')
	return Bot{
		ID:      "pillo",
		Name:    "Pillo",
		Tagline: "terse",
		Color:   PixGray,
		Sprite:  mustSprite(body, 8, 5, -1, -1),
		Voice: Voice{
			Lead: 60, Gap: 16, TypeMillis: 20, TypeStep: 2,
			IdleMillis: 980, Blink: true,
			Lines: []string{
				"{msg}.",
				"Got it. {msg}.",
				"Done. {msg}.",
			},
			Emotes: []Emote{
				{},
				{Glyph: "·"},
			},
			Greeting:   "Pillo. Ask.",
			GreetEmote: Emote{},
		},
		Cues: Cues{
			Chat:   []string{"chat-03"},
			Errors: []string{"error-01", "error-03"},
			Select: "quick-notify-01",
		},
	}
}

func luvah() Bot {
	body := mustBody("luvah", []string{
		"................",
		"...PP....PP.....",
		"..PPPP..PPPP....",
		"..PPPPPPPPPPP...",
		".PPPPPPPPPPPPP..",
		".PPPPPPPPPPPPPP.",
		"..PPPPPPPPPPPPP.",
		"...PPPPPPPPPPPP.",
		"....PPPPPPPPPP..",
		".....PPPPPPPP...",
		"......PPPPPP....",
		".......PPPP.....",
		"........PP......",
		"................",
		"................",
		"................",
	}, 'P')
	return Bot{
		ID:      "luvah",
		Name:    "Luvah",
		Tagline: "kind",
		Color:   PixPink,
		Sprite:  mustSprite(body, 6, 3, -1, -1),
		Voice: Voice{
			Lead: 320, Gap: 80, TypeMillis: 42, TypeStep: 1,
			IdleMillis: 560, Blink: true,
			Lines: []string{
				"Sweet. {msg}.",
				"I am here. {msg}.",
				"You are doing fine. {msg}.",
			},
			Emotes: []Emote{
				{Glyph: "♥"},
				{Glyph: "✧", Squint: true},
				{Glyph: "♡"},
			},
			Greeting:   "Hello, dear. I am Luvah.",
			GreetEmote: Emote{Glyph: "♥"},
		},
		Cues: Cues{
			Chat:   []string{"chat-04", "chat-01"},
			Errors: []string{"error-01"},
			Select: "quick-notify-03",
		},
	}
}

func gnarf() Bot {
	body := mustBody("gnarf", []string{
		"................",
		".....UUUUUUU....",
		"....UUUUUUUUUU..",
		"...UUUUUUUUUUU..",
		"..UUUUUUUUUUUUU.",
		".UUUUUUUUUUUUUU.",
		".UUUUUUUUUUUUUU.",
		"UUUUUUUUUUUUUUUU",
		".UUUUUUUUUUUUUU.",
		"..UUUUUUUUUUUU..",
		"...UUUUUUUUUU...",
		"....UUUUUUUU....",
		".....UUUUUU.....",
		"................",
		"................",
		"................",
	}, 'U')
	return Bot{
		ID:      "gnarf",
		Name:    "Gnarf",
		Tagline: "loud",
		Color:   PixPurple,
		Sprite:  mustSprite(body, 4, 7, -1, -1),
		Voice: Voice{
			Lead: 90, Gap: 28, TypeMillis: 18, TypeStep: 2,
			IdleMillis: 280, Blink: true, Bob: true,
			Lines: []string{
				"HEARD!! {msg}!!",
				"LOUD!! {msg}!!",
				"BOOP!! {msg}!!",
			},
			Emotes: []Emote{
				{Glyph: "♪"},
				{Glyph: "♫"},
				{Glyph: "🔊"},
			},
			Greeting:   "GNARF!! SPEAKERS UP!!",
			GreetEmote: Emote{Glyph: "♪"},
		},
		Cues: Cues{
			Chat:   []string{"chat-02", "chat-04"},
			Errors: []string{"error-02", "error-03"},
			Select: "quick-notify-01",
			Boop:   "quick-notify-03",
		},
	}
}

// Open eyes: two vertical ovals, top shifted right so they tilt.
var eyesOpen = []string{
	".EE..EE",
	".EE..EE",
	"EE..EE.",
	"EE..EE.",
}

// Blink keeps the same ovals and only drops pixels, so the shape stays
// recognizable while the eyes squint.
var eyesBlink = []string{
	".E....E",
	".EE..EE",
	".E...E.",
	"E....E.",
}

var dingusBadge = []string{
	".OOO.",
	"OSSSO",
	"O.S.O",
	".OOO.",
}

var letterPix = map[rune]Pix{
	'W': PixWhite,
	'E': PixEye,
	'O': PixOrange,
	'S': PixStar,
	'G': PixGreen,
	'A': PixGray,
	'P': PixPink,
	'U': PixPurple,
}

func mustBody(name string, rows []string, fill rune) Frame {
	if len(rows) != SpriteH {
		panic(fmt.Sprintf("%s: want %d rows, got %d", name, SpriteH, len(rows)))
	}
	g := make(Frame, SpriteH)
	for y, row := range rows {
		rs := []rune(row)
		if len(rs) != SpriteW {
			panic(fmt.Sprintf("%s row %d: want %d cols, got %d", name, y, SpriteW, len(rs)))
		}
		g[y] = make([]Pix, SpriteW)
		for x, r := range rs {
			if r == '.' {
				continue
			}
			if r != fill {
				panic(fmt.Sprintf("%s row %d col %d: unexpected %q", name, y, x, r))
			}
			g[y][x] = letterPix[r]
		}
	}
	return g
}

func mustSprite(body Frame, eyeX, eyeY, badgeX, badgeY int) Sprite {
	open := stampEyes(body, eyeX, eyeY, eyesOpen)
	blink := stampEyes(body, eyeX, eyeY, eyesBlink)
	if badgeX >= 0 {
		open = stampBadge(open, badgeX, badgeY)
		blink = stampBadge(blink, badgeX, badgeY)
	}
	if countPix(open, PixEye) < 12 {
		panic("open eyes are too small")
	}
	if !eyesInside(open, body) || !eyesInside(blink, body) {
		panic("eyes landed off the body")
	}
	return Sprite{Open: open, Blink: blink}
}

func stampEyes(body Frame, x, y int, pattern []string) Frame {
	g := cloneFrame(body)
	if err := stamp(g, pattern, x, y, isBody); err != nil {
		panic(err)
	}
	return g
}

func stampBadge(body Frame, x, y int) Frame {
	g := cloneFrame(body)
	if err := stamp(g, dingusBadge, x, y, func(p Pix) bool {
		return p == PixWhite || p == PixClear
	}); err != nil {
		panic(err)
	}
	return g
}

func stamp(dst Frame, rows []string, x, y int, allow func(Pix) bool) error {
	for dy, row := range rows {
		for dx, r := range row {
			if r == '.' {
				continue
			}
			p, ok := letterPix[r]
			if !ok {
				return fmt.Errorf("unknown pixel %q", r)
			}
			yy, xx := y+dy, x+dx
			if yy < 0 || xx < 0 || yy >= len(dst) || xx >= len(dst[yy]) {
				return fmt.Errorf("stamp out of bounds at %d,%d", xx, yy)
			}
			if allow != nil && !allow(dst[yy][xx]) {
				return fmt.Errorf("stamp refused %q at %d,%d", r, xx, yy)
			}
			dst[yy][xx] = p
		}
	}
	return nil
}

func isBody(p Pix) bool {
	switch p {
	case PixWhite, PixGreen, PixGray, PixPink, PixPurple:
		return true
	default:
		return false
	}
}

func eyesInside(frame, body Frame) bool {
	for y := range frame {
		for x := range frame[y] {
			if frame[y][x] == PixEye && !isBody(body[y][x]) {
				return false
			}
		}
	}
	return true
}

func countPix(frame Frame, p Pix) int {
	n := 0
	for y := range frame {
		for x := range frame[y] {
			if frame[y][x] == p {
				n++
			}
		}
	}
	return n
}

func cloneFrame(src Frame) Frame {
	out := make(Frame, len(src))
	for y := range src {
		out[y] = append([]Pix(nil), src[y]...)
	}
	return out
}
