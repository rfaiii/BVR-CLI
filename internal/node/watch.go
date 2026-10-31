package node

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// FSEvent is a filesystem change pushed over the NODE WebSocket.
// Op is create, write, remove, rename, or ready. ready is the
// acknowledgement that a watch subscription is active.
type FSEvent struct {
	ProjectID string `json:"project_id"`
	Path      string `json:"path"`
	Op        string `json:"op"`
	ETag      string `json:"etag,omitempty"`
}

// fsnotify on macOS uses kqueue, not the FSEvents C API. kqueue is
// CGO-free, which is what the Apple Silicon release build needs, and
// it requires a watch on every directory. startWatch adds those and
// adds directories created later. .git and node_modules are skipped
// so a checkout does not exhaust the process file-descriptor limit.
func (p *sharedProject) startWatch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	p.watcher = watcher
	added := 0
	walkErr := filepath.WalkDir(p.Root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return filepath.SkipDir
		}
		name := entry.Name()
		if path != p.Root && watchSkipped(name) {
			return filepath.SkipDir
		}
		if added >= maxWatchDirs {
			return fmt.Errorf("stopped after %d directories", maxWatchDirs)
		}
		if err := watcher.Add(path); err != nil {
			return err
		}
		added++
		return nil
	})
	go p.loop()
	return walkErr
}

func (p *sharedProject) close() {
	p.once.Do(func() {
		close(p.done)
		if p.watcher != nil {
			_ = p.watcher.Close()
		}
	})
}

func (p *sharedProject) loop() {
	if p.watcher == nil {
		return
	}
	for {
		select {
		case <-p.done:
			return
		case ev, ok := <-p.watcher.Events:
			if !ok {
				return
			}
			p.onEvent(ev)
		case _, ok := <-p.watcher.Errors:
			if !ok {
				return
			}
		}
	}
}

func (p *sharedProject) onEvent(ev fsnotify.Event) {
	if ev.Name == "" || p.share == nil {
		return
	}
	op := mapFSOp(ev.Op)
	if op == "" {
		return
	}
	if !within(p.Root, ev.Name) {
		return
	}
	rel := slashRel(p.Root, ev.Name)
	if rel == "." || watchPathSkipped(rel) || scratchName(filepath.Base(ev.Name)) {
		return
	}
	info, statErr := lstatContained(p.Root, ev.Name)
	if statErr == nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0 && op == "create" && p.watcher != nil {
		if strings.HasPrefix(ev.Name, p.Root+string(filepath.Separator)) {
			_ = p.watcher.Add(ev.Name)
		}
	}
	var etag string
	if statErr == nil && (info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0) {
		etag, _ = lstatETag(p.Root, ev.Name, info)
	}
	p.share.publish(FSEvent{ProjectID: p.ID, Path: rel, Op: op, ETag: etag})
}

func mapFSOp(op fsnotify.Op) string {
	switch {
	case op&fsnotify.Remove != 0:
		return "remove"
	case op&fsnotify.Rename != 0:
		return "rename"
	case op&fsnotify.Create != 0:
		return "create"
	case op&fsnotify.Write != 0:
		return "write"
	default:
		return ""
	}
}

func watchPathSkipped(rel string) bool {
	for _, part := range strings.Split(filepath.ToSlash(rel), "/") {
		if watchSkipped(part) {
			return true
		}
	}
	return false
}

type watchFilter struct {
	id      string
	project string
	path    string
}

type watchSub struct {
	mu      sync.Mutex
	filters []watchFilter
	send    func(WSFrame) error
}

func (s *FileShare) subscribe(send func(WSFrame) error) *watchSub {
	sub := &watchSub{send: send}
	s.mu.Lock()
	s.subs[sub] = struct{}{}
	s.mu.Unlock()
	return sub
}

func (s *FileShare) unsubscribe(sub *watchSub) {
	if s == nil || sub == nil {
		return
	}
	s.mu.Lock()
	delete(s.subs, sub)
	s.mu.Unlock()
}

func (s *FileShare) addWatch(sub *watchSub, id, projectID, rel string) error {
	if s == nil {
		return errNoProject
	}
	project, err := s.project(projectID)
	if err != nil {
		return err
	}
	project.mu.RLock()
	_, err = confine(project.Root, rel)
	project.mu.RUnlock()
	if err != nil {
		return err
	}
	path := "."
	if strings.TrimSpace(rel) != "" && rel != "." {
		path = filepath.ToSlash(filepath.Clean(rel))
	}
	sub.mu.Lock()
	sub.filters = append(sub.filters, watchFilter{id: id, project: projectID, path: path})
	sub.mu.Unlock()
	return sub.send(WSFrame{
		ID:   id,
		Type: "fs_event",
		Event: &FSEvent{
			ProjectID: projectID,
			Path:      path,
			Op:        "ready",
		},
	})
}

func (s *FileShare) publish(ev FSEvent) {
	s.mu.RLock()
	subs := make([]*watchSub, 0, len(s.subs))
	for sub := range s.subs {
		subs = append(subs, sub)
	}
	s.mu.RUnlock()
	for _, sub := range subs {
		sub.deliver(ev)
	}
}

func (sub *watchSub) deliver(ev FSEvent) {
	if sub == nil || sub.send == nil {
		return
	}
	sub.mu.Lock()
	filters := append([]watchFilter(nil), sub.filters...)
	sub.mu.Unlock()
	for _, filter := range filters {
		if filter.project != ev.ProjectID || !watchMatch(filter.path, ev.Path) {
			continue
		}
		_ = sub.send(WSFrame{ID: filter.id, Type: "fs_event", Event: &ev})
	}
}

func watchMatch(filter, eventPath string) bool {
	if filter == "" || filter == "." {
		return true
	}
	eventPath = filepath.ToSlash(eventPath)
	return eventPath == filter || strings.HasPrefix(eventPath, filter+"/")
}
