package dialog

import (
	"context"
	"image"
	"strings"
	"time"
	"unicode/utf8"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/richavery/bvr-cli/internal/bots"
	"github.com/richavery/bvr-cli/internal/ui/audio"
	"github.com/richavery/bvr-cli/internal/ui/common"
)

// BotsID is the bot picker dialog.
const BotsID = "bots"

// ActionPlayCue asks the UI to play bundled sounds through the existing
// audio backend, then run Cmd. Types are event names such as "chat-01".
type ActionPlayCue struct {
	Types []string
	Cmd   tea.Cmd
}

type botsMode int

const (
	botsPicker botsMode = iota
	botsChat
)

type botsTickMsg struct{}

type writerTickMsg struct{}

type streamOpenMsg struct {
	gen int
	ch  <-chan bots.Reply
	err error
}

type streamChunkMsg struct {
	gen   int
	text  string
	emote string
	boop  bool
	done  bool
}

type chatLine struct {
	fromBot bool
	emote   string
	squint  bool
	full    string
	shown   int
}

// Bots is the NES-style roster and the demo chat that follows a pick.
type Bots struct {
	com       *common.Common
	backend   bots.Backend
	roster    []bots.Bot
	mode      botsMode
	index     int
	clock     int
	muted     bool
	volume    string
	color     bots.ColorMode
	lines     []chatLine
	input     textinput.Model
	writing   bool
	streaming bool
	stream    <-chan bots.Reply
	gen       int
	cancel    context.CancelFunc
	width     int
	pillo     *bots.Player

	panelRect image.Rectangle
	closeRect image.Rectangle
}

var _ Dialog = (*Bots)(nil)

// NewBots builds the picker. volume is the app audio setting ("25",
// "silent", …). backend may be nil, in which case the demo backend is used.
func NewBots(com *common.Common, profile colorprofile.Profile, volume string, backend bots.Backend) *Bots {
	if backend == nil {
		backend = bots.NewFake()
	}
	b := &Bots{
		com:     com,
		backend: backend,
		roster:  backend.List(),
		volume:  volume,
		color:   bots.ModeFromProfile(profile),
		width:   96,
	}
	if len(b.roster) == 0 {
		b.roster = bots.Catalog()
	}
	in := textinput.New()
	in.SetVirtualCursor(false)
	in.Placeholder = "Say something"
	in.Prompt = "> "
	if com != nil && com.Styles != nil {
		in.SetStyles(com.Styles.TextInput)
	}
	in.Focus()
	b.input = in
	return b
}

// ID implements Dialog.
func (b *Bots) ID() string { return BotsID }

// InitCmd starts the idle animation.
func (b *Bots) InitCmd() tea.Cmd { return b.tickCmd() }

// SelectedID is the highlighted bot.
func (b *Bots) SelectedID() string {
	if len(b.roster) == 0 {
		return ""
	}
	return b.roster[b.index].ID
}

// InChat reports whether the dialog is in the conversation view.
func (b *Bots) InChat() bool { return b.mode == botsChat }

// Muted reports the in-dialog mute. App-wide silent volume and NO_SOUND
// also suppress cues.
func (b *Bots) Muted() bool { return b.muted }

// SetWidth forces the layout width used before the first Draw.
func (b *Bots) SetWidth(w int) { b.width = w }

// HandleMsg implements Dialog.
func (b *Bots) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.MouseClickMsg:
		if msg.Button != uv.MouseLeft {
			return nil
		}
		pt := image.Pt(msg.X, msg.Y)
		if pt.In(b.closeRect) || (!b.panelRect.Empty() && !pt.In(b.panelRect)) {
			b.stop()
			return ActionClose{}
		}
		return nil
	case botsTickMsg:
		b.clock++
		b.tickPillo()
		return ActionCmd{Cmd: b.tickCmd()}
	case writerTickMsg:
		if b.mode != botsChat {
			return nil
		}
		if b.reveal() {
			return ActionCmd{Cmd: b.writerCmd()}
		}
		b.writing = false
		return nil
	case streamOpenMsg:
		if msg.gen != b.gen {
			return nil
		}
		if msg.err != nil || msg.ch == nil {
			return b.fail()
		}
		b.stream = msg.ch
		return ActionCmd{Cmd: waitChunk(msg.gen, msg.ch)}
	case streamChunkMsg:
		if msg.gen != b.gen {
			return nil
		}
		if msg.done {
			b.streaming = false
			b.stream = nil
			if b.behind() && !b.writing {
				b.writing = true
				return ActionCmd{Cmd: b.writerCmd()}
			}
			return nil
		}
		fresh := b.lastBotEmpty()
		b.appendChunk(msg)
		var cues []string
		if fresh {
			bot := b.current()
			cues = []string{bot.ChatCue(b.gen + 1)}
			if msg.boop {
				cues = append(cues, bot.Cues.Boop)
			}
		}
		next := waitChunk(msg.gen, b.stream)
		if !b.writing {
			b.writing = true
			next = batchCmd(next, b.writerCmd())
		}
		return b.play(cues, next)
	case tea.KeyPressMsg:
		return b.handleKey(msg)
	default:
		return nil
	}
}

func (b *Bots) handleKey(msg tea.KeyPressMsg) Action {
	if b.mode == botsChat {
		return b.chatKey(msg)
	}
	switch msg.String() {
	case "esc", "alt+esc":
		b.stop()
		return ActionClose{}
	case "enter":
		return b.openChat()
	case "left", "h":
		b.move(bots.DirLeft)
	case "right", "l":
		b.move(bots.DirRight)
	case "up", "k":
		b.move(bots.DirUp)
	case "down", "j":
		b.move(bots.DirDown)
	case "m":
		return b.toggleMute()
	}
	return nil
}

func (b *Bots) chatKey(msg tea.KeyPressMsg) Action {
	switch msg.String() {
	case "esc", "alt+esc":
		b.back()
		return nil
	case "enter":
		return b.submit()
	case "ctrl+m":
		return b.toggleMute()
	default:
		var cmd tea.Cmd
		b.input, cmd = b.input.Update(msg)
		if cmd != nil {
			return ActionCmd{Cmd: cmd}
		}
		return nil
	}
}

func (b *Bots) move(dir bots.Direction) {
	prev := b.index
	cols, _ := b.layout()
	b.index = bots.Move(b.index, len(b.roster), cols, dir)
	if b.index != prev {
		b.playPilloSelect()
	}
}

func (b *Bots) tickPillo() {
	b.ensurePillo()
	if !b.syncPillo() {
		b.pillo.Advance()
	}
}

func (b *Bots) ensurePillo() {
	if b.pillo == nil {
		b.pillo = bots.NewPlayer(bots.PilloClip())
	}
}

func (b *Bots) syncPillo() bool {
	b.ensurePillo()
	if bots.ReduceMotion() {
		if b.pillo.Name() != "rest" {
			b.pillo.Play("rest")
			return true
		}
		return false
	}
	if b.pillo.Name() == "select" && !b.pillo.Done() {
		return false
	}
	want := "idle"
	if b.mode == botsChat && b.current().ID == "pillo" {
		switch {
		case b.streaming && b.lastBotEmpty():
			want = "skeptical"
		case b.behind() || b.writing:
			want = "talk"
		}
	}
	if b.pillo.Name() != want {
		b.pillo.Play(want)
		return true
	}
	return false
}

func (b *Bots) playPilloSelect() {
	if b.current().ID != "pillo" || bots.ReduceMotion() {
		return
	}
	b.ensurePillo()
	b.pillo.Play("select")
}

func (b *Bots) spriteArt(bot bots.Bot, selected, squint bool) (string, int) {
	if bot.ID == "pillo" {
		b.ensurePillo()
		rows := b.pillo.Rows()
		return bots.RenderGrid(rows, bots.PilloBody, bots.PilloEye, b.color, selected), len(rows[0])
	}
	blink, bob := pose(bot, b.clock, squint)
	return bots.Render(bot.Sprite, blink, b.color, bob, selected), bots.SpriteW
}

func (b *Bots) layout() (cols int, list bool) {
	return bots.Columns(b.layoutWidth(), len(b.roster))
}

func (b *Bots) layoutWidth() int {
	w := b.width
	if w <= 0 {
		w = 80
	}
	return w
}

func (b *Bots) openChat() Action {
	if len(b.roster) == 0 {
		return nil
	}
	bot := b.current()
	b.mode = botsChat
	b.lines = []chatLine{{
		fromBot: true,
		emote:   bot.Voice.GreetEmote.Glyph,
		squint:  bot.Voice.GreetEmote.Squint,
		full:    bot.Voice.Greeting,
	}}
	b.writing = true
	b.streaming = false
	b.input.SetValue("")
	b.input.Focus()
	b.playPilloSelect()
	cues := []string{bot.Cues.Select}
	if bot.Cues.Boop != "" {
		cues = append(cues, bot.Cues.Boop)
	}
	return b.play(cues, b.writerCmd())
}

func (b *Bots) submit() Action {
	text := strings.TrimSpace(b.input.Value())
	if text == "" || b.streaming {
		return nil
	}
	b.input.SetValue("")
	b.lines = append(b.lines, chatLine{
		full:  text,
		shown: utf8.RuneCountInString(text),
	})
	b.lines = append(b.lines, chatLine{fromBot: true})
	b.streaming = true
	bot := b.current()
	return b.play([]string{bot.ChatCue(len(text))}, b.sendCmd(text))
}

func (b *Bots) fail() Action {
	b.streaming = false
	if n := len(b.lines); n > 0 && b.lines[n-1].fromBot {
		b.lines[n-1].full = "Demo line dropped."
		b.lines[n-1].shown = utf8.RuneCountInString(b.lines[n-1].full)
		b.lines[n-1].emote = ""
	}
	return b.play([]string{b.current().ErrorCue(0)}, nil)
}

func (b *Bots) toggleMute() Action {
	b.muted = !b.muted
	// Play the click only when unmuting, so mute itself goes quiet.
	if b.muted {
		return nil
	}
	return b.play([]string{"quick-notify-02"}, nil)
}

func (b *Bots) back() {
	b.stop()
	b.mode = botsPicker
	b.lines = nil
	b.writing = false
}

func (b *Bots) stop() {
	if b.cancel != nil {
		b.cancel()
		b.cancel = nil
	}
	b.stream = nil
	b.gen++
	b.streaming = false
}

func (b *Bots) lastBotEmpty() bool {
	if len(b.lines) == 0 {
		return true
	}
	ln := b.lines[len(b.lines)-1]
	return ln.fromBot && ln.full == ""
}

func (b *Bots) appendChunk(msg streamChunkMsg) {
	if len(b.lines) == 0 || !b.lines[len(b.lines)-1].fromBot {
		b.lines = append(b.lines, chatLine{fromBot: true})
	}
	ln := &b.lines[len(b.lines)-1]
	ln.full += msg.text
	if msg.emote != "" && ln.emote == "" {
		ln.emote = msg.emote
		for _, em := range b.current().Voice.Emotes {
			if em.Glyph == msg.emote {
				ln.squint = em.Squint
				break
			}
		}
	}
}

func (b *Bots) sendCmd(text string) tea.Cmd {
	if b.cancel != nil {
		b.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	b.cancel = cancel
	b.gen++
	gen := b.gen
	id := b.current().ID
	backend := b.backend
	return func() tea.Msg {
		ch, err := backend.Send(ctx, id, text)
		return streamOpenMsg{gen: gen, ch: ch, err: err}
	}
}

func batchCmd(a, b tea.Cmd) tea.Cmd {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	default:
		return tea.Batch(a, b)
	}
}

func waitChunk(gen int, ch <-chan bots.Reply) tea.Cmd {
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		r, ok := <-ch
		if !ok {
			return streamChunkMsg{gen: gen, done: true}
		}
		return streamChunkMsg{gen: gen, text: r.Text, emote: r.Emote, boop: r.Boop}
	}
}

func (b *Bots) play(cues []string, cmd tea.Cmd) Action {
	kept := b.keepCues(cues)
	if len(kept) == 0 && cmd == nil {
		return nil
	}
	if len(kept) == 0 {
		return ActionCmd{Cmd: cmd}
	}
	return ActionPlayCue{Types: kept, Cmd: cmd}
}

func (b *Bots) keepCues(cues []string) []string {
	if !b.soundOn() {
		return nil
	}
	out := make([]string, 0, len(cues))
	for _, c := range cues {
		if c != "" {
			out = append(out, c)
		}
	}
	return out
}

func (b *Bots) soundOn() bool {
	if b.muted {
		return false
	}
	return bots.Audible(audio.NormalizeVolume(b.volume))
}

func (b *Bots) tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return botsTickMsg{}
	})
}

func (b *Bots) writerCmd() tea.Cmd {
	d := 30
	if bot := b.current(); bot.Voice.TypeMillis > 0 {
		d = bot.Voice.TypeMillis
	}
	return tea.Tick(time.Duration(d)*time.Millisecond, func(time.Time) tea.Msg {
		return writerTickMsg{}
	})
}

func (b *Bots) reveal() bool {
	if !b.behind() {
		return false
	}
	ln := &b.lines[len(b.lines)-1]
	step := b.current().Voice.TypeStep
	if step < 1 {
		step = 1
	}
	ln.shown += step
	n := utf8.RuneCountInString(ln.full)
	if ln.shown > n {
		ln.shown = n
	}
	return b.behind()
}

func (b *Bots) behind() bool {
	if len(b.lines) == 0 {
		return false
	}
	ln := b.lines[len(b.lines)-1]
	if !ln.fromBot {
		return false
	}
	return ln.shown < utf8.RuneCountInString(ln.full)
}

func (b *Bots) current() bots.Bot {
	if len(b.roster) == 0 {
		return bots.Bot{}
	}
	if b.index < 0 || b.index >= len(b.roster) {
		b.index = 0
	}
	return b.roster[b.index]
}

// Draw implements Dialog.
func (b *Bots) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	if b.com == nil || b.com.Styles == nil {
		return nil
	}
	t := b.com.Styles
	frame := t.Dialog.View.GetHorizontalFrameSize()
	total := area.Dx()
	if total > 100 {
		total = 100
	}
	if total < frame+8 {
		total = frame + 8
	}
	inner := total - frame
	if inner < 8 {
		inner = 8
	}
	b.width = inner

	var body string
	var cur *tea.Cursor
	if b.mode == botsChat {
		body, cur = b.chatBody(inner)
	} else {
		body = b.pickerBody(inner)
	}
	title := b.titleLine(inner)
	help := b.helpLine(inner)
	content := title + "\n" + body + "\n" + help
	view := t.Dialog.View.Width(total).Render(content)

	vw, vh := lipgloss.Size(view)
	vw = min(vw, area.Dx())
	vh = min(vh, area.Dy())
	center := common.CenterRect(area, vw, vh)
	b.panelRect = center

	if cur != nil {
		inputH := lipgloss.Height(b.input.View())
		if inputH < 1 {
			inputH = 1
		}
		cur.X += t.Dialog.View.GetBorderLeftSize() + t.Dialog.View.GetPaddingLeft()
		cur.Y += t.Dialog.View.GetBorderTopSize() + t.Dialog.View.GetPaddingTop()
		// The input is the last line of the chat body.
		cur.Y += lipgloss.Height(title) + lipgloss.Height(body) - inputH
	}
	DrawCenterCursor(scr, area, view, cur)

	closeText := "[X]"
	if inner < 20 {
		closeText = "X"
	}
	cw := ansi.StringWidth(closeText)
	cx := center.Min.X + vw - cw - 1
	if cx < center.Min.X {
		cx = center.Min.X
	}
	b.closeRect = image.Rect(cx, center.Min.Y, cx+cw, center.Min.Y+1)
	label := t.Dialog.FileBrowser.Close.Render(closeText)
	uv.NewStyledString(label).Draw(scr, b.closeRect)
	return cur
}

func (b *Bots) titleLine(width int) string {
	name := "BOTS"
	if b.mode == botsChat {
		name = strings.ToUpper(b.current().Name)
	}
	badge := "DEMO"
	if b.color != bots.ColorNone {
		badge = bots.Paint("DEMO", bots.PixOrange, b.color)
	}
	state := "SND"
	if !b.soundOn() {
		state = "MUTE"
	}
	right := badge + " " + state
	gap := width - ansi.StringWidth(name) - ansi.StringWidth(right)
	if gap < 1 {
		return ansi.Truncate(name+" "+right, width, "")
	}
	return name + strings.Repeat(" ", gap) + right
}

func (b *Bots) helpLine(width int) string {
	text := "←/→ hjkl move   enter chat   m mute   esc close"
	if b.mode == botsChat {
		text = "enter send   esc back   ctrl+m mute"
	}
	if ansi.StringWidth(text) > width {
		return ansi.Truncate(text, width, "")
	}
	return text
}

func (b *Bots) pickerBody(width int) string {
	_, list := b.layout()
	if list {
		return b.renderList(width)
	}
	cols, _ := b.layout()
	var rows []string
	for i := 0; i < len(b.roster); i += cols {
		end := min(i+cols, len(b.roster))
		cards := make([]string, 0, end-i)
		for j := i; j < end; j++ {
			cards = append(cards, b.card(j))
		}
		rows = append(rows, joinCards(cards))
	}
	return strings.Join(rows, "\n")
}

func (b *Bots) card(i int) string {
	bot := b.roster[i]
	sel := i == b.index
	art, artW := b.spriteArt(bot, sel, false)
	name := centerTo(bots.Paint(bots.ChunkyName(bot.Name), bot.Color, b.color), artW)
	mark := " "
	if sel && (b.clock/4)%2 == 0 {
		mark = "▶"
	}
	cur := centerTo(mark, artW)
	return art + "\n" + name + "\n" + cur
}

func (b *Bots) renderList(width int) string {
	var lines []string
	for i, bot := range b.roster {
		mark := " "
		if i == b.index {
			if (b.clock/4)%2 == 0 {
				mark = "▶"
			} else {
				mark = "·"
			}
		}
		line := mark + " " + bot.Name + "  " + bot.Tagline
		if ansi.StringWidth(line) > width {
			line = ansi.Truncate(line, width, "")
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (b *Bots) chatBody(width int) (string, *tea.Cursor) {
	bot := b.current()
	squint := false
	if n := len(b.lines); n > 0 && b.lines[n-1].squint {
		squint = true
	}
	art, artW := b.spriteArt(bot, true, squint)
	box := b.messageBox(width)
	var view string
	if width >= artW+20 {
		box = b.messageBox(width - artW - 1)
		view = lipgloss.JoinHorizontal(lipgloss.Top, art, " ", box)
	} else {
		view = art + "\n" + box
	}
	promptW := lipgloss.Width(b.input.Prompt)
	b.input.SetWidth(max(1, width-promptW-1))
	view += "\n" + b.input.View()
	var cur *tea.Cursor
	if c := b.input.Cursor(); c != nil {
		copied := *c
		cur = &copied
	}
	return view, cur
}

func (b *Bots) messageBox(width int) string {
	if width < 6 {
		width = 6
	}
	inner := width - 2
	text := b.transcript(inner)
	top := "╔" + strings.Repeat("═", width-2) + "╗"
	bottom := "╚" + strings.Repeat("═", width-2) + "╝"
	var out strings.Builder
	out.WriteString(top)
	for _, ln := range strings.Split(text, "\n") {
		if ansi.StringWidth(ln) > inner {
			ln = ansi.Truncate(ln, inner, "")
		}
		pad := inner - ansi.StringWidth(ln)
		if pad < 0 {
			pad = 0
		}
		out.WriteString("\n║")
		out.WriteString(ln)
		out.WriteString(strings.Repeat(" ", pad))
		out.WriteString("║")
	}
	out.WriteString("\n")
	out.WriteString(bottom)
	return out.String()
}

func (b *Bots) transcript(inner int) string {
	var lines []string
	for i, ln := range b.lines {
		text := shownText(ln)
		revealing := ln.fromBot && ln.shown < utf8.RuneCountInString(ln.full)
		if revealing && i == len(b.lines)-1 {
			text += "▌"
		}
		if ln.fromBot && text == "" && b.streaming && i == len(b.lines)-1 {
			text = "…"
		}
		prefix := "YOU "
		if ln.fromBot {
			prefix = ""
			if ln.emote != "" {
				prefix = ln.emote + " "
			}
		}
		lines = append(lines, wrapText(prefix+text, inner)...)
	}
	const maxLines = 8
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	for len(lines) < 3 {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func shownText(ln chatLine) string {
	rs := []rune(ln.full)
	if ln.shown >= len(rs) {
		return ln.full
	}
	if ln.shown <= 0 {
		return ""
	}
	return string(rs[:ln.shown])
}

func pose(bot bots.Bot, clock int, forceBlink bool) (blink bool, bob int) {
	step := bot.Voice.IdleMillis / 100
	if step < 2 {
		step = 2
	}
	phase := clock % step
	if bot.Voice.Blink && (forceBlink || phase == step-1) {
		blink = true
	}
	if bot.Voice.Bob && phase%2 == 1 {
		bob = 1
	}
	return blink, bob
}

func joinCards(cards []string) string {
	if len(cards) == 0 {
		return ""
	}
	if len(cards) == 1 {
		return cards[0]
	}
	parts := make([]string, 0, len(cards)*2-1)
	gap := strings.Repeat(" ", bots.CardGap)
	for i, card := range cards {
		if i > 0 {
			parts = append(parts, gap)
		}
		parts = append(parts, card)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func centerTo(s string, width int) string {
	w := ansi.StringWidth(s)
	if w >= width {
		return s
	}
	left := (width - w) / 2
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", width-w-left)
}

func wrapText(s string, width int) []string {
	if width < 1 || s == "" {
		return []string{s}
	}
	var lines []string
	var cur string
	curW := 0
	for _, r := range s {
		rw := ansi.StringWidth(string(r))
		if rw < 1 {
			rw = 1
		}
		if curW+rw > width && cur != "" {
			lines = append(lines, cur)
			cur = ""
			curW = 0
		}
		cur += string(r)
		curW += rw
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}
