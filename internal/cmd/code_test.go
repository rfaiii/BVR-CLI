package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNativeCodeTool_UsesConfiguredPath(t *testing.T) {
	// exec.LookPath on Windows only accepts a name whose extension is in
	// PATHEXT, so the fixture has to carry one there. The mode bits are
	// not what decides it — a 0o755 "codetool" is simply not an
	// executable on that platform, and LookPath is correct to say so.
	name := "codetool"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BVR_CODETOOL", path)

	got, err := nativeCodeTool()
	if err != nil {
		t.Fatal(err)
	}
	if got != path {
		t.Fatalf("nativeCodeTool() = %q, want %q", got, path)
	}
}

func TestNativeCodeTool_RejectsMissingConfiguredPath(t *testing.T) {
	t.Setenv("BVR_CODETOOL", filepath.Join(t.TempDir(), "missing"))

	if _, err := nativeCodeTool(); err == nil {
		t.Fatal("nativeCodeTool() returned nil error for a missing configured executable")
	}
}
