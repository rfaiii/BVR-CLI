package node

import (
	"os"
	"path/filepath"
	"strings"
)

// The helpers below are the only places that turn a project path into a
// filesystem call. Each one repeats the containment check in this
// function so a path that reached the call is the same value
// strings.HasPrefix just accepted. confine() does the same check, plus
// filepath.IsLocal and EvalSymlinks, before it returns a path.

func rootPrefix(root string) string {
	return root + string(filepath.Separator)
}

func lstatContained(root, full string) (os.FileInfo, error) {
	if strings.HasPrefix(full, rootPrefix(root)) {
		return os.Lstat(full)
	}
	if full == root && strings.HasPrefix(root, string(filepath.Separator)) {
		return os.Lstat(root)
	}
	if full == root && strings.HasPrefix(root, filepath.VolumeName(root)+string(filepath.Separator)) {
		return os.Lstat(root)
	}
	return nil, errPathEscape
}

func readFileContained(root, full string) ([]byte, error) {
	if strings.HasPrefix(full, rootPrefix(root)) {
		return os.ReadFile(full)
	}
	return nil, errPathEscape
}

func readDirContained(root, full string) ([]os.DirEntry, error) {
	if strings.HasPrefix(full, rootPrefix(root)) {
		return os.ReadDir(full)
	}
	if full == root && strings.HasPrefix(root, string(filepath.Separator)) {
		return os.ReadDir(root)
	}
	if full == root && strings.HasPrefix(root, filepath.VolumeName(root)+string(filepath.Separator)) {
		return os.ReadDir(root)
	}
	return nil, errPathEscape
}

func openContained(root, full string) (*os.File, error) {
	if strings.HasPrefix(full, rootPrefix(root)) {
		return os.Open(full)
	}
	if full == root && strings.HasPrefix(root, string(filepath.Separator)) {
		return os.Open(root)
	}
	return nil, errPathEscape
}

func removeContained(root, full string) error {
	if strings.HasPrefix(full, rootPrefix(root)) {
		return os.Remove(full)
	}
	return errPathEscape
}

func renameContained(root, src, dst string) error {
	if strings.HasPrefix(src, rootPrefix(root)) && strings.HasPrefix(dst, rootPrefix(root)) {
		return os.Rename(src, dst)
	}
	return errPathEscape
}

func readlinkContained(root, full string) (string, error) {
	if strings.HasPrefix(full, rootPrefix(root)) {
		return os.Readlink(full)
	}
	return "", errPathEscape
}

// evalContained resolves full and accepts it only when the resolved
// path is still inside root. Callers use this on the deepest existing
// parent before creating the missing suffix.
func evalContained(root, full string) (string, error) {
	if strings.HasPrefix(full, rootPrefix(root)) {
		resolved, err := filepath.EvalSymlinks(full)
		if err != nil {
			return "", err
		}
		resolved = filepath.Clean(resolved)
		if strings.HasPrefix(resolved, rootPrefix(root)) {
			return resolved, nil
		}
		return "", errPathEscape
	}
	if full == root && strings.HasPrefix(full, string(filepath.Separator)) {
		resolved, err := filepath.EvalSymlinks(full)
		if err != nil {
			return "", err
		}
		resolved = filepath.Clean(resolved)
		if strings.HasPrefix(resolved, string(filepath.Separator)) {
			return resolved, nil
		}
	}
	if full == root && strings.HasPrefix(full, filepath.VolumeName(full)+string(filepath.Separator)) {
		resolved, err := filepath.EvalSymlinks(full)
		if err != nil {
			return "", err
		}
		resolved = filepath.Clean(resolved)
		if strings.HasPrefix(resolved, filepath.VolumeName(resolved)+string(filepath.Separator)) {
			return resolved, nil
		}
	}
	return "", errPathEscape
}
