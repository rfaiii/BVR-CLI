package imagelookup

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeImage_ResizesAndReturnsJPEG(t *testing.T) {
	t.Parallel()
	srcPath := writeJPEGTestImage(t, 4000, 3000)

	tmpPath, cleanup, err := NormalizeImage(srcPath, 1600)
	require.NoError(t, err)
	defer cleanup()
	require.FileExists(t, tmpPath)

	// Verify the normalized file is a valid JPEG and within size limits.
	f, err := os.Open(tmpPath)
	require.NoError(t, err)
	defer f.Close()

	img, _, err := image.Decode(f)
	require.NoError(t, err)
	bounds := img.Bounds()
	// Longest edge should be <= 1600.
	require.LessOrEqual(t, bounds.Dx(), 1600)
	require.LessOrEqual(t, bounds.Dy(), 1600)
	// Aspect ratio should be preserved (4000:3000 = 4:3).
	require.InDelta(t, float64(bounds.Dx())/float64(bounds.Dy()), 4.0/3.0, 0.05)
}

func TestNormalizeImage_PreservesSmallImage(t *testing.T) {
	t.Parallel()
	srcPath := writeJPEGTestImage(t, 800, 600)
	tmpPath, cleanup, err := NormalizeImage(srcPath, 1600)
	require.NoError(t, err)
	defer cleanup()

	f, err := os.Open(tmpPath)
	require.NoError(t, err)
	defer f.Close()
	img, _, err := image.Decode(f)
	require.NoError(t, err)
	bounds := img.Bounds()
	// Image is already under 1600, but aspect ratio preserved.
	require.Equal(t, 800, bounds.Dx())
	require.Equal(t, 600, bounds.Dy())
}

func TestNormalizeImage_TooSmall(t *testing.T) {
	t.Parallel()
	srcPath := writeJPEGTestImage(t, 50, 50)
	_, _, err := NormalizeImage(srcPath, 1600)
	require.Error(t, err)
	require.Contains(t, err.Error(), "too small")
}

func TestNormalizeImage_MissingFile(t *testing.T) {
	t.Parallel()
	_, _, err := NormalizeImage("/nonexistent.png", 1600)
	require.Error(t, err)
}

// writeJPEGTestImage creates a JPEG file of the given dimensions and returns
// its path.
func writeJPEGTestImage(t *testing.T, width, height int) string {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.Set(x, y, color.RGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: 128,
				A: 255,
			})
		}
	}
	path := filepath.Join(t.TempDir(), "test.jpg")
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()
	require.NoError(t, jpeg.Encode(f, img, &jpeg.Options{Quality: 85}))
	return path
}

// writePNGTestImage creates a PNG file of the given dimensions and returns
// its path.
func writePNGTestImage(t *testing.T, width, height int) string {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.Set(x, y, color.RGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: 128,
				A: 255,
			})
		}
	}
	path := filepath.Join(t.TempDir(), "test.png")
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()
	require.NoError(t, png.Encode(f, img))
	return path
}
