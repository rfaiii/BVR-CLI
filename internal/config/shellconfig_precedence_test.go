package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/richavery/bvr-cli/internal/config"
	"github.com/stretchr/testify/require"
)

// TestShellConfigDotBvrrcTakesPrecedence verifies that a project-local
// .bvrrc overrides bvrrc in the same directory on conflicting settings.
func TestShellConfigDotBvrrcTakesPrecedence(t *testing.T) {
	isolated := t.TempDir()
	t.Setenv("HOME", isolated)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(isolated, ".config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(isolated, ".local", "share"))

	workDir := t.TempDir()
	dataDir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(workDir, "bvrrc"),
		[]byte("option notifications bell\n"), 0o644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(workDir, ".bvrrc"),
		[]byte("option notifications osc\n"), 0o644,
	))

	store, err := config.Load(workDir, dataDir, false)
	require.NoError(t, err)
	require.Equal(t, "osc", store.Config().Options.Notifications,
		".bvrrc should win over bvrrc")
}
