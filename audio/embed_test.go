package audio

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetSoundPathExtractsRequestedFileOnly(t *testing.T) {
	name := "STARTUP-SONG-FLAT.wav"
	if _, err := os.Stat(filepath.Join("sounds", name)); err != nil {
		// The embedded FS is compiled in; check the source tree instead.
		if _, srcErr := os.Stat("sounds/" + name); srcErr != nil {
			t.Skipf("sound %s not present in bundle", name)
		}
	}

	path, err := GetSoundPath(name)
	if err != nil {
		t.Fatalf("GetSoundPath(%q) = %v", name, err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("extracted file missing: %v", err)
	}

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("read temp dir: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected only %q extracted, found %d files", name, len(entries))
	}

	second, err := GetSoundPath(name)
	if err != nil || second != path {
		t.Errorf("repeat call = (%q, %v), want (%q, nil)", second, err, path)
	}
}

func TestGetSoundPathRejectsTraversal(t *testing.T) {
	for _, name := range []string{"../../etc/passwd", "sounds/CHAT-01-FLAT.wav", "/tmp/evil.wav"} {
		if got, err := GetSoundPath(name); err == nil {
			t.Errorf("GetSoundPath(%q) = %q, want error", name, got)
		}
	}
}

func TestGetSoundPathUnknownFile(t *testing.T) {
	if _, err := GetSoundPath("DOES-NOT-EXIST.wav"); err == nil {
		t.Error("expected error for unknown sound file")
	}
}
