package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRealFS_PlanBuild_DirCreation(t *testing.T) {
	dir := t.TempDir()
	fs := &RealFS{Root: dir}
	p, err := fs.PlanBuild(map[string][]byte{
		"a/b/c/Dockerfile": []byte("FROM scratch"),
	})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if len(p.Dirs) == 0 {
		t.Fatal("expected dirs to be tracked")
	}
	if err := fs.Commit(p); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "a", "b", "c", "Dockerfile")); err != nil {
		t.Fatal("nested file should exist")
	}
}

func TestRealFS_PlanBuild_ReplaceFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := &RealFS{Root: dir}
	p, err := fs.PlanBuild(map[string][]byte{"Dockerfile": []byte("new")})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if len(p.Replace) != 1 {
		t.Fatalf("expected 1 replace, got %d", len(p.Replace))
	}
	if err := fs.Commit(p); err != nil {
		t.Fatalf("commit: %v", err)
	}
	c, _ := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if string(c) != "new" {
		t.Fatalf("expected new content, got %q", string(c))
	}
}

func TestMemFS_Commit_BackupCollision(t *testing.T) {
	m := NewMemFS()
	m.Put("Dockerfile", []byte("old"))
	m.Put("Dockerfile.devopsgen.bak", []byte("existing"))
	_, err := m.PlanBuild(map[string][]byte{"Dockerfile": []byte("new")})
	if err == nil || !strings.Contains(err.Error(), "backup") {
		t.Fatalf("expected backup collision error: %v", err)
	}
}

func TestMemFS_Commit_WriteErr(t *testing.T) {
	m := NewMemFS()
	m.WriteErr = os.ErrPermission
	p, _ := m.PlanBuild(map[string][]byte{"Dockerfile": []byte("new")})
	err := m.Commit(p)
	if err == nil || !strings.Contains(err.Error(), "write") {
		t.Fatalf("expected write error: %v", err)
	}
}

func TestMemFS_Commit_RestoreErr(t *testing.T) {
	m := NewMemFS()
	m.Put("Dockerfile", []byte("old"))
	m.RestoreErr = os.ErrPermission
	p, _ := m.PlanBuild(map[string][]byte{"Dockerfile": []byte("new")})
	err := m.Commit(p)
	if err == nil || !strings.Contains(err.Error(), "restore") {
		t.Fatalf("expected restore error: %v", err)
	}
}
