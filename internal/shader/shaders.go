package shader

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed *.glsl
var embeddedFS embed.FS

func Read(name string) ([]byte, error) {
	if name == "" {
		return nil, errors.New("shader name is empty")
	}
	if !strings.HasSuffix(name, ".glsl") {
		return nil, fmt.Errorf("shader %q must have .glsl extension", name)
	}
	data, err := fs.ReadFile(embeddedFS, name)
	if err != nil {
		return nil, fmt.Errorf("read shader %s: %w", name, err)
	}
	return data, nil
}

func List() ([]string, error) {
	entries, err := fs.Glob(embeddedFS, "*.glsl")
	if err != nil {
		return nil, fmt.Errorf("list shaders: %w", err)
	}
	return entries, nil
}

func Validate(name string) error {
	_, err := Read(name)
	return err
}

func Defaults() []string {
	return []string{
		"cursor_warp.glsl",
		"cursor_sweep.glsl",
		"cursor_tail.glsl",
		"rectangle_boom_cursor.glsl",
		"ripple_cursor.glsl",
		"ripple_rectangle_cursor.glsl",
		"sonic_boom_cursor.glsl",
	}
}

func TargetDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = "."
	}
	return filepath.Join(home, ".config", "ghostty", "shaders")
}
