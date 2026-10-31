package bots

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompactGridsAreSmallAndEyed(t *testing.T) {
	t.Parallel()
	want := []string{".E", "EE", "E."}
	for _, id := range []string{"dingus", "brobro", "pillo", "luvah", "gnarf"} {
		rows, _, _, ok := Compact(id)
		require.True(t, ok, id)
		require.Len(t, rows, CompactH, id)
		for _, row := range rows {
			require.Len(t, row, CompactW, id)
		}
		require.GreaterOrEqual(t, CompactW, 8)
		require.LessOrEqual(t, CompactW, 11)
		boxes := eyeBoxes(rows)
		require.Len(t, boxes, 2, id)
		require.Equal(t, boxes[0].w, boxes[1].w, id)
		require.Equal(t, boxes[0].h, boxes[1].h, id)
		require.Equal(t, want, cropEye(rows, boxes[0]), id)
		require.Equal(t, want, cropEye(rows, boxes[1]), id)
		require.Less(t, boxes[1].y, boxes[0].y, id+" right eye sits higher")
	}
}

func TestPilloSmallFramesMatchTheLargeTimings(t *testing.T) {
	t.Parallel()
	clip := PilloClipSmall()
	require.Equal(t, CompactW, clip.W)
	require.Equal(t, CompactH, clip.H)
	require.Len(t, clip.Frames, 6)
	want := []string{".E", "EE", "E."}
	bottom := ""
	for i, rows := range clip.Frames {
		require.Len(t, rows, CompactH, i)
		for _, row := range rows {
			require.Len(t, row, CompactW, i)
		}
		require.Contains(t, rows[len(rows)-1], "B", "frame %d", i)
		if i == 0 {
			bottom = rows[len(rows)-1]
		}
		boxes := eyeBoxes(rows)
		require.Len(t, boxes, 2, i)
		require.Equal(t, want, cropEye(rows, boxes[0]), i)
		require.Equal(t, want, cropEye(rows, boxes[1]), i)
	}
	// Squish keeps the rest pose's feet.
	require.Equal(t, bottom, clip.Frames[4][len(clip.Frames[4])-1])

	p := NewPlayer(clip)
	p.Play("select")
	got := make([]int, 0, 6)
	for range 6 {
		got = append(got, frameIndex(p))
		p.Advance()
	}
	require.Equal(t, []int{4, 4, 5, 5, 0, 0}, got)
	require.True(t, p.Done())
}

func TestWriteCompactMedia(t *testing.T) {
	if os.Getenv("UPDATE_GOLDEN") != "1" {
		t.Skip()
	}
	dir := filepath.Join("..", "..", "docs", "media", "bots")
	require.NoError(t, writePNG(filepath.Join(dir, "small-preview.png"), RenderCompactPreview()))
	require.NoError(t, writePNG(filepath.Join(dir, "chat-preview.png"), RenderChatPreview()))
	require.NoError(t, WriteCompactPilloGIF(filepath.Join(dir, "pillo-small.gif")))
}

func TestSquintKeepsBothEyes(t *testing.T) {
	t.Parallel()
	rows, _, _, ok := Compact("dingus")
	require.True(t, ok)
	squint := Squint(rows)
	boxes := eyeBoxes(squint)
	require.Len(t, boxes, 2)
	require.Contains(t, strings.Join(squint, ""), "E")
}
