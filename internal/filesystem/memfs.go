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
	// BackupErr, if set, is returned when creating backups.
	BackupErr error
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

// statLocked assumes the mutex is already held.
func (m *MemFS) statLocked(p string) (fileInfo, bool, error) {
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

func (m *MemFS) stat(p string) (fileInfo, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.statLocked(p)
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
	m.mu.Lock()
	defer m.mu.Unlock()

	p := Plan{Backup: map[string]string{}, Files: map[string][]byte{}}

	paths := make([]string, 0, len(artifacts))
	for rel := range artifacts {
		paths = append(paths, rel)
	}
	sortStrings(paths)

	for _, rel := range paths {
		content := artifacts[rel]
		if strings.HasPrefix(rel, "/") || strings.HasPrefix(rel, "\\") {
			return Plan{}, fmt.Errorf("filesystem: path %q escapes target root", rel)
		}
		clean := m.clean(rel)
		if clean == "" || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") {
			return Plan{}, fmt.Errorf("filesystem: path %q escapes target root", rel)
		}
		p.Files[clean] = content

		// Reject symlink targets.
		if _, isSymlink := m.symlinks[clean]; isSymlink {
			return Plan{}, fmt.Errorf("filesystem: refusing to overwrite symlink %q", clean)
		}
		info, _, err := m.statLocked(clean)
		if err == nil {
			if info.IsDir() {
				return Plan{}, fmt.Errorf("filesystem: refusing to overwrite directory %q", clean)
			}
			p.Replace = append(p.Replace, clean)
			p.Conflicts = append(p.Conflicts, clean)
			backup := clean + ".devopsgen.bak"
			if _, ok := m.files[backup]; ok {
				return Plan{}, fmt.Errorf("filesystem: backup path %q already exists", backup)
			}
			p.Backup[clean] = backup
		} else if errors.Is(err, errNotFound) {
			p.Create = append(p.Create, clean)
			dir := filepath.Dir(clean)
			if !dirExists(p.Dirs, dir) {
				p.Dirs = append(p.Dirs, dir)
			}
		} else {
			return Plan{}, fmt.Errorf("filesystem: stat %q: %w", clean, err)
		}
	}
	sortStrings(p.Dirs)
	return p, nil
}

// DryRun reports conflicts without writing anything.
func (m *MemFS) DryRun(p Plan) (Plan, error) {
	return p, nil
}

// Commit applies the plan with best-effort rollback.
func (m *MemFS) Commit(p Plan) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	created := make([]string, 0, len(p.Create))
	var rollbackErrors []string

	// Backup existing files before replacing.
	// Sort backup origins for deterministic rollback order.
	backupOrigins := make([]string, 0, len(p.Backup))
	for orig := range p.Backup {
		backupOrigins = append(backupOrigins, orig)
	}
	sortStrings(backupOrigins)

	for _, orig := range backupOrigins {
		backup := p.Backup[orig]
		if m.BackupErr != nil {
			rollbackErrors = append(rollbackErrors, fmt.Sprintf("backup %q: %v", orig, m.BackupErr))
			// Roll back any backups already made (reverse order).
			for i := len(backupOrigins) - 1; i >= 0; i-- {
				o := backupOrigins[i]
				if o == orig {
					break
				}
				b := p.Backup[o]
				if c, ok := m.files[b]; ok {
					m.files[o] = c
					delete(m.files, b)
				}
			}
			m.rollbackCreateLocked(created)
			return fmt.Errorf("filesystem: backup %q: %v; rollback errors: %v", orig, m.BackupErr, strings.Join(rollbackErrors, "; "))
		}
		if c, ok := m.files[orig]; ok {
			m.files[backup] = c
		}
		delete(m.files, orig)
	}

	// Create directories.
	for _, dir := range p.Dirs {
		m.mkdirAllLocked(dir)
	}

	// Write new files.
	for _, dst := range p.Create {
		if m.WriteErr != nil {
			rollbackErrors = append(rollbackErrors, fmt.Sprintf("write %q: %v", dst, m.WriteErr))
			m.rollbackReplaceLocked(p.Backup)
			m.rollbackCreateLocked(created)
			return fmt.Errorf("filesystem: write %q: %v; rollback errors: %v", dst, m.WriteErr, strings.Join(rollbackErrors, "; "))
		}
		m.mkdirAllLocked(filepath.Dir(dst))
		m.files[dst] = p.Files[dst]
		created = append(created, dst)
	}

	// Write replaced files.
	for _, dst := range p.Replace {
		if m.WriteErr != nil {
			rollbackErrors = append(rollbackErrors, fmt.Sprintf("write %q: %v", dst, m.WriteErr))
			m.rollbackReplaceLocked(p.Backup)
			m.rollbackCreateLocked(created)
			return fmt.Errorf("filesystem: write %q: %v; rollback errors: %v", dst, m.WriteErr, strings.Join(rollbackErrors, "; "))
		}
		m.files[dst] = p.Files[dst]
	}

	// Success: discard backups.
	for _, backup := range p.Backup {
		delete(m.files, backup)
	}
	if m.RestoreErr != nil {
		m.rollbackReplaceLocked(p.Backup)
		return fmt.Errorf("filesystem: restore failed: %w; rollback errors: %v", m.RestoreErr, strings.Join(rollbackErrors, "; "))
	}
	return nil
}

func (m *MemFS) rollbackCreateLocked(paths []string) error {
	for i := len(paths) - 1; i >= 0; i-- {
		delete(m.files, paths[i])
	}
	return nil
}

func (m *MemFS) rollbackReplaceLocked(backup map[string]string) error {
	var first error
	// Sort backup keys for deterministic restore order.
	origs := make([]string, 0, len(backup))
	for orig := range backup {
		origs = append(origs, orig)
	}
	sortStrings(origs)

	for i := len(origs) - 1; i >= 0; i-- {
		orig := origs[i]
		b := backup[orig]
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
