//go:build darwin

package audio

import (
	"fmt"
	"log/slog"
	"os/exec"

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

		volFlag := fmt.Sprintf("%.2f", float64(percent)/100)

		// afplay is native to macOS
		cmd := exec.Command("afplay", "-v", volFlag, path)
		if err := cmd.Start(); err != nil {
			slog.Error("Failed to start afplay", "error", err)
			return err
		}

		// We do not wait for the command to finish so we don't block the UI
		go func() {
			_ = cmd.Wait()
		}()

		return nil
	}
}
