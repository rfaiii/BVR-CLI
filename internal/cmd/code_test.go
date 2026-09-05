package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNativeCodeTool_UsesConfiguredPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "codetool")
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
