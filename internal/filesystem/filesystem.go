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
	for rel, content := range artifacts {
		clean, err := f.resolveArtifactPath(rel)
		if err != nil {
			return Plan{}, err
		}
		p.Files[clean] = content
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
			p.Backup[clean] = clean + ".devopsgen.bak"
		} else if errors.Is(statErr, os.ErrNotExist) {
			p.Create = append(p.Create, clean)
		} else {
			return Plan{}, fmt.Errorf("filesystem: stat %q: %w", clean, statErr)
		}
	}
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
	rollbackFailed := false

	// Backup existing files before replacing.
	for orig, backup := range p.Backup {
		if err := os.Rename(orig, backup); err != nil {
			return fmt.Errorf("filesystem: backup %q: %w", orig, err)
		}
	}

	// Write new files.
	for _, dst := range p.Create {
		if err := f.writeFile(dst, p.Files[dst]); err != nil {
			rollbackFailed = true
			_ = f.rollbackCreate(created)
			_ = f.rollbackReplace(p.Backup)
			return fmt.Errorf("filesystem: write %q: %w", dst, err)
		}
		created = append(created, dst)
	}

	// Write replaced files (content was already rendered in memory).
	for _, dst := range p.Replace {
		if err := f.writeFile(dst, p.Files[dst]); err != nil {
			rollbackFailed = true
			_ = f.rollbackCreate(created)
			_ = f.rollbackReplace(p.Backup)
			return fmt.Errorf("filesystem: write %q: %w", dst, err)
		}
	}

	// Success: discard backups. If a restore step fails we roll back.
	for _, backup := range p.Backup {
		_ = os.Remove(backup)
	}
	if rollbackFailed {
		return errors.New("filesystem: rollback failed")
	}
	return nil
}

func (f *RealFS) rollbackCreate(paths []string) error {
	var first error
	for i := len(paths) - 1; i >= 0; i-- {
		if err := os.Remove(paths[i]); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (f *RealFS) rollbackReplace(backup map[string]string) error {
	var first error
	for orig, b := range backup {
		if _, err := os.Stat(orig); err == nil {
			_ = os.Remove(orig)
		}
		if err := os.Rename(b, orig); err != nil && first == nil {
			first = err
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
