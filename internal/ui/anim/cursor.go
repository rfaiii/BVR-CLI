package anim

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// CursorEffect renders a lightweight cursor trail in the editor/status area.
// It is intentionally dependency-free beyond Bubble Tea/Lip Gloss so it can
// live inside the main app without pulling in incompatible graphics stacks.
type CursorEffect struct {
	width   int
	tick    time.Time
	enabled bool
	history []cursorPoint
	maxHist int
	spacing int
	styles  cursorStyles
}

type cursorPoint struct {
	x    int
	tick time.Time
}

type cursorStyles struct {
	head lipgloss.Style
	tail lipgloss.Style
}

func defaultCursorStyles() cursorStyles {
	return cursorStyles{
		head: lipgloss.NewStyle().Foreground(lipgloss.Color("#3bf66b")).Bold(true),
		tail: lipgloss.NewStyle().Foreground(lipgloss.Color("#1a7a3a")),
	}
}

// NewCursorEffect creates a cursor trail renderer.
func NewCursorEffect(width int) *CursorEffect {
	return &CursorEffect{
		width:   width,
		maxHist: 12,
		spacing: 2,
		styles:  defaultCursorStyles(),
	}
}

// SetEnabled toggles the effect.
func (c *CursorEffect) SetEnabled(on bool) { c.enabled = on }

// Resize updates the available cell width.
func (c *CursorEffect) Resize(width int) {
	if width < 0 {
		width = 0
	}
	c.width = width
}

// Tick is a Bubble Tea tick message for cursor animation frames.
type TickMsg struct{ time.Time }

// Tick starts the cursor animation loop.
func (c *CursorEffect) Tick() tea.Cmd {
	if !c.enabled {
		return nil
	}
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg{t}
	})
}

// Update advances the cursor trail state.
func (c *CursorEffect) Update(msg TickMsg) tea.Cmd {
	if !c.enabled {
		return nil
	}
	c.tick = msg.Time
	if c.width <= 0 {
		return c.Tick()
	}
	// Advance the head.
	c.history = append(c.history, cursorPoint{x: 0, tick: msg.Time})
	if len(c.history) > c.maxHist {
		c.history = c.history[len(c.history)-c.maxHist:]
	}
	return c.Tick()
}

// Render returns the cursor effect string for the given editor width.
func (c *CursorEffect) Render(editorWidth int) string {
	if !c.enabled || c.width <= 0 || editorWidth <= 0 {
		return ""
	}
	if len(c.history) == 0 {
		return ""
	}
	step := max(1, c.spacing)
	positions := c.samplePositions(editorWidth, step)
	if len(positions) == 0 {
		return ""
	}
	fields := make([]string, editorWidth)
	for i := range fields {
		fields[i] = " "
	}
	for i, x := range positions {
		if x < 0 || x >= editorWidth {
			continue
		}
		if i == 0 {
			fields[x] = c.styles.head.Render("▸")
			continue
		}
		ch := "·"
		switch i % 3 {
		case 1:
			ch = "·"
		case 2:
			ch = "•"
		default:
			ch = "∙"
		}
		fields[x] = c.styles.tail.Render(ch)
	}
	return strings.Join(fields, "")
}

func (c *CursorEffect) samplePositions(editorWidth, step int) []int {
	if len(c.history) == 0 {
		return nil
	}
	head := c.history[len(c.history)-1].x
	positions := []int{head}
	if len(c.history) == 1 {
		return positions
	}
	tail := c.history[0].x
	delta := tail - head
	steps := len(c.history) - 1
	if steps <= 0 {
		return positions
	}
	for i := 1; i < len(c.history); i++ {
		idx := float64(i) / float64(steps)
		x := int(float64(head) + float64(delta)*idx)
		x = clamp(x, 0, editorWidth-1)
		if i%step == 0 {
			positions = append(positions, x)
		}
	}
	return positions
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// CursorEffectForTesting exposes test-oriented helpers.
type CursorEffectForTesting struct {
	*CursorEffect
}

// NewCursorEffectForTesting creates an effect with deterministic defaults.
func NewCursorEffectForTesting(width int) *CursorEffectForTesting {
	return &CursorEffectForTesting{NewCursorEffect(width)}
}

// Advance advances history with synthetic positions for assertions.
func (c *CursorEffectForTesting) Advance(positions ...int) {
	for _, x := range positions {
		c.history = append(c.history, cursorPoint{x: x, tick: time.Now()})
		if len(c.history) > c.maxHist {
			c.history = c.history[len(c.history)-c.maxHist:]
		}
	}
}

// SamplePositions exposes sampling for assertions.
func (c *CursorEffectForTesting) SamplePositions(editorWidth, step int) []int {
	return c.CursorEffect.samplePositions(editorWidth, step)
}

// String renders the effect with enabled forced on for assertions.
func (c *CursorEffectForTesting) String(editorWidth int) string {
	c.CursorEffect.enabled = true
	return c.CursorEffect.Render(editorWidth)
}

// RenderedRunes exposes the underlying rune map for precise assertions.
func (c *CursorEffectForTesting) RenderedRunes(editorWidth int) []rune {
	return []rune(c.String(editorWidth))
}

// MustRuneAt asserts the rune at a cell index for test failures.
func (c *CursorEffectForTesting) MustRuneAt(editorWidth, index int) rune {
	runes := c.RenderedRunes(editorWidth)
	if index < 0 || index >= len(runes) {
		return 0
	}
	return runes[index]
}

// CharCount returns the number of runes in the rendered effect.
func (c *CursorEffectForTesting) CharCount(editorWidth int) int {
	return len(c.RenderedRunes(editorWidth))
}

// RenderCount returns the number of non-space runes in the rendered effect.
func (c *CursorEffectForTesting) RenderCount(editorWidth int) int {
	count := 0
	for _, r := range c.RenderedRunes(editorWidth) {
		if r != ' ' {
			count++
		}
	}
	return count
}

func (c *CursorEffectForTesting) MustContainRune(editorWidth int, expected rune) {
	runes := c.RenderedRunes(editorWidth)
	found := false
	for _, r := range runes {
		if r == expected {
			found = true
			break
		}
	}
	if !found {
		panic(fmt.Sprintf("expected effect to contain rune %q in width=%d; got %q", expected, editorWidth, string(runes)))
	}
}

func (c *CursorEffectForTesting) MustNotContainRune(editorWidth int, unexpected rune) {
	runes := c.RenderedRunes(editorWidth)
	for _, r := range runes {
		if r == unexpected {
			panic(fmt.Sprintf("expected effect to not contain rune %q in width=%d; got %q", unexpected, editorWidth, string(runes)))
		}
	}
}

func (c *CursorEffectForTesting) MustRenderExactly(editorWidth int, expected string) {
	got := c.String(editorWidth)
	if got != expected {
		panic(fmt.Sprintf("expected %q, got %q", expected, got))
	}
}
