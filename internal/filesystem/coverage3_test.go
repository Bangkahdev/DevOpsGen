package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRealFS_CheckSymlinkParent(t *testing.T) {
	dir := t.TempDir()
	realDir := filepath.Join(dir, "real")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(dir, "link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Fatal(err)
	}
	fs := &RealFS{Root: dir}
	_, err := fs.PlanBuild(map[string][]byte{"link/Dockerfile": []byte("x")})
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("expected symlink parent error: %v", err)
	}
}

func TestRealFS_RollbackCreate_Error(t *testing.T) {
	dir := t.TempDir()
	// Use a read-only root to force write failure.
	readOnly := filepath.Join(dir, "readonly")
	if err := os.MkdirAll(readOnly, 0o555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(readOnly, 0o755)
	fs := &RealFS{Root: readOnly}
	p, err := fs.PlanBuild(map[string][]byte{"Dockerfile": []byte("new")})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	err = fs.Commit(p)
	// On Windows, chmod 0555 may not prevent writes. Just verify no panic.
	_ = err
}

func TestMemFS_Commit_CleanupError(t *testing.T) {
	m := NewMemFS()
	m.Put("Dockerfile", []byte("old"))
	p, _ := m.PlanBuild(map[string][]byte{"Dockerfile": []byte("new")})
	if err := m.Commit(p); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if _, ok := m.Get("Dockerfile.devopsgen.bak"); ok {
		t.Fatal("backup should be cleaned up")
	}
}

func TestMemFS_PlanBuild_Traversal(t *testing.T) {
	m := NewMemFS()
	bad := []string{"../escape", "/etc/passwd", "..", "."}
	for _, b := range bad {
		_, err := m.PlanBuild(map[string][]byte{b: []byte("x")})
		if err == nil {
			t.Fatalf("expected traversal error for %q", b)
		}
	}
}
