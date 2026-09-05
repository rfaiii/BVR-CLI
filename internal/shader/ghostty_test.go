package shader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallGhosttyWritesConfig(t *testing.T) {
	base := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", base)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	if err := InstallGhostty(nil, true); err != nil {
		t.Fatalf("InstallGhostty: %v", err)
	}

	cfg := filepath.Join(base, ".config", "ghostty", "config")
	data, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "custom-shader-animation = always") {
		t.Fatalf("expected animation setting, got:\n%s", body)
	}
	for _, name := range Defaults() {
		if !strings.Contains(body, "custom-shader = "+filepath.Join(TargetDir(), name)) {
			t.Fatalf("expected shader %q in config, got:\n%s", name, body)
		}
	}
}

func TestInstallGhosttyHonorsForceAndPreservesWhenDisabled(t *testing.T) {
	base := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", base)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	cfg := filepath.Join(base, ".config", "ghostty", "config")
	if err := os.MkdirAll(filepath.Dir(cfg), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(cfg, []byte("custom-shader = old.glsl\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if err := InstallGhostty(nil, false); err == nil {
		t.Fatalf("expected error when force=false and existing custom-shader config exists")
	}
	data, _ := os.ReadFile(cfg)
	if !strings.Contains(string(data), "old.glsl") {
		t.Fatalf("expected existing custom-shader to remain unchanged, got:\n%s", string(data))
	}

	if err := InstallGhostty(nil, true); err != nil {
		t.Fatalf("InstallGhostty with force: %v", err)
	}
	data, _ = os.ReadFile(cfg)
	if strings.Contains(string(data), "old.glsl") {
		t.Fatalf("expected old custom-shader to be removed, got:\n%s", string(data))
	}
}
