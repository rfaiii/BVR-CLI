//go:build linux

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

		paplayArgs := []string{path}
		paplayArgs = append(paplayArgs, fmt.Sprintf("--volume=%d", 65536*percent/100))

		// Try paplay (PulseAudio) first, fallback to aplay (ALSA)
		cmd := exec.Command("paplay", paplayArgs...)
		if err := cmd.Start(); err != nil {
			cmd = exec.Command("aplay", "-q", path)
			if err := cmd.Start(); err != nil {
				slog.Error("Failed to start audio playback on linux", "error", err)
				return err
			}
		}

		// We do not wait for the command to finish so we don't block the UI
		go func() {
			_ = cmd.Wait()
		}()

		return nil
	}
}
