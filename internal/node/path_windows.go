//go:build windows

package node

import (
	"os"
	"path/filepath"
	"strings"
)

func mkdirAllContained(root, full string) error {
	if full == root {
		return nil
	}
	if strings.HasPrefix(full, rootPrefix(root)) {
		return os.MkdirAll(full, 0o755)
	}
	return errPathEscape
}

func writeAtomicContained(root, full string, content []byte) error {
	if !strings.HasPrefix(full, rootPrefix(root)) {
		return errPathEscape
	}
	parent := filepath.Dir(full)
	if !strings.HasPrefix(parent, rootPrefix(root)) && parent != root {
		return errPathEscape
	}
	resolved, err := filepath.EvalSymlinks(parent)
	if err != nil {
		return err
	}
	resolved = filepath.Clean(resolved)
	if strings.HasPrefix(resolved, rootPrefix(root)) || resolved == root {
		if resolved != parent {
			return errSymlink
		}
		return os.WriteFile(full, content, 0o644)
	}
	return errPathEscape
}
