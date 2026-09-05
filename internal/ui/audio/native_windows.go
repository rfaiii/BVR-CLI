//go:build windows

package audio

import (
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"

	"github.com/richavery/bvr-cli/audio"
)

func init() {
	defaultAudioFunc = func(title, message, audioType string) error {
		filename := getAudioFilename(audioType)
		path, err := audio.GetSoundPath(filename)
		if err != nil {
			slog.Error("Failed to get audio path for native playback", "error", err)
			return err
		}

		// Windows uses PowerShell to play audio files natively via .NET Media.SoundPlayer
		// Use -WindowStyle Hidden to prevent flashes
		script := fmt.Sprintf(`(New-Object System.Media.SoundPlayer '%s').PlaySync()`, filepath.Clean(path))
		cmd := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", script)
		
		if err := cmd.Start(); err != nil {
			slog.Error("Failed to start powershell audio playback", "error", err)
			return err
		}

		// We do not wait for the command to finish so we don't block the UI
		go func() {
			_ = cmd.Wait()
		}()

		return nil
	}
}
