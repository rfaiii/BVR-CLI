package imagelookup

import (
	"fmt"
	"image"
	"os"

	"github.com/disintegration/imaging"
)

// NormalizeImage opens srcPath, decodes it (JPEG, PNG, or WebP), resizes the
// longest edge to maxDimension while preserving aspect ratio, and writes a
// JPEG copy to a temporary file. The original file is left untouched. The
// caller must invoke the returned cleanup function to remove the temp file.
func NormalizeImage(srcPath string, maxDimension int) (string, func(), error) {
	if maxDimension <= 0 {
		return "", nil, fmt.Errorf("max dimension must be positive")
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", nil, fmt.Errorf("open source image: %w", err)
	}
	defer srcFile.Close()

	img, _, err := image.Decode(srcFile)
	if err != nil {
		return "", nil, fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w < 128 || h < 128 {
		return "", nil, fmt.Errorf("image is too small; use at least 128x128 pixels")
	}

	// Compute target dimensions preserving aspect ratio so the longest edge
	// equals maxDimension (or stays smaller if already below the limit).
	var newW, newH int
	if w > h {
		newW = min(maxDimension, w)
		newH = h * newW / w
	} else {
		newH = min(maxDimension, h)
		newW = w * newH / h
	}

	// Resize using a high-quality Lanczos filter.
	resized := imaging.Resize(img, newW, newH, imaging.Lanczos)

	tmpFile, err := os.CreateTemp("", "bvr-lookup-*.jpg")
	if err != nil {
		return "", nil, fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	if err := imaging.Save(resized, tmpPath, imaging.JPEGQuality(85)); err != nil {
		os.Remove(tmpPath)
		return "", nil, fmt.Errorf("save normalized image: %w", err)
	}

	cleanup := func() {
		_ = os.Remove(tmpPath)
	}

	return tmpPath, cleanup, nil
}
