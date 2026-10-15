package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRealFS_PlanBuildAndCommit(t *testing.T) {
	dir := t.TempDir()
	fs := &RealFS{Root: dir}

	artifacts := map[string][]byte{
		"Dockerfile":               []byte("FROM scratch"),
		".github/workflows/ci.yml": []byte("name: CI"),
	}
	p, err := fs.PlanBuild(artifacts)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if len(p.Create) != 2 {
		t.Fatalf("expected 2 create, got %d", len(p.Create))
	}
	if err := fs.Commit(p); err != nil {
		t.Fatalf("commit: %v", err)
	}
	c, err := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if err != nil || string(c) != "FROM scratch" {
		t.Fatalf("Dockerfile not written: %v %q", err, c)
	}
}

func TestRealFS_PlanBuildTraversal(t *testing.T) {
	dir := t.TempDir()
	fs := &RealFS{Root: dir}
	for _, p := range []string{"../escape", "/etc/passwd", "out/../../etc"} {
		_, err := fs.PlanBuild(map[string][]byte{p: []byte("x")})
		if err == nil {
			t.Fatalf("expected traversal error for %q", p)
		}
	}
}

func TestRealFS_PlanBuildSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "Dockerfile")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	fs := &RealFS{Root: dir}
	_, err := fs.PlanBuild(map[string][]byte{"Dockerfile": []byte("new")})
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("expected symlink rejection, got %v", err)
	}
}

func TestRealFS_PlanBuildConflict(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := &RealFS{Root: dir}
	p, err := fs.PlanBuild(map[string][]byte{"Dockerfile": []byte("new")})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if len(p.Conflicts) != 1 {
		t.Fatalf("expected conflict, got %v", p.Conflicts)
	}
	if err := fs.Commit(p); err != nil {
		t.Fatalf("commit: %v", err)
	}
	c, _ := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if string(c) != "new" {
		t.Fatalf("expected new content, got %q", c)
	}
	if _, err := os.Stat(filepath.Join(dir, "Dockerfile.devopsgen.bak")); err == nil {
		t.Fatal("backup should be removed after successful commit")
	}
}

func TestRealFS_DryRunNoWrites(t *testing.T) {
	dir := t.TempDir()
	fs := &RealFS{Root: dir}
	p, _ := fs.PlanBuild(map[string][]byte{"Dockerfile": []byte("new")})
	if _, err := fs.DryRun(p); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "Dockerfile")); err == nil {
		t.Fatal("dry-run must not write files")
	}
}
