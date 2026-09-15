package imagelookup

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	_ "golang.org/x/image/webp"
)

// ValidateImage checks that the file at path exists, is within the size limit,
// uses a supported image format, and has dimensions of at least 128x128
// pixels. It does not modify the original file.
func ValidateImage(path string, maxBytes int64) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("read image metadata: %w", err)
	}

	if info.Size() > maxBytes {
		return fmt.Errorf("image is too large: maximum is %d MB", maxBytes/(1024*1024))
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		return fmt.Errorf("unsupported image format; use JPG, PNG, or WebP")
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open image: %w", err)
	}
	defer file.Close()

	cfg, _, err := image.DecodeConfig(file)
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}

	if cfg.Width < 128 || cfg.Height < 128 {
		return fmt.Errorf("image is too small; use at least 128x128 pixels")
	}

	return nil
}
