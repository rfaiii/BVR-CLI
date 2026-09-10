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

func TestBeaverAudioPool(t *testing.T) {
	want := []string{
		"WHAT-FLAT.wav",
		"OH-BEAV-01.wav",
		"OH-BEAV-02.wav",
		"OH-BEAV-03.wav",
		"OH-BEAV-04.wav",
		"OH-BEAV-05.wav",
	}
	for choice, expected := range want {
		if got := beaverAudioFilename(choice); got != expected {
			t.Errorf("beaverAudioFilename(%d) = %q, want %q", choice, got, expected)
		}
	}
	allowed := make(map[string]bool, len(want))
	for _, filename := range want {
		allowed[filename] = true
	}
	for i := 0; i < 100; i++ {
		if filename := getAudioFilename("beaver"); !allowed[filename] {
			t.Fatalf("getAudioFilename(\"beaver\") returned unexpected file %q", filename)
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
