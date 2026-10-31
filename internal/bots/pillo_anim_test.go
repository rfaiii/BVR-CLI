package bots

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPilloFramesShareCanvasAndCleanEyes(t *testing.T) {
	t.Parallel()
	clip := PilloClip()
	require.Len(t, clip.Frames, 6)
	wide := []string{
		".EE",
		"EEE",
		"EEE",
		"EE.",
	}
	small := []string{
		".E",
		"EE",
		"E.",
	}
	for i, rows := range clip.Frames {
		require.Lenf(t, rows, pilloH, "frame %d height", i+1)
		for _, row := range rows {
			require.Lenf(t, row, pilloW, "frame %d width", i+1)
		}
		require.Contains(t, rows[len(rows)-1], "B", "frame %d must sit on the bottom row", i+1)
		boxes := eyeBoxes(rows)
		require.Lenf(t, boxes, 2, "frame %d", i+1)
		require.Equal(t, boxes[0].w, boxes[1].w)
		require.Equal(t, boxes[0].h, boxes[1].h)
		want := wide
		if i == 5 {
			want = small
		}
		require.Equal(t, want, cropEye(rows, boxes[0]), "frame %d left", i+1)
		require.Equal(t, want, cropEye(rows, boxes[1]), "frame %d right", i+1)
	}
}

func TestPilloAnimationsAndReduceMotion(t *testing.T) {
	p := NewPlayer(PilloClip())
	p.Play("select")
	// Squish, squish, hop, hop, rest, rest, then done.
	got := make([]int, 0, 6)
	for range 6 {
		got = append(got, frameIndex(p))
		p.Advance()
	}
	require.Equal(t, []int{4, 4, 5, 5, 0, 0}, got)
	require.True(t, p.Done())

	p.Play("idle")
	require.Equal(t, 0, frameIndex(p))
	for range 8 {
		p.Advance()
	}
	require.Equal(t, 2, frameIndex(p), "glance up-right after the rest hold")

	t.Setenv("BVR_REDUCE_MOTION", "1")
	p.Advance()
	require.Equal(t, "rest", p.Name())
	require.Equal(t, 0, frameIndex(p))
	require.Equal(t, p.clip.Frames[0], p.Rows())
}

func TestPilloFrameGolden(t *testing.T) {
	t.Parallel()
	var b strings.Builder
	for i, rows := range PilloClip().Frames {
		b.WriteString("frame ")
		b.WriteByte(byte('1' + i))
		b.WriteByte('\n')
		for _, row := range rows {
			b.WriteString(row)
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}
	path := filepath.Join("testdata", "pillo-frames.golden")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		require.NoError(t, os.WriteFile(path, []byte(b.String()), 0o644))
		require.NoError(t, WritePilloGIF(filepath.Join("..", "..", "docs", "media", "bots", "pillo-anim.gif")))
		require.NoError(t, WritePilloStrip(filepath.Join("..", "..", "docs", "media", "bots", "pillo-frames.png")))
	}
	want, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, string(want), b.String())
}

func frameIndex(p *Player) int {
	return p.clip.Seqs[p.name].Steps[p.step].Frame
}

func cropEye(rows []string, b eyeBox) []string {
	out := make([]string, b.h)
	for y := 0; y < b.h; y++ {
		line := rows[b.y+y][b.x : b.x+b.w]
		line = strings.ReplaceAll(line, "B", ".")
		out[y] = line
	}
	return out
}
