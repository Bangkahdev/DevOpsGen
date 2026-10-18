package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRealFS_Commit_WriteFailureRollbackCreate(t *testing.T) {
	dir := t.TempDir()
	// Create a file where a directory is needed, so MkdirAll fails.
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := &RealFS{Root: dir}
	p, err := fs.PlanBuild(map[string][]byte{"blocker/sub/Dockerfile": []byte("new")})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	err = fs.Commit(p)
	if err == nil {
		t.Fatal("expected commit error when parent is a file")
	}
}

func TestRealFS_Commit_WriteFailureRollbackReplace(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := &RealFS{Root: dir}
	// Replace Dockerfile but target a path under a file blocker.
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := fs.PlanBuild(map[string][]byte{"blocker/Dockerfile": []byte("new")})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	err = fs.Commit(p)
	if err == nil {
		t.Fatal("expected commit error")
	}
}

func TestRealFS_WriteFile_MkdirFailure(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := &RealFS{Root: dir}
	err := fs.writeFile(filepath.Join(dir, "blocker", "Dockerfile"), []byte("x"))
	if err == nil {
		t.Fatal("expected writeFile error when parent is a file")
	}
}
