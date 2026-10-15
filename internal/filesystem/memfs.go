package filesystem

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// MemFS is an in-memory filesystem used by tests to simulate I/O failures
// without touching the real operating system. It implements FS.
type MemFS struct {
	mu       sync.Mutex
	dirs     map[string]bool
	files    map[string][]byte
	symlinks map[string]string
	// WriteErr, if set, is returned by Commit's write steps.
	WriteErr error
	// RestoreErr, if set, is returned by Commit's restore step.
	RestoreErr error
}

// NewMemFS returns an empty in-memory filesystem.
func NewMemFS() *MemFS {
	return &MemFS{
		dirs:     map[string]bool{"": true},
		files:    map[string][]byte{},
		symlinks: map[string]string{},
	}
}

func (m *MemFS) clean(p string) string {
	p = filepath.ToSlash(filepath.Clean(p))
	p = strings.ReplaceAll(p, "\\", "/")
	p = strings.TrimPrefix(p, "/")
	return p
}

func (m *MemFS) mkdirAllLocked(p string) {
	p = m.clean(p)
	for {
		m.dirs[p] = true
		i := strings.LastIndex(p, "/")
		if i < 0 {
			return
		}
		p = p[:i]
	}
}

var errNotFound = errors.New("not found")

func (m *MemFS) stat(p string) (fileInfo, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p = m.clean(p)
	if d, ok := m.dirs[p]; ok {
		return fileInfo{isDir: d}, false, nil
	}
	if c, ok := m.files[p]; ok {
		return fileInfo{content: c}, false, nil
	}
	if _, ok := m.symlinks[p]; ok {
		return fileInfo{}, true, nil
	}
	return fileInfo{}, false, errNotFound
}

type fileInfo struct {
	content []byte
	isDir   bool
}

func (i fileInfo) Name() string       { return "" }
func (i fileInfo) IsDir() bool        { return i.isDir }
func (i fileInfo) Size() int64        { return int64(len(i.content)) }
func (i fileInfo) ModTime() time.Time { return time.Time{} }
func (i fileInfo) Sys() interface{}   { return nil }

// PlanBuild computes the change plan.
func (m *MemFS) PlanBuild(artifacts map[string][]byte) (Plan, error) {
	p := Plan{Backup: map[string]string{}, Files: map[string][]byte{}}
	for rel, content := range artifacts {
		if strings.HasPrefix(rel, "/") || strings.HasPrefix(rel, "\\") {
			return Plan{}, fmt.Errorf("filesystem: path %q escapes target root", rel)
		}
		clean := m.clean(rel)
		if clean == "" || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") {
			return Plan{}, fmt.Errorf("filesystem: path %q escapes target root", rel)
		}
		p.Files[clean] = content
		info, isSymlink, err := m.stat(clean)
		if err == nil {
			if isSymlink {
				return Plan{}, fmt.Errorf("filesystem: refusing to overwrite symlink %q", clean)
			}
			if info.IsDir() {
				return Plan{}, fmt.Errorf("filesystem: refusing to overwrite directory %q", clean)
			}
			p.Replace = append(p.Replace, clean)
			p.Conflicts = append(p.Conflicts, clean)
			p.Backup[clean] = clean + ".devopsgen.bak"
		} else if errors.Is(err, errNotFound) {
			p.Create = append(p.Create, clean)
		} else {
			return Plan{}, fmt.Errorf("filesystem: stat %q: %w", clean, err)
		}
	}
	return p, nil
}

// DryRun reports conflicts without writing anything.
func (m *MemFS) DryRun(p Plan) (Plan, error) {
	return p, nil
}

// Commit applies the plan with best-effort rollback.
// The whole operation runs under a single lock; rollback helpers are
// inline to avoid re-locking.
func (m *MemFS) Commit(p Plan) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	created := make([]string, 0, len(p.Create))

	// Backup existing files before replacing.
	for orig, backup := range p.Backup {
		m.files[backup] = m.files[orig]
		delete(m.files, orig)
	}

	// Write new files.
	for _, dst := range p.Create {
		if m.WriteErr != nil {
			m.rollbackCreateLocked(created)
			m.rollbackReplaceLocked(p.Backup)
			return fmt.Errorf("filesystem: write %q: %w", dst, m.WriteErr)
		}
		m.mkdirAllLocked(filepath.Dir(dst))
		m.files[dst] = p.Files[dst]
		created = append(created, dst)
	}

	// Write replaced files.
	for _, dst := range p.Replace {
		if m.WriteErr != nil {
			m.rollbackCreateLocked(created)
			m.rollbackReplaceLocked(p.Backup)
			return fmt.Errorf("filesystem: write %q: %w", dst, m.WriteErr)
		}
		m.files[dst] = p.Files[dst]
	}

	// Success: discard backups. If restore itself is configured to fail,
	// treat it as a rollback failure (best-effort restore failed).
	for _, backup := range p.Backup {
		delete(m.files, backup)
	}
	if m.RestoreErr != nil {
		m.rollbackReplaceLocked(p.Backup)
		return fmt.Errorf("filesystem: restore failed: %w", m.RestoreErr)
	}
	return nil
}

func (m *MemFS) rollbackCreateLocked(paths []string) error {
	var first error
	for i := len(paths) - 1; i >= 0; i-- {
		delete(m.files, paths[i])
	}
	return first
}

func (m *MemFS) rollbackReplaceLocked(backup map[string]string) error {
	var first error
	for orig, b := range backup {
		delete(m.files, orig)
		if c, ok := m.files[b]; ok {
			m.files[orig] = c
			delete(m.files, b)
		} else if first == nil {
			first = errors.New("filesystem: backup missing during rollback")
		}
	}
	return first
}

// Get returns the content of a stored file.
func (m *MemFS) Get(p string) ([]byte, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.files[m.clean(p)]
	return c, ok
}

// Put writes content directly (helper for test setup).
func (m *MemFS) Put(p string, content []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mkdirAllLocked(filepath.Dir(p))
	m.files[m.clean(p)] = content
}

// Mkdir creates a directory entry (helper for test setup).
func (m *MemFS) Mkdir(p string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dirs[m.clean(p)] = true
}

// Symlink creates a symlink target.
func (m *MemFS) Symlink(p string, target string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.symlinks[m.clean(p)] = target
}
