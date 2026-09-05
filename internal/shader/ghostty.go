package shader

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func InstallGhostty(shaders []string, force bool) error {
	if len(shaders) == 0 {
		shaders = Defaults()
	}

	cfgDir := filepath.Join(os.Getenv("HOME"), ".config", "ghostty")
	cfgPath := filepath.Join(cfgDir, "config")
	shaderDir := TargetDir()

	if err := os.MkdirAll(shaderDir, 0o755); err != nil {
		return fmt.Errorf("create shader dir: %w", err)
	}

	missing := []string{}
	for _, name := range shaders {
		data, err := Read(name)
		if err != nil {
			missing = append(missing, name)
			continue
		}
		target := filepath.Join(shaderDir, name)
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return fmt.Errorf("write shader %s: %w", name, err)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing shaders: %v", missing)
	}

	if _, err := os.Stat(cfgPath); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(cfgDir, 0o755); err != nil {
			return fmt.Errorf("create ghostty config dir: %w", err)
		}
		if err := os.WriteFile(cfgPath, []byte{}, 0o644); err != nil {
			return fmt.Errorf("create ghostty config: %w", err)
		}
	}

	lines, err := readLines(cfgPath)
	if err != nil {
		return err
	}

	shaderSet := make(map[string]struct{}, len(shaders))
	for _, s := range shaders {
		shaderSet[s] = struct{}{}
	}

	var newLines []string
	updated := false
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#") {
			newLines = append(newLines, line)
			continue
		}
		if strings.HasPrefix(t, "custom-shader = ") {
			if !force {
				return fmt.Errorf("existing custom-shader config found; rerun with --force to overwrite")
			}
			updated = true
			continue
		}
		newLines = append(newLines, line)
	}

	if !updated {
		newLines = append(newLines, "", "# Added by bvr-cli shader installer")
		newLines = append(newLines, "custom-shader-animation = always")
		var keys []string
		for k := range shaderSet {
			keys = append(keys, k)
		}
		slices.SortFunc(keys, strings.Compare)
		for _, name := range keys {
			newLines = append(newLines, fmt.Sprintf("custom-shader = %s", filepath.Join(shaderDir, name)))
		}
	}

	if err := writeLines(cfgPath, newLines); err != nil {
		return err
	}
	return nil
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		out = append(out, s.Text())
	}
	return out, s.Err()
}

func writeLines(path string, lines []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return w.Flush()
}
