// Package filesystem owns filesystem safety: path validation, change
// planning, conflict detection, dry-run, backup, commit, and best-effort
// rollback.
//
// Operations must be injectable so I/O failures can be simulated in tests.
package filesystem

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Plan describes the intended filesystem changes for a generation run.
type Plan struct {
	// Create holds paths that will be written for the first time.
	Create []string
	// Replace holds existing paths that will be overwritten.
	Replace []string
	// Conflicts holds paths that would be overwritten without consent.
	Conflicts []string
	// Backup maps replaced paths to their backup locations.
	Backup map[string]string
	// Files maps destination paths to their rendered content.
	Files map[string][]byte
	// Dirs holds directories created by the transaction.
	Dirs []string
}

// FS is the filesystem operation interface used by the engine.
type FS interface {
	// PlanBuild computes the change plan for the given artifacts.
	PlanBuild(artifacts map[string][]byte) (Plan, error)
	// Commit applies the plan. On failure it performs best-effort rollback.
	Commit(Plan) error
	// DryRun reports conflicts without writing anything.
	DryRun(Plan) (Plan, error)
}

// RealFS implements FS against the real operating system.
type RealFS struct {
	Root string
}

func (f *RealFS) PlanBuild(artifacts map[string][]byte) (Plan, error) {
	if err := f.validateRoot(); err != nil {
		return Plan{}, err
	}
	p := Plan{Backup: map[string]string{}, Files: map[string][]byte{}}

	// Sort artifact paths for deterministic planning (NFR-04).
	paths := make([]string, 0, len(artifacts))
	for rel := range artifacts {
		paths = append(paths, rel)
	}
	sortStrings(paths)

	for _, rel := range paths {
		content := artifacts[rel]
		clean, err := f.resolveArtifactPath(rel)
		if err != nil {
			return Plan{}, err
		}
		p.Files[clean] = content

		// Validate that no parent component of the target is a symlink,
		// which could redirect writes outside the target root.
		if err := f.checkSymlinkParents(clean); err != nil {
			return Plan{}, err
		}

		info, statErr := os.Lstat(clean)
		if statErr == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				return Plan{}, fmt.Errorf("filesystem: refusing to overwrite symlink %q", clean)
			}
			if info.IsDir() {
				return Plan{}, fmt.Errorf("filesystem: refusing to overwrite directory %q", clean)
			}
			p.Replace = append(p.Replace, clean)
			p.Conflicts = append(p.Conflicts, clean)
			backup := clean + ".devopsgen.bak"
			// Reject backup collision.
			if _, statErr2 := os.Lstat(backup); statErr2 == nil {
				return Plan{}, fmt.Errorf("filesystem: backup path %q already exists", backup)
			}
			p.Backup[clean] = backup
		} else if errors.Is(statErr, os.ErrNotExist) {
			p.Create = append(p.Create, clean)
			// Track parent directories that will need creation.
			dir := filepath.Dir(clean)
			if !dirExists(p.Dirs, dir) {
				p.Dirs = append(p.Dirs, dir)
			}
		} else {
			// lstat failed for a reason other than "not exists" (e.g. a parent
			// component is a regular file, so the target path cannot be
			// resolved on some platforms). Treat the target as a create
			// candidate; Commit will surface the real error (and roll back)
			// when it cannot create the parent directories.
			p.Create = append(p.Create, clean)
			dir := filepath.Dir(clean)
			if !dirExists(p.Dirs, dir) {
				p.Dirs = append(p.Dirs, dir)
			}
		}
	}
	// Sort directories deepest-first for creation, and reverse for cleanup.
	sortStrings(p.Dirs)
	return p, nil
}

// DryRun reports conflicts without writing anything.
func (f *RealFS) DryRun(p Plan) (Plan, error) {
	return p, nil
}

// Commit applies the plan with best-effort rollback on failure.
func (f *RealFS) Commit(p Plan) error {
	if err := f.validateRoot(); err != nil {
		return err
	}
	created := make([]string, 0, len(p.Create))
	createdDirs := make([]string, 0, len(p.Dirs))
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
		if err := os.Rename(orig, backup); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Sprintf("backup %q: %v", orig, err))
			// Roll back any backups already made (reverse order).
			for i := len(backupOrigins) - 1; i >= 0; i-- {
				o := backupOrigins[i]
				if o == orig {
					break
				}
				b := p.Backup[o]
				if _, statErr := os.Lstat(b); statErr == nil {
					if rerr := os.Rename(b, o); rerr != nil {
						rollbackErrors = append(rollbackErrors, fmt.Sprintf("rollback %q: %v", o, rerr))
					}
				}
			}
			if rerr := f.rollbackCreate(created); rerr != nil {
				rollbackErrors = append(rollbackErrors, fmt.Sprintf("rollback-created: %v", rerr))
			}
			if rerr := f.rollbackRemoveDirs(createdDirs); rerr != nil {
				rollbackErrors = append(rollbackErrors, fmt.Sprintf("rollback-dirs: %v", rerr))
			}
			return fmt.Errorf("filesystem: backup %q: %v; rollback errors: %v", orig, err, strings.Join(rollbackErrors, "; "))
		}
	}

	// Create directories first.
	for _, dir := range p.Dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Sprintf("mkdir %q: %v", dir, err))
			if rerr := f.rollbackReplace(p.Backup); rerr != nil {
				rollbackErrors = append(rollbackErrors, fmt.Sprintf("rollback-replace: %v", rerr))
			}
			if rerr := f.rollbackCreate(created); rerr != nil {
				rollbackErrors = append(rollbackErrors, fmt.Sprintf("rollback-created: %v", rerr))
			}
			if rerr := f.rollbackRemoveDirs(createdDirs); rerr != nil {
				rollbackErrors = append(rollbackErrors, fmt.Sprintf("rollback-dirs: %v", rerr))
			}
			return fmt.Errorf("filesystem: mkdir %q: %v; rollback errors: %v", dir, err, strings.Join(rollbackErrors, "; "))
		}
		createdDirs = append(createdDirs, dir)
	}

	// Write new files.
	for _, dst := range p.Create {
		if err := f.writeFile(dst, p.Files[dst]); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Sprintf("write %q: %v", dst, err))
			if rerr := f.rollbackReplace(p.Backup); rerr != nil {
				rollbackErrors = append(rollbackErrors, fmt.Sprintf("rollback-replace: %v", rerr))
			}
			if rerr := f.rollbackCreate(created); rerr != nil {
				rollbackErrors = append(rollbackErrors, fmt.Sprintf("rollback-created: %v", rerr))
			}
			if rerr := f.rollbackRemoveDirs(createdDirs); rerr != nil {
				rollbackErrors = append(rollbackErrors, fmt.Sprintf("rollback-dirs: %v", rerr))
			}
			return fmt.Errorf("filesystem: write %q: %v; rollback errors: %v", dst, err, strings.Join(rollbackErrors, "; "))
		}
		created = append(created, dst)
	}

	// Write replaced files.
	for _, dst := range p.Replace {
		if err := f.writeFile(dst, p.Files[dst]); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Sprintf("write %q: %v", dst, err))
			if rerr := f.rollbackReplace(p.Backup); rerr != nil {
				rollbackErrors = append(rollbackErrors, fmt.Sprintf("rollback-replace: %v", rerr))
			}
			if rerr := f.rollbackCreate(created); rerr != nil {
				rollbackErrors = append(rollbackErrors, fmt.Sprintf("rollback-created: %v", rerr))
			}
			if rerr := f.rollbackRemoveDirs(createdDirs); rerr != nil {
				rollbackErrors = append(rollbackErrors, fmt.Sprintf("rollback-dirs: %v", rerr))
			}
			return fmt.Errorf("filesystem: write %q: %v; rollback errors: %v", dst, err, strings.Join(rollbackErrors, "; "))
		}
	}

	// Success: discard backups.
	var cleanupErrors []string
	for _, backup := range p.Backup {
		if err := os.Remove(backup); err != nil && !errors.Is(err, os.ErrNotExist) {
			cleanupErrors = append(cleanupErrors, err.Error())
		}
	}
	if len(cleanupErrors) > 0 {
		return fmt.Errorf("filesystem: commit succeeded but backup cleanup failed: %s", strings.Join(cleanupErrors, "; "))
	}
	return nil
}

func (f *RealFS) rollbackCreate(paths []string) error {
	var first error
	for i := len(paths) - 1; i >= 0; i-- {
		if err := os.Remove(paths[i]); err != nil && !errors.Is(err, os.ErrNotExist) {
			if first == nil {
				first = err
			}
		}
	}
	return first
}

func (f *RealFS) rollbackReplace(backup map[string]string) error {
	var first error
	// Sort backup keys for deterministic restore order.
	origs := make([]string, 0, len(backup))
	for orig := range backup {
		origs = append(origs, orig)
	}
	sortStrings(origs)

	// Restore in reverse order (last backed up first restored).
	for i := len(origs) - 1; i >= 0; i-- {
		orig := origs[i]
		b := backup[orig]
		// Remove the (possibly partially written) replacement.
		_ = os.Remove(orig)
		if err := os.Rename(b, orig); err != nil {
			if first == nil {
				first = fmt.Errorf("restore %q from %q: %w", orig, b, err)
			}
		}
	}
	return first
}

func (f *RealFS) rollbackRemoveDirs(dirs []string) error {
	var first error
	// Remove in reverse order (deepest first).
	for i := len(dirs) - 1; i >= 0; i-- {
		if err := os.Remove(dirs[i]); err != nil && !errors.Is(err, os.ErrNotExist) {
			// Ignore "directory not empty" errors; only fail on real errors.
			if !errors.Is(err, os.ErrExist) && !strings.Contains(err.Error(), "not empty") {
				if first == nil {
					first = err
				}
			}
		}
	}
	return first
}

// writeFile creates parent directories and writes content to dst.
func (f *RealFS) writeFile(dst string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0o644)
}

// resolveArtifactPath validates and resolves a relative artifact path.
func (f *RealFS) resolveArtifactPath(rel string) (string, error) {
	rel = filepath.ToSlash(filepath.Clean(rel))
	rel = strings.ReplaceAll(rel, "\\", "/")
	if rel == "." || rel == ".." || strings.HasPrefix(rel, "/") || strings.HasPrefix(rel, "../") {
		return "", fmt.Errorf("filesystem: path %q escapes target root", rel)
	}
	abs, err := filepath.Abs(filepath.Join(f.Root, rel))
	if err != nil {
		return "", err
	}
	rootAbs := f.Root
	if !filepath.IsAbs(rootAbs) {
		rootAbs, err = filepath.Abs(rootAbs)
		if err != nil {
			return "", err
		}
	}
	if abs != rootAbs && !strings.HasPrefix(abs, rootAbs+string(filepath.Separator)) {
		return "", fmt.Errorf("filesystem: path %q escapes target root", rel)
	}
	return abs, nil
}

// checkSymlinkParents verifies that no parent directory component of the
// target path is a symlink, which could redirect writes outside the root.
func (f *RealFS) checkSymlinkParents(target string) error {
	dir := filepath.Dir(target)
	rootAbs := f.Root
	if !filepath.IsAbs(rootAbs) {
		var err error
		rootAbs, err = filepath.Abs(rootAbs)
		if err != nil {
			return err
		}
	}
	// Walk up from the target's parent directory to the root.
	for cur := dir; cur != rootAbs && cur != filepath.Dir(rootAbs); cur = filepath.Dir(cur) {
		info, err := os.Lstat(cur)
		if err != nil {
			// If we can't stat the parent (doesn't exist, or a path component
			// is a regular file -> ENOTDIR), skip the symlink check. Commit
			// will surface the real filesystem error when it tries to create
			// the directory or write the file.
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("filesystem: parent directory %q is a symlink, refusing to write outside target", cur)
		}
		if cur == filepath.Dir(cur) {
			break // reached filesystem root
		}
	}
	return nil
}

func (f *RealFS) validateRoot() error {
	if strings.TrimSpace(f.Root) == "" {
		return errors.New("filesystem: root is not configured")
	}
	info, err := os.Stat(f.Root)
	if err != nil {
		return fmt.Errorf("filesystem: root %q: %w", f.Root, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("filesystem: root %q is not a directory", f.Root)
	}
	return nil
}

func dirExists(dirs []string, dir string) bool {
	for _, d := range dirs {
		if d == dir {
			return true
		}
	}
	return false
}

func sortStrings(s []string) {
	sort.Strings(s)
}
