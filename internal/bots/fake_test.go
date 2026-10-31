package bots

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCatalogPersonalities(t *testing.T) {
	t.Parallel()
	roster := Catalog()
	require.Len(t, roster, 5)
	require.Equal(t, []string{"dingus", "brobro", "pillo", "luvah", "gnarf"}, ids(roster))

	dingus, _ := ByID("dingus")
	bro, _ := ByID("brobro")
	pillo, _ := ByID("pillo")
	luvah, _ := ByID("luvah")
	gnarf, _ := ByID("gnarf")

	require.Less(t, bro.Voice.TypeMillis, dingus.Voice.TypeMillis)
	require.Greater(t, bro.Voice.TypeStep, dingus.Voice.TypeStep)
	require.True(t, bro.Voice.Bob)
	require.False(t, dingus.Voice.Bob)
	require.Greater(t, pillo.Voice.IdleMillis, luvah.Voice.IdleMillis)
	require.Greater(t, luvah.Voice.TypeMillis, gnarf.Voice.TypeMillis)

	require.Empty(t, dingus.Cues.Boop)
	require.Equal(t, "quick-notify-03", gnarf.Cues.Boop)
	require.NotEmpty(t, dingus.Cues.Chat)
	require.NotEmpty(t, dingus.Cues.Errors)
	require.NotEmpty(t, dingus.Cues.Select)

	require.Contains(t, emoteGlyphs(luvah), "♥")
	require.Contains(t, emoteGlyphs(gnarf), "♪")
	require.Contains(t, emoteGlyphs(bro), "↗")
	require.True(t, dingus.Voice.GreetEmote.Squint)

	for _, line := range pillo.Voice.Lines {
		require.NotContains(t, line, "!")
		require.Less(t, len(line), 24)
	}
	for _, line := range gnarf.Voice.Lines {
		stripped := strings.ReplaceAll(line, "{msg}", "")
		require.Equal(t, strings.ToUpper(stripped), stripped)
	}
	for _, line := range bro.Voice.Lines {
		require.Contains(t, line, "!!")
	}
}

func TestFakeSendStreamsPersonalityAndCancels(t *testing.T) {
	t.Parallel()
	f := NewFake().Instant()
	var _ Backend = f

	_, err := f.Send(context.Background(), "nope", "hi")
	require.Error(t, err)

	ch, err := f.Send(context.Background(), "dingus", "hello there")
	require.NoError(t, err)
	text, emote, boop := readAll(t, ch)
	require.Contains(t, text, "hello there")
	require.Contains(t, text, "Understood")
	require.NotEmpty(t, emote)
	require.False(t, boop)

	ch, err = f.Send(context.Background(), "pillo", "status")
	require.NoError(t, err)
	text, _, _ = readAll(t, ch)
	require.Contains(t, text, "status")
	require.Less(t, len(strings.Fields(text)), 6)

	ch, err = f.Send(context.Background(), "gnarf", "play a song")
	require.NoError(t, err)
	text, emote, boop = readAll(t, ch)
	require.Contains(t, text, "PLAY A SONG")
	require.NotContains(t, text, "play a song")
	require.True(t, boop)
	require.NotEmpty(t, emote)

	ch, err = f.Send(context.Background(), "luvah", "thanks")
	require.NoError(t, err)
	text, emote, _ = readAll(t, ch)
	require.Contains(t, text, "thanks")
	require.NotEmpty(t, emote)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	slow := NewFake()
	ch, err = slow.Send(ctx, "brobro", "go")
	require.NoError(t, err)
	done := make(chan struct{})
	go func() {
		for range ch {
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("canceled send did not close")
	}
}

func TestMoveAndColumns(t *testing.T) {
	t.Parallel()
	require.Equal(t, 1, Move(0, 5, 5, DirRight))
	require.Equal(t, 0, Move(4, 5, 5, DirRight))
	require.Equal(t, 4, Move(0, 5, 5, DirLeft))
	require.Equal(t, 2, Move(0, 5, 2, DirDown))
	require.Equal(t, 4, Move(2, 5, 2, DirDown))
	require.Equal(t, 0, Move(4, 5, 2, DirDown))
	require.Equal(t, 4, Move(0, 5, 2, DirUp))

	cols, list := Columns(96, 5)
	require.Equal(t, 5, cols)
	require.False(t, list)
	cols, list = Columns(70, 5)
	require.Equal(t, 4, cols)
	require.False(t, list)
	cols, list = Columns(54, 5)
	require.Equal(t, 3, cols)
	require.False(t, list)
	cols, list = Columns(12, 5)
	require.Equal(t, 1, cols)
	require.True(t, list)
}

func ids(roster []Bot) []string {
	out := make([]string, len(roster))
	for i, b := range roster {
		out[i] = b.ID
	}
	return out
}

func emoteGlyphs(b Bot) string {
	var glyphs []string
	for _, em := range b.Voice.Emotes {
		glyphs = append(glyphs, em.Glyph)
	}
	glyphs = append(glyphs, b.Voice.GreetEmote.Glyph)
	return strings.Join(glyphs, " ")
}

func readAll(t *testing.T, ch <-chan Reply) (text, emote string, boop bool) {
	t.Helper()
	var b strings.Builder
	n := 0
	for r := range ch {
		b.WriteString(r.Text)
		if n == 0 {
			emote = r.Emote
			boop = r.Boop
		} else {
			require.Empty(t, r.Emote)
			require.False(t, r.Boop)
		}
		n++
	}
	require.GreaterOrEqual(t, n, 1)
	return b.String(), emote, boop
}
