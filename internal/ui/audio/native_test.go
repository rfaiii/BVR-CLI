package audio

import "testing"

func TestQuickNotifyMappings(t *testing.T) {
	tests := map[string]string{
		"quick-notify-01": "QUICK-NOTIFY-01-FLAT.wav",
		"quick-notify-02": "QUICK-NOTIFY-02-FLAT.wav",
		"quick-notify-03": "QUICK-NOTIFY-03-FLAT.wav",
		"notification":    "QUICK-NOTIFY-01-FLAT.wav",
	}
	for event, want := range tests {
		if got := getAudioFilename(event); got != want {
			t.Errorf("getAudioFilename(%q) = %q, want %q", event, got, want)
		}
	}
}

func TestVolumeDefaultsAndCompatibility(t *testing.T) {
	tests := map[string]string{
		"":        "25",
		"high":    "100",
		"low":     "50",
		"25":      "25",
		"50":      "50",
		"75":      "75",
		"100":     "100",
		"silent":  "silent",
		"unknown": "25",
	}
	for input, want := range tests {
		if got := NormalizeVolume(input); got != want {
			t.Errorf("NormalizeVolume(%q) = %q, want %q", input, got, want)
		}
	}
	if got := VolumePercent(""); got != 25 {
		t.Fatalf("VolumePercent(\"\") = %d, want 25", got)
	}
}
