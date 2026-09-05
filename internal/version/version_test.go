package version

import "testing"

func TestShortVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"devel", "devel"},
		{"v1.2.3", "v1.2.3"},
		{"v0.87.1-0.20260731174531-4d...", "v0.87.1"},
		{"v1.0.0-alpha.1", "v1.0.0-alpha.1"},
		{"", ""},
	}

	origVersion := Version
	defer func() { Version = origVersion }()

	for _, tt := range tests {
		Version = tt.input
		result := ShortVersion()
		if result != tt.expected {
			t.Errorf("ShortVersion() = %q, want %q", result, tt.expected)
		}
	}
}
