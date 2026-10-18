package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRealFS_ValidateRoot_Empty(t *testing.T) {
	fs := &RealFS{Root: ""}
	if err := fs.validateRoot(); err == nil {
		t.Fatal("expected error for empty root")
	}
}

func TestRealFS_ValidateRoot_NotDir(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := &RealFS{Root: file}
	if err := fs.validateRoot(); err == nil {
		t.Fatal("expected error for non-directory root")
	}
}

func TestRealFS_ValidateRoot_NotExist(t *testing.T) {
	fs := &RealFS{Root: "/nonexistent/path/xyz"}
	if err := fs.validateRoot(); err == nil {
		t.Fatal("expected error for non-existent root")
	}
}

func TestMemFS_Get_Missing(t *testing.T) {
	m := NewMemFS()
	if _, ok := m.Get("nonexistent"); ok {
		t.Fatal("expected not found")
	}
}
