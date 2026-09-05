package shader

import (
	"strings"
	"testing"
)

func TestDefaults(t *testing.T) {
	defs := Defaults()
	if len(defs) == 0 {
		t.Fatal("expected non-empty default shader list")
	}
	for _, name := range defs {
		if !strings.HasSuffix(name, ".glsl") {
			t.Errorf("default shader %q missing .glsl suffix", name)
		}
	}
}

func TestReadEmbedded(t *testing.T) {
	for _, name := range Defaults() {
		data, err := Read(name)
		if err != nil {
			t.Fatalf("Read(%q) failed: %v", name, err)
		}
		if len(data) == 0 {
			t.Errorf("Read(%q) returned empty data", name)
		}
		if !strings.Contains(string(data), "mainImage") {
			t.Errorf("Read(%q) missing expected shader mainImage entry", name)
		}
	}
}

func TestListReturnsMatches(t *testing.T) {
	list, err := List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != len(Defaults()) {
		t.Fatalf("List() = %d, want %d", len(list), len(Defaults()))
	}
}

func TestValidateRejectsBadNames(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"missing extension", "foo"},
		{"path traversal", "../cursor_warp.glsl"},
	}
	for _, tc := range cases {
		if err := Validate(tc.in); err == nil {
			t.Errorf("Validate(%q) returned nil, want error", tc.in)
		}
	}
}
