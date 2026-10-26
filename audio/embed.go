package audio

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

//go:embed sounds/*.wav
var soundsFS embed.FS

var (
	tempDirOnce sync.Once
	tempDir     string
	tempErr     error

	extractMu sync.Mutex
	extracted = make(map[string]bool)
)

// GetSoundPath returns the absolute path to a bundled sound file, extracting
// it from the embedded bundle on first request. Only the requested file is
// written to disk, so the first sound played costs one small file copy rather
// than the whole bundle.
func GetSoundPath(filename string) (string, error) {
	if filename != filepath.Base(filename) || filename == "." || filename == ".." {
		return "", fmt.Errorf("invalid sound file name %q", filename)
	}

	tempDirOnce.Do(func() {
		tempDir, tempErr = os.MkdirTemp("", "bvr-audio-*")
		if tempErr != nil {
			slog.Error("Failed to create temp dir for audio", "error", tempErr)
		}
	})
	if tempErr != nil {
		return "", tempErr
	}

	destPath := filepath.Join(tempDir, filename)

	extractMu.Lock()
	defer extractMu.Unlock()

	if extracted[filename] {
		return destPath, nil
	}

	data, err := soundsFS.ReadFile("sounds/" + filename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("sound file %s not found in bundle", filename)
		}
		return "", err
	}

	if err := os.WriteFile(destPath, data, 0o644); err != nil {
		slog.Error("Failed to write sound file to temp dir", "file", filename, "error", err)
		return "", err
	}
	extracted[filename] = true

	return destPath, nil
}

// Cleanup removes the temporary audio files. This should be called on exit.
func Cleanup() {
	if tempDir != "" {
		os.RemoveAll(tempDir)
	}
}
