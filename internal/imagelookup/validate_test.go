package imagelookup

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateImage_Valid(t *testing.T) {
	t.Parallel()
	path := writeTestImage(t, 200, 200)
	err := ValidateImage(path, 12*1024*1024)
	require.NoError(t, err)
}

func TestValidateImage_TooLarge(t *testing.T) {
	t.Parallel()
	path := writeTestImage(t, 200, 200)
	// 1 byte limit should fail for a non-empty file.
	err := ValidateImage(path, 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "too large")
}

func TestValidateImage_UnsupportedFormat(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.bmp")
	f, err := os.Create(path)
	require.NoError(t, err)
	_, _ = f.Write([]byte{0x42, 0x4D})
	f.Close()
	err = ValidateImage(path, 12*1024*1024)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported image format")
}

func TestValidateImage_TooSmall(t *testing.T) {
	t.Parallel()
	path := writeTestImage(t, 50, 50)
	err := ValidateImage(path, 12*1024*1024)
	require.Error(t, err)
	require.Contains(t, err.Error(), "too small")
}

func TestValidateImage_MissingFile(t *testing.T) {
	t.Parallel()
	err := ValidateImage("/nonexistent/image.png", 12*1024*1024)
	require.Error(t, err)
}

// writeTestImage creates a small PNG file and returns its path.
func writeTestImage(t *testing.T, width, height int) string {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.Set(x, y, color.RGBA{R: 128, G: 64, B: 200, A: 255})
		}
	}
	path := filepath.Join(t.TempDir(), "test.png")
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()
	require.NoError(t, png.Encode(f, img))
	return path
}
