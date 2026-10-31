package node

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	// MaxFileBytes is the largest file a peer may read or write in one
	// request.
	MaxFileBytes   = 32 << 20
	maxListEntries = 10000
	maxListDepth   = 8
	maxWatchDirs   = 8192
)

var (
	errPathEscape   = errors.New("path escapes the shared project")
	errSymlink      = errors.New("symlinks are not followed")
	errNoProject    = errors.New("no such shared project")
	errETagRequired = errors.New("expected_etag is required")
	errNotEmpty     = errors.New("directory is not empty")
	errIsDir        = errors.New("path is a directory")
)

// ConflictError is returned when a write would clobber another device.
// ETag is the current hash, or empty when the file is gone. Msg, when
// set, replaces that sentence (directory already exists, and so on).
type ConflictError struct {
	ETag string
	Msg  string
}

func (e *ConflictError) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	if e.ETag == "" {
		return "conflict: file is missing"
	}
	return "conflict: current etag " + e.ETag
}

// ProjectInfo is one directory MAIN opted into with --project.
type ProjectInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Root string `json:"root"`
}

// FileEntry is one path inside a shared project. Paths use forward
// slashes and are relative to the project root.
type FileEntry struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Type    string    `json:"type"`
	Size    int64     `json:"size,omitempty"`
	ETag    string    `json:"etag,omitempty"`
	ModTime time.Time `json:"mod_time,omitempty"`
}

// FileList is a directory listing.
type FileList struct {
	Project string      `json:"project"`
	Path    string      `json:"path"`
	Entries []FileEntry `json:"entries"`
}

// FileContent is a file body plus the etag a later write must send.
type FileContent struct {
	Path    string    `json:"path"`
	ETag    string    `json:"etag"`
	Size    int64     `json:"size"`
	Content string    `json:"content"`
	ModTime time.Time `json:"mod_time,omitempty"`
}

// FileShare is the set of project roots MAIN is willing to serve.
// Writes use an optimistic etag: the client sends the hash it read,
// and a per-project mutex makes the check and the write atomic. An
// empty expected etag means "create, and fail if the path exists."
type FileShare struct {
	mu       sync.RWMutex
	projects map[string]*sharedProject
	subs     map[*watchSub]struct{}
	warning  string
}

type sharedProject struct {
	mu      sync.RWMutex
	once    sync.Once
	ID      string
	Name    string
	Root    string
	watcher *fsnotify.Watcher
	done    chan struct{}
	share   *FileShare
}

// NewFileShare opens each root. Relative paths must already be
// absolute. A symlink used as the root is resolved, and that real
// directory is the boundary peers cannot leave.
func NewFileShare(roots []string) (*FileShare, error) {
	share := &FileShare{
		projects: map[string]*sharedProject{},
		subs:     map[*watchSub]struct{}{},
	}
	used := map[string]struct{}{}
	var warnings []string
	for _, root := range roots {
		canonical, err := canonicalRoot(root)
		if err != nil {
			share.Close()
			return nil, err
		}
		id := uniqueProjectID(filepath.Base(canonical), canonical, used)
		project := &sharedProject{
			ID:    id,
			Name:  filepath.Base(canonical),
			Root:  canonical,
			done:  make(chan struct{}),
			share: share,
		}
		if err := project.startWatch(); err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %s", id, err.Error()))
		}
		share.projects[id] = project
	}
	if len(warnings) > 0 {
		share.warning = "Watch partial: " + strings.Join(warnings, "; ")
	}
	return share, nil
}

// WatchWarning is set when fsnotify could not watch every directory.
func (s *FileShare) WatchWarning() string {
	if s == nil {
		return ""
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.warning
}

// Projects lists the opted-in roots.
func (s *FileShare) Projects() []ProjectInfo {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ProjectInfo, 0, len(s.projects))
	for _, project := range s.projects {
		out = append(out, ProjectInfo{ID: project.ID, Name: project.Name, Root: project.Root})
	}
	sortProjects(out)
	return out
}

// Close stops filesystem watches.
func (s *FileShare) Close() {
	if s == nil {
		return
	}
	s.mu.Lock()
	projects := make([]*sharedProject, 0, len(s.projects))
	for _, project := range s.projects {
		projects = append(projects, project)
	}
	s.mu.Unlock()
	for _, project := range projects {
		project.close()
	}
}

func (s *FileShare) project(id string) (*sharedProject, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	project, ok := s.projects[id]
	if !ok {
		return nil, errNoProject
	}
	return project, nil
}

// List returns the directory at rel, or the file itself when rel is a
// file. depth 0 returns just that path. depth 1 is the children.
func (s *FileShare) List(projectID, rel string, depth int) (FileList, error) {
	project, err := s.project(projectID)
	if err != nil {
		return FileList{}, err
	}
	return project.list(rel, depth)
}

// ReadFile returns the bytes and etag of a regular file.
func (s *FileShare) ReadFile(projectID, rel string) (FileContent, error) {
	project, err := s.project(projectID)
	if err != nil {
		return FileContent{}, err
	}
	return project.read(rel)
}

// WriteFile creates or replaces a file. expected is required. An empty
// string creates the file only when it does not already exist.
func (s *FileShare) WriteFile(projectID, rel string, content []byte, expected *string) (FileEntry, error) {
	project, err := s.project(projectID)
	if err != nil {
		return FileEntry{}, err
	}
	return project.write(rel, content, expected)
}

// Mkdir creates a directory. Parents inside the project are created.
func (s *FileShare) Mkdir(projectID, rel string) (FileEntry, error) {
	project, err := s.project(projectID)
	if err != nil {
		return FileEntry{}, err
	}
	return project.mkdir(rel)
}

// DeleteFile removes a file or an empty directory. Files and symlinks
// require expected to match the current etag. Directories ignore it.
func (s *FileShare) DeleteFile(projectID, rel string, expected *string) error {
	project, err := s.project(projectID)
	if err != nil {
		return err
	}
	return project.remove(rel, expected)
}

// RenameFile moves a path inside the same project. Files require an
// etag. The destination must not exist.
func (s *FileShare) RenameFile(projectID, from, to string, expected *string) (FileEntry, error) {
	project, err := s.project(projectID)
	if err != nil {
		return FileEntry{}, err
	}
	return project.rename(from, to, expected)
}

func (p *sharedProject) list(rel string, depth int) (FileList, error) {
	if depth < 0 {
		depth = 0
	}
	if depth > maxListDepth {
		depth = maxListDepth
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	full, err := confine(p.Root, rel)
	if err != nil {
		return FileList{}, err
	}
	info, err := os.Lstat(full)
	if err != nil {
		return FileList{}, err
	}
	list := FileList{Project: p.ID, Path: slashRel(p.Root, full)}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		entry, err := entryFor(p.Root, full, info)
		if err != nil {
			return FileList{}, err
		}
		list.Entries = []FileEntry{entry}
		return list, nil
	}
	if depth == 0 {
		entry, err := entryFor(p.Root, full, info)
		if err != nil {
			return FileList{}, err
		}
		list.Entries = []FileEntry{entry}
		return list, nil
	}
	entries, err := walkFiles(p.Root, full, depth)
	if err != nil {
		return FileList{}, err
	}
	list.Entries = entries
	return list, nil
}

func (p *sharedProject) read(rel string) (FileContent, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	full, err := confine(p.Root, rel)
	if err != nil {
		return FileContent{}, err
	}
	info, err := os.Lstat(full)
	if err != nil {
		return FileContent{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return FileContent{}, errSymlink
	}
	if info.IsDir() {
		return FileContent{}, errIsDir
	}
	if !info.Mode().IsRegular() {
		return FileContent{}, fmt.Errorf("not a regular file")
	}
	if info.Size() > MaxFileBytes {
		return FileContent{}, fmt.Errorf("file exceeds %d bytes", MaxFileBytes)
	}
	data, err := os.ReadFile(full)
	if err != nil {
		return FileContent{}, err
	}
	return FileContent{
		Path:    slashRel(p.Root, full),
		ETag:    etagBytes(data),
		Size:    int64(len(data)),
		Content: string(data),
		ModTime: info.ModTime().UTC(),
	}, nil
}

func (p *sharedProject) write(rel string, content []byte, expected *string) (FileEntry, error) {
	if expected == nil {
		return FileEntry{}, errETagRequired
	}
	if len(content) > MaxFileBytes {
		return FileEntry{}, fmt.Errorf("file exceeds %d bytes", MaxFileBytes)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	full, err := confine(p.Root, rel)
	if err != nil {
		return FileEntry{}, err
	}
	if full == p.Root {
		return FileEntry{}, errIsDir
	}
	info, statErr := os.Lstat(full)
	switch {
	case statErr != nil && !os.IsNotExist(statErr):
		return FileEntry{}, statErr
	case os.IsNotExist(statErr):
		if *expected != "" {
			return FileEntry{}, &ConflictError{}
		}
	default:
		if info.Mode()&os.ModeSymlink != 0 {
			return FileEntry{}, errSymlink
		}
		if info.IsDir() {
			return FileEntry{}, errIsDir
		}
		current, err := fileETag(full)
		if err != nil {
			return FileEntry{}, err
		}
		if current != *expected {
			return FileEntry{}, &ConflictError{ETag: current}
		}
	}
	parent := filepath.Dir(full)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return FileEntry{}, err
	}
	if !within(p.Root, parent) {
		return FileEntry{}, errPathEscape
	}
	tmp, err := os.CreateTemp(parent, ".bvrnode-*")
	if err != nil {
		return FileEntry{}, err
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		return FileEntry{}, err
	}
	if err := tmp.Close(); err != nil {
		return FileEntry{}, err
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return FileEntry{}, err
	}
	if err := os.Rename(tmpName, full); err != nil {
		return FileEntry{}, err
	}
	cleanup = false
	info, err = os.Lstat(full)
	if err != nil {
		return FileEntry{}, err
	}
	entry, err := entryFor(p.Root, full, info)
	if err != nil {
		return FileEntry{}, err
	}
	entry.ETag = etagBytes(content)
	return entry, nil
}

func (p *sharedProject) mkdir(rel string) (FileEntry, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	full, err := confine(p.Root, rel)
	if err != nil {
		return FileEntry{}, err
	}
	if full == p.Root {
		return FileEntry{}, fmt.Errorf("project root already exists")
	}
	if info, err := os.Lstat(full); err == nil {
		if info.IsDir() {
			return FileEntry{}, &ConflictError{Msg: "directory already exists"}
		}
		return FileEntry{}, fmt.Errorf("path exists and is not a directory")
	} else if !os.IsNotExist(err) {
		return FileEntry{}, err
	}
	if err := os.MkdirAll(full, 0o755); err != nil {
		return FileEntry{}, err
	}
	if !within(p.Root, full) {
		return FileEntry{}, errPathEscape
	}
	if p.watcher != nil {
		_ = p.watcher.Add(full)
	}
	info, err := os.Lstat(full)
	if err != nil {
		return FileEntry{}, err
	}
	return entryFor(p.Root, full, info)
}

func (p *sharedProject) remove(rel string, expected *string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	full, err := confine(p.Root, rel)
	if err != nil {
		return err
	}
	if full == p.Root {
		return fmt.Errorf("refusing to delete the project root")
	}
	info, err := os.Lstat(full)
	if err != nil {
		return err
	}
	if info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
		empty, err := dirEmpty(full)
		if err != nil {
			return err
		}
		if !empty {
			return errNotEmpty
		}
		return os.Remove(full)
	}
	if expected == nil {
		return errETagRequired
	}
	current, err := lstatETag(full, info)
	if err != nil {
		return err
	}
	if current != *expected {
		return &ConflictError{ETag: current}
	}
	return os.Remove(full)
}

func (p *sharedProject) rename(from, to string, expected *string) (FileEntry, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	src, err := confine(p.Root, from)
	if err != nil {
		return FileEntry{}, err
	}
	dst, err := confine(p.Root, to)
	if err != nil {
		return FileEntry{}, err
	}
	if src == p.Root || dst == p.Root {
		return FileEntry{}, fmt.Errorf("refusing to rename the project root")
	}
	info, err := os.Lstat(src)
	if err != nil {
		return FileEntry{}, err
	}
	if _, err := os.Lstat(dst); err == nil {
		return FileEntry{}, &ConflictError{Msg: "destination already exists"}
	} else if !os.IsNotExist(err) {
		return FileEntry{}, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		if expected == nil {
			return FileEntry{}, errETagRequired
		}
		current, err := lstatETag(src, info)
		if err != nil {
			return FileEntry{}, err
		}
		if current != *expected {
			return FileEntry{}, &ConflictError{ETag: current}
		}
	}
	parent := filepath.Dir(dst)
	parentInfo, err := os.Lstat(parent)
	if err != nil {
		return FileEntry{}, err
	}
	if !parentInfo.IsDir() || parentInfo.Mode()&os.ModeSymlink != 0 {
		return FileEntry{}, fmt.Errorf("destination parent is not a directory")
	}
	if err := os.Rename(src, dst); err != nil {
		return FileEntry{}, err
	}
	info, err = os.Lstat(dst)
	if err != nil {
		return FileEntry{}, err
	}
	return entryFor(p.Root, dst, info)
}

func canonicalRoot(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("project %s: %w", abs, err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", resolved)
	}
	return resolved, nil
}

func uniqueProjectID(base, root string, used map[string]struct{}) string {
	id := safeID(base)
	if _, ok := used[id]; ok {
		sum := sha256.Sum256([]byte(root))
		id = id + "-" + hex.EncodeToString(sum[:4])
	}
	used[id] = struct{}{}
	return id
}

func safeID(base string) string {
	var b strings.Builder
	for _, r := range base {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	id := strings.Trim(b.String(), "-.")
	if id == "" {
		return "project"
	}
	return id
}

// confine resolves rel inside root. Absolute paths, ".." segments, and
// any symlink along the way are rejected, including a symlink whose
// target happens to land back inside the root.
func confine(root, request string) (string, error) {
	if strings.ContainsRune(request, 0) {
		return "", errPathEscape
	}
	request = strings.TrimPrefix(request, "./")
	if request == "" || request == "." {
		return root, nil
	}
	if filepath.IsAbs(request) {
		return "", errPathEscape
	}
	cleaned := filepath.Clean(request)
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", errPathEscape
	}
	full := filepath.Join(root, cleaned)
	if !within(root, full) {
		return "", errPathEscape
	}
	current := root
	for _, part := range strings.Split(cleaned, string(filepath.Separator)) {
		if part == "" || part == "." {
			continue
		}
		if part == ".." {
			return "", errPathEscape
		}
		next := filepath.Join(current, part)
		info, err := os.Lstat(next)
		if err != nil {
			if os.IsNotExist(err) {
				current = next
				continue
			}
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", errSymlink
		}
		current = next
	}
	if !within(root, current) {
		return "", errPathEscape
	}
	return current, nil
}

func within(root, path string) bool {
	root = filepath.Clean(root)
	path = filepath.Clean(path)
	if path == root {
		return true
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

func slashRel(root, full string) string {
	if filepath.Clean(full) == filepath.Clean(root) {
		return "."
	}
	rel, err := filepath.Rel(root, full)
	if err != nil {
		return filepath.ToSlash(full)
	}
	return filepath.ToSlash(rel)
}

func entryFor(root, full string, info os.FileInfo) (FileEntry, error) {
	entry := FileEntry{
		Name:    filepath.Base(full),
		Path:    slashRel(root, full),
		Size:    info.Size(),
		ModTime: info.ModTime().UTC(),
	}
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		entry.Type = "symlink"
		entry.Size = 0
		etag, err := lstatETag(full, info)
		if err != nil {
			return FileEntry{}, err
		}
		entry.ETag = etag
	case info.IsDir():
		entry.Type = "dir"
		entry.Size = 0
	case info.Mode().IsRegular():
		entry.Type = "file"
		etag, err := fileETag(full)
		if err != nil {
			return FileEntry{}, err
		}
		entry.ETag = etag
	default:
		entry.Type = "other"
	}
	return entry, nil
}

func walkFiles(root, dir string, depth int) ([]FileEntry, error) {
	var entries []FileEntry
	var walk func(string, int) error
	walk = func(current string, remaining int) error {
		if len(entries) > maxListEntries {
			return fmt.Errorf("listing exceeds %d entries", maxListEntries)
		}
		children, err := os.ReadDir(current)
		if err != nil {
			return err
		}
		for _, child := range children {
			if len(entries) > maxListEntries {
				return fmt.Errorf("listing exceeds %d entries", maxListEntries)
			}
			full := filepath.Join(current, child.Name())
			info, err := os.Lstat(full)
			if err != nil {
				return err
			}
			if scratchName(child.Name()) {
				continue
			}
			if watchSkipped(child.Name()) && info.IsDir() {
				continue
			}
			entry, err := entryFor(root, full, info)
			if err != nil {
				return err
			}
			entries = append(entries, entry)
			if info.IsDir() && info.Mode()&os.ModeSymlink == 0 && remaining > 1 {
				if err := walk(full, remaining-1); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := walk(dir, depth); err != nil {
		return nil, err
	}
	return entries, nil
}

func fileETag(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, io.LimitReader(file, MaxFileBytes+1)); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func etagBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func lstatETag(path string, info os.FileInfo) (string, error) {
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return "", err
		}
		return etagBytes([]byte("symlink:" + target)), nil
	}
	if info.IsDir() {
		return "", nil
	}
	return fileETag(path)
}

func dirEmpty(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()
	_, err = file.Readdirnames(1)
	if errors.Is(err, io.EOF) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

func watchSkipped(name string) bool {
	switch name {
	case ".git", "node_modules":
		return true
	default:
		return false
	}
}

// scratchName marks the temporary file WriteFile uses while replacing
// a path. It must not show up in listings or watch events.
func scratchName(name string) bool {
	return strings.HasPrefix(name, ".bvrnode-")
}

func sortProjects(list []ProjectInfo) {
	for i := 1; i < len(list); i++ {
		for j := i; j > 0 && list[j].ID < list[j-1].ID; j-- {
			list[j], list[j-1] = list[j-1], list[j]
		}
	}
}
