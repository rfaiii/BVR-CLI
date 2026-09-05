package config_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/richavery/bvr-cli/internal/config"
	"github.com/stretchr/testify/require"
)

// isolateReloadEnv points HOME/XDG/BVR_* at a throwaway directory so a
// reload only sees the config files the test writes. No t.Parallel(): these
// tests Setenv.
func isolateReloadEnv(t *testing.T) (workDir, dataDir string) {
	t.Helper()
	isolated := t.TempDir()
	t.Setenv("HOME", isolated)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(isolated, ".config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(isolated, ".local", "share"))
	t.Setenv("BVR_GLOBAL_CONFIG", filepath.Join(isolated, ".config", "bvr"))
	t.Setenv("BVR_GLOBAL_DATA", filepath.Join(isolated, ".local", "share", "bvr"))
	return t.TempDir(), t.TempDir()
}

// TestReloadFromDisk_PicksUpEditedBvrrc verifies that a bvrrc edit on
// disk is reflected in memory after a reload, matching how JSON config
// already behaves.
func TestReloadFromDisk_PicksUpEditedBvrrc(t *testing.T) {
	workDir, dataDir := isolateReloadEnv(t)
	rcPath := filepath.Join(workDir, "bvrrc")
	require.NoError(t, os.WriteFile(rcPath, []byte("option notifications bell\n"), 0o644))

	store, err := config.Load(workDir, dataDir, false)
	require.NoError(t, err)
	require.Equal(t, "bell", store.Config().Options.Notifications)

	require.NoError(t, os.WriteFile(rcPath, []byte("option notifications osc\n"), 0o644))
	require.NoError(t, store.ReloadFromDisk(context.Background()))
	require.Equal(t, "osc", store.Config().Options.Notifications,
		"reload must re-execute the bvrrc and pick up the new value")
}

// TestReloadFromDisk_FailingBvrrcKeepsOldConfig verifies that a bvrrc
// which fails to load during a reload does not clobber the in-memory config.
// The reload must return an error and leave the previously-loaded config
// intact (the swap happens only after a successful parse).
func TestReloadFromDisk_FailingBvrrcKeepsOldConfig(t *testing.T) {
	workDir, dataDir := isolateReloadEnv(t)
	rcPath := filepath.Join(workDir, "bvrrc")
	require.NoError(t, os.WriteFile(rcPath, []byte("option notifications bell\n"), 0o644))

	store, err := config.Load(workDir, dataDir, false)
	require.NoError(t, err)
	require.Equal(t, "bell", store.Config().Options.Notifications)

	// Overwrite with a bvrrc that errors at load time (unknown option key).
	require.NoError(t, os.WriteFile(rcPath, []byte("option totally-bogus-key value\n"), 0o644))

	err = store.ReloadFromDisk(context.Background())
	require.Error(t, err, "a failing bvrrc must surface a reload error")
	require.Equal(t, "bell", store.Config().Options.Notifications,
		"in-memory config must be preserved when a reload fails")
}

// TestReloadFromDisk_HangingBvrrcIsInterruptible is the core regression test
// for the reload-path blocker: config reloads run while holding the store's
// write lock, so a bvrrc that blocks (busy loop, hung command substitution)
// must be interruptible via the reload context rather than wedging the store
// forever. A cancelled reload must return an error and leave the old config
// in place. The test bounds its own wait so a regression can't hang CI.
func TestReloadFromDisk_HangingBvrrcIsInterruptible(t *testing.T) {
	workDir, dataDir := isolateReloadEnv(t)
	rcPath := filepath.Join(workDir, "bvrrc")
	require.NoError(t, os.WriteFile(rcPath, []byte("option notifications bell\n"), 0o644))

	store, err := config.Load(workDir, dataDir, false)
	require.NoError(t, err)
	require.Equal(t, "bell", store.Config().Options.Notifications)

	// Overwrite with a bvrrc that never terminates.
	require.NoError(t, os.WriteFile(rcPath, []byte("while true; do :; done\n"), 0o644))

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	t.Cleanup(cancel)

	done := make(chan error, 1)
	go func() {
		done <- store.ReloadFromDisk(ctx)
	}()

	select {
	case err := <-done:
		require.Error(t, err, "a cancelled reload must fail, not succeed")
	case <-time.After(3 * time.Second):
		t.Fatal("ReloadFromDisk did not return after context cancellation; the store would be wedged")
	}

	require.Equal(t, "bell", store.Config().Options.Notifications,
		"in-memory config must be preserved when a reload is cancelled")
}

// TestLoad_TracksNotYetCreatedGlobalBvrrc verifies that config.Load tracks
// the global bvrrc path even when the file does not exist yet, so a bvrrc
// created after startup is detected as a staleness change. Previously only
// successfully-loaded paths were tracked, so a mid-session global bvrrc went
// unnoticed until something else triggered a reload.
//
// Project-level bvrrc files are discovered by walking the tree for existing
// files, so a not-yet-created project bvrrc is still only picked up on the
// next reload; the global path is the common case this covers.
func TestLoad_TracksNotYetCreatedGlobalBvrrc(t *testing.T) {
	workDir, dataDir := isolateReloadEnv(t)
	globalRC := filepath.Join(t.TempDir(), "bvrrc")
	t.Setenv("BVR_GLOBAL_CONFIG", filepath.Dir(globalRC))

	// A provider must be configured so Load runs past its early
	// "not configured" return and reaches the staleness snapshot capture.
	require.NoError(t, os.WriteFile(
		filepath.Join(workDir, "bvrrc"),
		[]byte("provider add openai --api-key k\n"), 0o644))

	// Load with no global bvrrc present.
	store, err := config.Load(workDir, dataDir, false)
	require.NoError(t, err)
	require.False(t, store.ConfigStaleness().Dirty, "fresh load should be clean")

	// Create the global bvrrc after startup.
	require.NoError(t, os.WriteFile(globalRC, []byte("option debug true\n"), 0o644))

	staleness := store.ConfigStaleness()
	require.True(t, staleness.Dirty, "creating a global bvrrc must be detected")
	require.Contains(t, staleness.Changed, globalRC)
}
