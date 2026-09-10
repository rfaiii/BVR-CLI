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
	defaultAudioFunc = func(title, message, audioType, volume string) error {
		percent := VolumePercent(volume)
		if percent == 0 {
			return nil
		}

		filename := getAudioFilename(audioType)
		path, err := audio.GetSoundPath(filename)
		if err != nil {
			slog.Error("Failed to get audio path for native playback", "error", err)
			return err
		}

		// Use WPF MediaPlayer so Windows can honor the same 25/50/75/100 scale
		// as the macOS and Linux native backends.
		script := fmt.Sprintf(`Add-Type -AssemblyName PresentationCore; $p=New-Object System.Windows.Media.MediaPlayer; $p.Open([Uri]::new('%s')); $p.Volume=%.2f; $p.Play(); Start-Sleep -Milliseconds 2000; $p.Close()`, filepath.Clean(path), float64(percent)/100)
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
