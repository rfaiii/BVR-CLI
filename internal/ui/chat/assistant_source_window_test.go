package chat

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/richavery/bvr-cli/internal/message"
	"github.com/richavery/bvr-cli/internal/ui/styles"
	"github.com/stretchr/testify/require"
)

// The tests here cover the SOURCE window (thinkingWindow / thinkingTailCut),
// which cuts the reasoning text before glamour sees it. That is distinct from
// the rendered-line window covered in assistant_thinking_window_test.go, whose
// fixtures are one soft-broken paragraph and therefore take the untouched
// full-render path by design. These cover the multi-block shapes the cut rules
// were written for.

// sourceWindowMessage builds a still-thinking assistant message whose
// reasoning is exactly src.
func sourceWindowMessage(id, src string) *message.Message {
	return &message.Message{
		ID:   id,
		Role: message.Assistant,
		Parts: []message.ContentPart{message.ReasoningContent{
			Thinking:  src,
			StartedAt: testStartedAt,
		}},
	}
}

// sourceWindowBlocks is sized so that even the tail-window view — whose
// window wants maxExpandedThinkingTailLines*2 source lines — has to cut
// something, and so the first block is far outside any window. Five lines
// per block on average, so 200 blocks clear the ~800-line gate.
const sourceWindowBlocks = 200

func sourceWindowTag(i int) string { return fmt.Sprintf("B%03d", i) }

// sourceWindowSource builds markdown blocks covering the constructs the tail
// cut has to reason about — paragraphs, headings, loose and tight bullet
// lists, ordered lists, fenced code, tables, block quotes — each block
// separated from its neighbour by a blank line. Every line starts with its
// block's tag so an assertion can name the block a window begins in.
func sourceWindowSource() string {
	var blocks []string
	for i := 1; i <= sourceWindowBlocks; i++ {
		tag := sourceWindowTag(i)
		var b strings.Builder
		switch i % 7 {
		case 0:
			b.WriteString("```go\n")
			b.WriteString(tag + " func main() {\n")
			b.WriteString(tag + "   println(\"inside a fenced block\")\n")
			b.WriteString(tag + " }\n```\n")
		case 1:
			b.WriteString(tag + " first line of this paragraph\n")
			b.WriteString(tag + " second line of the paragraph\n")
			b.WriteString(tag + " third line, which ends the block\n")
		case 2:
			// Loose list: blank lines between sibling items, the one list
			// shape whose interior boundaries may be cut.
			for j := 1; j <= 4; j++ {
				fmt.Fprintf(&b, "- %s item %d with enough words to fill a line\n\n", tag, j)
			}
			b.WriteString("- " + tag + " closing item of the loose list\n")
		case 3:
			b.WriteString("# " + tag + " heading\n")
			b.WriteString(tag + " body line under the heading\n")
			b.WriteString(tag + " another body line right here\n")
		case 4:
			// Ordered items in one tight block: no interior separator, and
			// the boundary in front of it must be rejected.
			for j := 1; j <= 4; j++ {
				fmt.Fprintf(&b, "%d. %s ordered item\n", j, tag)
			}
		case 5:
			b.WriteString("| " + tag + " | left | right |\n")
			b.WriteString("|---|---|---|\n")
			b.WriteString("| " + tag + " | a | b |\n")
		case 6:
			b.WriteString("> " + tag + " quoted line one\n")
			b.WriteString("> quoted line two for the same block\n")
			b.WriteString("> " + tag + " quoted line three\n")
		}
		blocks = append(blocks, b.String())
	}
	return strings.Join(blocks, "\n")
}

// stripTail returns the last k display lines of a rendered section with
// trailing space padding removed. The bordered ThinkingBox sizes itself to the
// widest line it was handed, so a window and the whole document can differ in
// padding while showing identical content; the padding is not what these tests
// are about, the reasoning text is.
func stripTail(t *testing.T, rendered string, k int) []string {
	t.Helper()
	lines := strings.Split(ansi.Strip(rendered), "\n")
	require.GreaterOrEqual(t, len(lines), k, "render is too short to compare %d tail lines", k)
	out := make([]string, 0, k)
	for i := len(lines) - k; i < len(lines); i++ {
		out = append(out, strings.TrimRight(lines[i], " "))
	}
	return out
}

// TestSourceWindow_MultiBlockTailMatchesFullRender is the central safety
// claim: cutting the source at a markdown boundary and rendering only what is
// left must show the user exactly what rendering the whole document shows.
func TestSourceWindow_MultiBlockTailMatchesFullRender(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()
	src := sourceWindowSource()
	require.Greater(t, 1+strings.Count(src, "\n"), maxExpandedThinkingTailLines*2,
		"fixture must be long enough that both windowed modes have to cut")

	const width = 133
	full := NewAssistantMessageItem(&sty, sourceWindowMessage("sw-full", src)).(*AssistantMessageItem)
	full.thinkingViewMode = thinkingFullExpanded
	_ = full.RawRender(width)
	require.Zero(t, full.thinkingWindowStart, "full expansion must never window the source")

	for _, tc := range []struct {
		name string
		mode thinkingViewMode
		tail int
		ceil int
		// windowMax bounds the cut window as a fraction of the document:
		// the point of windowing is that the window stays small relative to
		// the trace, but block granularity sets the floor, so this is a
		// proportion rather than the drift limit itself.
		windowMax int
	}{
		{
			name:      "collapsed",
			mode:      thinkingCollapsed,
			tail:      maxCollapsedThinkingHeight,
			ceil:      maxCollapsedThinkingHeight + 5,
			windowMax: 100,
		},
		{
			name:      "tail-window",
			mode:      thinkingTailWindow,
			tail:      maxExpandedThinkingTailLines - 5,
			ceil:      maxExpandedThinkingTailLines + 5,
			windowMax: 500,
		},
	} {
		item := NewAssistantMessageItem(&sty, sourceWindowMessage("sw-"+tc.name, src)).(*AssistantMessageItem)
		item.thinkingViewMode = tc.mode
		_ = item.RawRender(width)

		require.Positive(t, item.thinkingWindowStart,
			"%s must have cut a source window instead of rendering all %d lines",
			tc.name, 1+strings.Count(src, "\n"))
		require.Less(t, item.thinkingWindowLines, tc.windowMax,
			"%s window must stay small against a %d-line document; got %d lines",
			tc.name, 1+strings.Count(src, "\n"), item.thinkingWindowLines)
		require.LessOrEqual(t, item.thinkingBoxHeight, tc.ceil,
			"%s box must stay bounded; got %d", tc.name, item.thinkingBoxHeight)
		require.GreaterOrEqual(t, item.thinkingBoxHeight, tc.tail,
			"%s box must show the whole window; got %d", tc.name, item.thinkingBoxHeight)

		require.Equal(t,
			stripTail(t, full.thinkingSec.out, tc.tail),
			stripTail(t, item.thinkingSec.out, tc.tail),
			"%s: the windowed render's last %d lines must match the full render", tc.name, tc.tail)

		plain := ansi.Strip(item.thinkingSec.out)
		require.NotContains(t, plain, sourceWindowTag(1)+" ", "the first block must be cut away")
		require.Contains(t, plain, sourceWindowTag(sourceWindowBlocks),
			"the last block must stay visible")

		t.Logf("%s: cut at byte %d of %d, %d source lines hidden, window %d lines, box %d tall",
			tc.name, item.thinkingWindowStart, len(src),
			strings.Count(src[:item.thinkingWindowStart], "\n"),
			item.thinkingWindowLines, item.thinkingBoxHeight)
	}
}

// TestSourceWindow_SingleParagraphIsNotCut pins the deliberate limit: a
// reasoning trace with no blank lines is one paragraph, and cutting inside it
// would re-wrap every line the user can see. Those blocks keep paying the full
// render, which is why the fixtures in assistant_thinking_window_test.go still
// take the unwindowed path.
func TestSourceWindow_SingleParagraphIsNotCut(t *testing.T) {
	t.Parallel()

	item := &AssistantMessageItem{thinkingViewMode: thinkingCollapsed}
	src := thinkingSource(thinkingLongLines)

	window, hidden := item.thinkingWindow(src, item.thinkingWindowTarget())

	require.Equal(t, src, window, "a single soft-broken paragraph must not be cut")
	require.Zero(t, hidden)
	require.Zero(t, item.thinkingWindowStart)
}

// TestSourceWindow_StaysStickyWhileAppending protects the reason the cut is
// remembered at all: handing streamingMarkdown a fresh tail on every flush
// would invalidate its stable-prefix cache and put the document's cost back on
// every streaming tick.
func TestSourceWindow_StaysStickyWhileAppending(t *testing.T) {
	t.Parallel()

	item := &AssistantMessageItem{thinkingViewMode: thinkingCollapsed}
	target := item.thinkingWindowTarget()
	grown := sourceWindowSource()

	window, hidden := item.thinkingWindow(grown, target)
	require.Positive(t, item.thinkingWindowStart)
	cut := item.thinkingWindowStart
	require.LessOrEqual(t, item.thinkingWindowLines, target*4,
		"the fixture's first window must be inside the drift limit, or the next "+
			"loop proves nothing")

	// Each append is one more blank-line-separated block of two lines.
	for range 3 {
		grown += "\nappended block line one\nappended block line two\n"
		next, nextHidden := item.thinkingWindow(grown, target)
		require.Equal(t, cut, item.thinkingWindowStart,
			"appending text must not move the cut")
		require.Equal(t, hidden, nextHidden,
			"lines before a fixed cut cannot change when text is only appended")
		require.True(t, strings.HasPrefix(next, window),
			"the window must stay a prefix-extension of itself, which is what "+
				"streamingMarkdown needs to reuse its cached render")
		window = next
	}

	// Past the drift limit the window is four times what the view shows, so
	// it is re-cut nearer the tail.
	for range 20 {
		grown += "\nmore reasoning line\nand another line here\n"
	}
	item.thinkingWindow(grown, target)
	require.Greater(t, item.thinkingWindowStart, cut,
		"a window that drifted past the view must be re-cut")
	require.LessOrEqual(t, item.thinkingWindowLines, target*2+6,
		"the re-cut window must be back in proportion with the view")
}

// TestSourceWindow_RecutWhenViewModeChanges: the collapsed cut keeps ~20
// source lines, which is nowhere near enough for the tail-window view, so
// changing modes must drop it and cut earlier.
func TestSourceWindow_RecutWhenViewModeChanges(t *testing.T) {
	t.Parallel()

	src := sourceWindowSource()
	item := &AssistantMessageItem{thinkingViewMode: thinkingCollapsed}
	collapsedWindow, _ := item.thinkingWindow(src, item.thinkingWindowTarget())
	cut := item.thinkingWindowStart
	require.Positive(t, cut)

	item.thinkingViewMode = thinkingTailWindow
	tailWindow, _ := item.thinkingWindow(src, item.thinkingWindowTarget())

	require.Less(t, item.thinkingWindowStart, cut,
		"the tail-window view shows twenty times the lines, so it must re-cut earlier")
	require.True(t, strings.HasSuffix(tailWindow, collapsedWindow),
		"the wider window must contain the narrower one's text")
	require.Greater(t, len(tailWindow), len(collapsedWindow))
}

// TestSourceWindow_RewrittenSourceDropsTheCut: a retry replaces the reasoning
// text, and an offset remembered from the old text is meaningless. The offset
// has to be re-validated against the source it is applied to.
func TestSourceWindow_RewrittenSourceDropsTheCut(t *testing.T) {
	t.Parallel()

	item := &AssistantMessageItem{thinkingViewMode: thinkingCollapsed}
	src := sourceWindowSource()
	_, _ = item.thinkingWindow(src, item.thinkingWindowTarget())
	require.Positive(t, item.thinkingWindowStart)

	// A different document of the same shape, where the old offset is still
	// inside the text: length is not enough to prove the cut is valid, so
	// validity is checked structurally.
	replaced := strings.ReplaceAll(src, sourceWindowTag(7), "B999")
	_, _ = item.thinkingWindow(replaced, item.thinkingWindowTarget())

	if item.thinkingWindowStart > 0 {
		require.True(t, isThinkingWindowStart(replaced, item.thinkingWindowStart),
			"a reused cut must still be a block boundary in the current source")
	}
}

// TestThinkingTailCut_RejectsBoundaryShapedHazards exercises the predicate
// directly, because several of these shapes cannot be reached reliably through
// a whole document. before + "\n\n" + after puts the candidate at
// len(before)+2.
func TestThinkingTailCut_RejectsBoundaryShapedHazards(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		before string
		after  string
		want   bool
	}{
		{"paragraph to paragraph", "B1 third line", "B2 first line\nB2 second line\n", true},
		{"heading to paragraph", "# B1 heading", "B2 first line\n", true},
		{"complete fence stays whole", "B1 paragraph line", "```\ncode\n```\n", true},
		{"unordered siblings", "- B1 item", "- B2 item\n", true},
		{"paragraph then bullet", "B1 paragraph line", "- B2 item\n", true},
		{"nested bullet cannot lead a window", "- B1 outer item", "  - B1 nested child\n", false},
		{"ordered list cannot be split", "1. B1 item", "2. B2 item\n", false},
		{"numbering cannot restart after a paragraph", "B1 paragraph line", "3. B2 item\n", false},
		{"open fence in the window", "B1 paragraph line", "```\ncode still open\n", false},
		{"setext underline cannot lead", "B1 paragraph line", "====\nB2 after the heading\n", false},
		{"table row cannot be left behind", "B1 cell | B1 cell", "B2 paragraph line\n", false},
		{"block quote cannot be left behind", "> B1 quoted", "B2 paragraph line\n", false},
		{"indented lazy continuation cannot be left behind", "B1 paragraph\n    B2 indented continuation", "B3 next block\n", false},
	} {
		src := tc.before + "\n\n" + tc.after
		p := len(tc.before) + 2
		require.Equal(t, tc.want, isSafeThinkingTailCut(src, p), tc.name)
	}
}

// TestThinkingTailCut_WalksBackToEnoughLines checks the size rule: the window
// is the smallest tail that still holds wantLines, and nothing is cut when the
// document is barely bigger than what the view shows.
func TestThinkingTailCut_WalksBackToEnoughLines(t *testing.T) {
	t.Parallel()

	var blocks []string
	for i := 1; i <= 50; i++ {
		blocks = append(blocks, fmt.Sprintf("K%02d line one\nK%02d line two\nK%02d line three", i, i, i))
	}
	src := strings.Join(blocks, "\n\n") + "\n"

	cut := thinkingTailCut(src, 20)
	require.Positive(t, cut)
	// Each block costs four source lines (three plus the blank separator), so
	// the latest boundary whose tail holds 20 lines is five blocks from the
	// end: blocks 46 through 50.
	require.Equal(t, "K46 line one", firstLineOf(src[cut:]),
		"the window must be as small as the markdown structure allows")

	// Barely bigger than the window is not worth cutting at all.
	require.Zero(t, thinkingTailCut(strings.Join(blocks[:5], "\n\n")+"\n", 20),
		"a document of forty lines must not be cut for a twenty-line window")
	require.Zero(t, thinkingTailCut(src, 0))
	require.Zero(t, thinkingTailCut("", 20))
}

func firstLineOf(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// TestSourceWindow_HintCountsElidedSourceLines: without the source count the
// affordance would claim a 700-line trace hid a dozen lines, which is the
// opposite of what windowing did — and the number is the user's only signal
// that earlier reasoning exists.
func TestSourceWindow_HintCountsElidedSourceLines(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()
	src := sourceWindowSource()
	item := NewAssistantMessageItem(&sty, sourceWindowMessage("sw-hint", src)).(*AssistantMessageItem)
	_ = item.RawRender(133)

	require.Positive(t, item.thinkingWindowStart)
	sourceLines := 1 + strings.Count(src, "\n")
	hiddenSource := strings.Count(src[:item.thinkingWindowStart], "\n")
	hint := strings.TrimSpace(ansi.Strip(strings.Split(item.thinkingSec.out, "\n")[0]))

	var reported int
	_, err := fmt.Sscanf(hint, "… (%d lines hidden)", &reported)
	require.NoError(t, err, "collapsed hint must carry a count: %q", hint)
	require.GreaterOrEqual(t, reported, hiddenSource,
		"the hint must at least account for the %d source lines the window cut", hiddenSource)
	require.Less(t, reported, sourceLines, "and it must not claim the whole document was hidden")

	// The tail-window affordance gets the same correction.
	item.thinkingViewMode = thinkingTailWindow
	item.thinkingSec.reset()
	_ = item.RawRender(133)
	require.Positive(t, item.thinkingWindowStart)
	hint = strings.TrimSpace(ansi.Strip(strings.Split(item.thinkingSec.out, "\n")[0]))
	_, err = fmt.Sscanf(hint, "… %d earlier lines hidden", &reported)
	require.NoError(t, err, "tail-window affordance must carry a count: %q", hint)
	require.GreaterOrEqual(t, reported, strings.Count(src[:item.thinkingWindowStart], "\n"))
}
