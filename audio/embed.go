package audio

import (
	"embed"
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
	tempDir     string
	extractOnce sync.Once
	extractErr  error
)

// GetSoundPath extracts the bundled sounds to a temporary directory
// and returns the absolute path to the requested sound file.
// If the sound does not exist, it returns an empty string.
func GetSoundPath(filename string) (string, error) {
	extractOnce.Do(func() {
		// Create a temporary directory for BVR audio
		tempDir, extractErr = os.MkdirTemp("", "bvr-audio-*")
		if extractErr != nil {
			slog.Error("Failed to create temp dir for audio", "error", extractErr)
			return
		}

		// Extract all wav files
		entries, err := fs.ReadDir(soundsFS, "sounds")
		if err != nil {
			extractErr = err
			return
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			data, err := soundsFS.ReadFile(filepath.Join("sounds", entry.Name()))
			if err != nil {
				slog.Error("Failed to read embedded sound file", "file", entry.Name(), "error", err)
				continue
			}

			destPath := filepath.Join(tempDir, entry.Name())
			if err := os.WriteFile(destPath, data, 0o644); err != nil {
				slog.Error("Failed to write sound file to temp dir", "file", entry.Name(), "error", err)
			}
		}
	})

	if extractErr != nil {
		return "", extractErr
	}

	targetPath := filepath.Join(tempDir, filename)
	if _, err := os.Stat(targetPath); err == nil {
		return targetPath, nil
	}

	return "", fmt.Errorf("sound file %s not found in bundle", filename)
}

// Cleanup removes the temporary audio files. This should be called on exit.
func Cleanup() {
	if tempDir != "" {
		os.RemoveAll(tempDir)
	}
}
