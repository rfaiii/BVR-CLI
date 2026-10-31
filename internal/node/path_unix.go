//go:build unix

package node

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// mkdirAllContained creates full inside root. Each new component is
// created with mkdirat on a directory opened O_NOFOLLOW, so a symlink
// swapped in after the lexical check is not followed.
func mkdirAllContained(root, full string) error {
	if full == root {
		return nil
	}
	if !strings.HasPrefix(full, rootPrefix(root)) {
		return errPathEscape
	}
	rel, err := filepath.Rel(root, full)
	if err != nil || !filepath.IsLocal(rel) {
		return errPathEscape
	}
	dirfd, err := openProjectRoot(root)
	if err != nil {
		return err
	}
	defer unix.Close(dirfd)
	fd := dirfd
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		if part == "" || part == "." {
			continue
		}
		if !filepath.IsLocal(part) {
			if fd != dirfd {
				unix.Close(fd)
			}
			return errPathEscape
		}
		next, err := unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		if err != nil {
			if mkErr := unix.Mkdirat(fd, part, 0o755); mkErr != nil && !errors.Is(mkErr, unix.EEXIST) {
				if fd != dirfd {
					unix.Close(fd)
				}
				return mkErr
			}
			next, err = unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
			if err != nil {
				if fd != dirfd {
					unix.Close(fd)
				}
				if errors.Is(err, unix.ELOOP) {
					return errSymlink
				}
				return err
			}
		}
		if fd != dirfd {
			unix.Close(fd)
		}
		fd = next
	}
	if fd != dirfd {
		unix.Close(fd)
	}
	return nil
}

// writeAtomicContained replaces full without following a symlink at the
// final component. The parent is resolved with EvalSymlinks and rejected
// when that resolution leaves root. The bytes are written to a scratch
// name via openat(O_NOFOLLOW) and renamed into place on that directory fd.
func writeAtomicContained(root, full string, content []byte) error {
	if !strings.HasPrefix(full, rootPrefix(root)) {
		return errPathEscape
	}
	parent := filepath.Dir(full)
	base := filepath.Base(full)
	if !filepath.IsLocal(base) {
		return errPathEscape
	}
	resolved, err := filepath.EvalSymlinks(parent)
	if err != nil {
		return err
	}
	resolved = filepath.Clean(resolved)
	if !(strings.HasPrefix(resolved, rootPrefix(root)) || resolved == root) {
		return errPathEscape
	}
	if resolved != parent {
		return errSymlink
	}
	dirfd, err := openDirInRoot(root, resolved)
	if err != nil {
		return err
	}
	defer unix.Close(dirfd)
	scratch := ".bvrnode-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	if !filepath.IsLocal(scratch) {
		return errPathEscape
	}
	fd, err := unix.Openat(dirfd, scratch, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0o644)
	if err != nil {
		return err
	}
	file := os.NewFile(uintptr(fd), scratch)
	_, werr := file.Write(content)
	cerr := file.Close()
	if werr != nil || cerr != nil {
		_ = unix.Unlinkat(dirfd, scratch, 0)
		if werr != nil {
			return werr
		}
		return cerr
	}
	if err := unix.Renameat(dirfd, scratch, dirfd, base); err != nil {
		_ = unix.Unlinkat(dirfd, scratch, 0)
		return err
	}
	return nil
}

// openProjectRoot opens the shared root. The HasPrefix check is what
// keeps the directory argument tied to a path that cannot climb out.
func openProjectRoot(root string) (int, error) {
	if strings.HasPrefix(root, string(filepath.Separator)) {
		return unix.Open(root, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	}
	if strings.HasPrefix(root, filepath.VolumeName(root)+string(filepath.Separator)) {
		return unix.Open(root, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	}
	return 0, errPathEscape
}

func openDirInRoot(root, dir string) (int, error) {
	if strings.HasPrefix(dir, rootPrefix(root)) {
		return unix.Open(dir, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	}
	if dir == root {
		return openProjectRoot(root)
	}
	return 0, errPathEscape
}
