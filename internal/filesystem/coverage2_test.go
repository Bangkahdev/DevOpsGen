package filesystem

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRealFS_Commit_BackupFailure(t *testing.T) {
	dir := t.TempDir()
	fs := &RealFS{Root: dir}
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile.devopsgen.bak"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := fs.PlanBuild(map[string][]byte{"Dockerfile": []byte("new")})
	if err == nil || !strings.Contains(err.Error(), "backup") {
		t.Fatalf("expected backup collision: %v", err)
	}
}

func TestMemFS_Commit_MultipleCreate(t *testing.T) {
	m := NewMemFS()
	p, _ := m.PlanBuild(map[string][]byte{
		"a/Dockerfile": []byte("FROM scratch"),
		"b/ci.yml":     []byte("name: CI"),
	})
	if err := m.Commit(p); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if _, ok := m.Get("a/Dockerfile"); !ok {
		t.Fatal("a/Dockerfile not written")
	}
	if _, ok := m.Get("b/ci.yml"); !ok {
		t.Fatal("b/ci.yml not written")
	}
}

func TestMemFS_Commit_ReplaceThenRollback(t *testing.T) {
	m := NewMemFS()
	m.Put("Dockerfile", []byte("old"))
	p, _ := m.PlanBuild(map[string][]byte{"Dockerfile": []byte("new")})
	m.WriteErr = errors.New("write failed")
	err := m.Commit(p)
	if err == nil || !strings.Contains(err.Error(), "rollback") {
		t.Fatalf("expected rollback error: %v", err)
	}
	c, ok := m.Get("Dockerfile")
	if !ok || string(c) != "old" {
		t.Fatalf("old file should be restored after rollback: %v %q", ok, c)
	}
}

func TestRealFS_ResolveArtifactPath_Absolute(t *testing.T) {
	fs := &RealFS{Root: "/tmp"}
	_, err := fs.resolveArtifactPath("/etc/passwd")
	if err == nil || !strings.Contains(err.Error(), "escapes") {
		t.Fatalf("expected escape error: %v", err)
	}
}

func TestRealFS_ResolveArtifactPath_DotDot(t *testing.T) {
	fs := &RealFS{Root: "/tmp"}
	_, err := fs.resolveArtifactPath("../escape")
	if err == nil || !strings.Contains(err.Error(), "escapes") {
		t.Fatalf("expected escape error: %v", err)
	}
}

func TestMemFS_Stat_NotFound(t *testing.T) {
	m := NewMemFS()
	_, _, err := m.stat("nonexistent")
	if !errors.Is(err, errNotFound) {
		t.Fatalf("expected errNotFound: %v", err)
	}
}
