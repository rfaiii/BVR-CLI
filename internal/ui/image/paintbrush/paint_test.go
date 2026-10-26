package paintbrush

import (
	"image"
	"image/color"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// gradient returns a test image whose pixels change on every row and column,
// so no two cells pick the same glyph and a dropped result shows up as a gap.
func gradient(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{
				R: uint8(x % 256), //nolint:gosec // test pattern
				G: uint8(y % 256), //nolint:gosec // test pattern
				B: 0x80,
				A: 0xff,
			})
		}
	}
	return img
}

// Paint used to wait for its workers but not for the goroutine that copied
// their results into the task slice, so processResults read entries that were
// still being written. The race is inside a single Paint call, which is why
// this repeats it from several canvases at once: that is how -race got enough
// scheduler pressure to see it.
func TestPaintCollectsEveryResult(t *testing.T) {
	t.Parallel()

	const (
		cols = 24
		rows = 16
	)

	results := make([]string, 4)
	var wg sync.WaitGroup
	for i := range results {
		wg.Add(1)
		go func() {
			defer wg.Done()
			canvas := New()
			canvas.SetImage(gradient(cols*8, rows*8))
			canvas.SetWidth(cols)
			canvas.SetHeight(rows)
			canvas.Paint()
			results[i] = canvas.GetResult()
		}()
	}
	wg.Wait()

	for i, got := range results {
		require.NotEmpty(t, got)
		// Every cell is written before the result is indexed, so the same image
		// at the same size always paints the same string.
		require.Equal(t, results[0], got, "canvas %d", i)
	}
}
