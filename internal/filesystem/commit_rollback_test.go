package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

// These tests drive RealFS.Commit through real I/O failures that behave the same
// on Linux, macOS, and Windows: a regular file standing where a directory is needed.

func setupBlocker(t *testing.T) (root, blocker string, f *RealFS) {
	t.Helper()
	root = t.TempDir()
	blocker = filepath.Join(root, "blocker")
	if err := os.WriteFile(blocker, []byte("i am a file"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root, blocker, &RealFS{Root: root}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func TestRealFS_Commit_WriteFailureRemovesCreatedFiles(t *testing.T) {
	root, blocker, f := setupBlocker(t)
	a := filepath.Join(root, "a.txt")
	bad := filepath.Join(blocker, "b.txt") // parent is a regular file -> write fails
	p := Plan{
		Create: []string{a, bad},
		Files:  map[string][]byte{a: []byte("A"), bad: []byte("B")},
	}
	if err := f.Commit(p); err == nil {
		t.Fatal("expected commit to fail")
	}
	if exists(a) {
		t.Error("file created before the failure must be rolled back")
	}
}

func TestRealFS_Commit_ReplaceWriteFailureRestoresBackup(t *testing.T) {
	root, blocker, f := setupBlocker(t)
	old := filepath.Join(root, "old.txt")
	bak := filepath.Join(root, "old.txt.bak")
	if err := os.WriteFile(old, []byte("OLD"), 0o644); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(blocker, "x.txt")
	p := Plan{
		Replace: []string{bad},
		Backup:  map[string]string{old: bak},
		Files:   map[string][]byte{bad: []byte("NEW")},
	}
	if err := f.Commit(p); err == nil {
		t.Fatal("expected commit to fail")
	}
	if got := mustRead(t, old); got != "OLD" {
		t.Errorf("old file not restored, got %q", got)
	}
	if exists(bak) {
		t.Error("backup file must be consumed by the restore")
	}
}

func TestRealFS_Commit_MkdirFailureRemovesCreatedDirs(t *testing.T) {
	root, blocker, f := setupBlocker(t)
	newDir := filepath.Join(root, "newdir")
	p := Plan{Dirs: []string{newDir, filepath.Join(blocker, "sub")}}
	if err := f.Commit(p); err == nil {
		t.Fatal("expected commit to fail")
	}
	if exists(newDir) {
		t.Error("directory created before the failure must be removed")
	}
}

// Regression: when the second backup rename failed, the backup that had already
// succeeded was never restored, leaving the user's original file renamed to *.bak.
func TestRealFS_Commit_BackupFailureRestoresEarlierBackups(t *testing.T) {
	root, _, f := setupBlocker(t)
	a := filepath.Join(root, "a.txt")
	if err := os.WriteFile(a, []byte("A"), 0o644); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(root, "b.txt") // does not exist -> rename fails
	p := Plan{
		Backup: map[string]string{a: a + ".bak", missing: missing + ".bak"},
	}
	if err := f.Commit(p); err == nil {
		t.Fatal("expected commit to fail")
	}
	if !exists(a) {
		t.Fatal("original a.txt was lost: earlier backup was not rolled back")
	}
	if got := mustRead(t, a); got != "A" {
		t.Errorf("a.txt content changed: %q", got)
	}
	if exists(a + ".bak") {
		t.Error("stray backup left behind")
	}
}

func TestRealFS_Commit_SuccessReplacesAndDiscardsBackup(t *testing.T) {
	root, _, f := setupBlocker(t)
	old := filepath.Join(root, "old.txt")
	bak := old + ".bak"
	if err := os.WriteFile(old, []byte("OLD"), 0o644); err != nil {
		t.Fatal(err)
	}
	p := Plan{
		Replace: []string{old},
		Backup:  map[string]string{old: bak},
		Files:   map[string][]byte{old: []byte("NEW")},
	}
	if err := f.Commit(p); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if got := mustRead(t, old); got != "NEW" {
		t.Errorf("content = %q", got)
	}
	if exists(bak) {
		t.Error("backup must be discarded after success")
	}
}
