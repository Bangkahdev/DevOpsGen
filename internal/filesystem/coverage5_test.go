package filesystem

import (
	"errors"
	"strings"
	"testing"
)

func TestMemFS_Commit_RollbackReplaceError(t *testing.T) {
	m := NewMemFS()
	m.Put("Dockerfile", []byte("old"))
	p, _ := m.PlanBuild(map[string][]byte{"Dockerfile": []byte("new")})
	m.RestoreErr = errors.New("restore failed")
	err := m.Commit(p)
	if err == nil || !strings.Contains(err.Error(), "restore failed") {
		t.Fatalf("expected restore error: %v", err)
	}
}

func TestMemFS_Commit_RollbackRemoveDirsError(t *testing.T) {
	m := NewMemFS()
	p, _ := m.PlanBuild(map[string][]byte{"newdir/Dockerfile": []byte("new")})
	m.WriteErr = errors.New("write failed")
	err := m.Commit(p)
	if err == nil || !strings.Contains(err.Error(), "rollback") {
		t.Fatalf("expected rollback error: %v", err)
	}
}

func TestRealFS_CheckSymlinkParent_RootReached(t *testing.T) {
	dir := t.TempDir()
	fs := &RealFS{Root: dir}
	err := fs.checkSymlinkParents(dir)
	if err != nil {
		t.Fatalf("expected no error for root, got: %v", err)
	}
}

func TestSortStrings_Empty(t *testing.T) {
	s := []string{}
	sortStrings(s)
	if len(s) != 0 {
		t.Fatal("expected empty slice")
	}
}

func TestSortStrings_Single(t *testing.T) {
	s := []string{"b"}
	sortStrings(s)
	if s[0] != "b" {
		t.Fatal("expected single element unchanged")
	}
}

func TestSortStrings_AlreadySorted(t *testing.T) {
	s := []string{"a", "b", "c"}
	sortStrings(s)
	for i := 0; i < len(s)-1; i++ {
		if s[i] > s[i+1] {
			t.Fatal("not sorted")
		}
	}
}

func TestSortStrings_ReverseSorted(t *testing.T) {
	s := []string{"c", "b", "a"}
	sortStrings(s)
	for i := 0; i < len(s)-1; i++ {
		if s[i] > s[i+1] {
			t.Fatal("not sorted")
		}
	}
}

func TestMemFS_FileInfoMethods(t *testing.T) {
	m := NewMemFS()
	m.Put("test.txt", []byte("hello"))
	fi, _, err := m.stat("test.txt")
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if fi.Name() != "" {
		t.Fatalf("expected empty name, got %q", fi.Name())
	}
	if fi.IsDir() {
		t.Fatal("expected not dir")
	}
	if fi.Size() != 5 {
		t.Fatalf("expected size 5, got %d", fi.Size())
	}
	_ = fi.ModTime()
	_ = fi.Sys()
}

func TestMemFS_FileInfoDir(t *testing.T) {
	m := NewMemFS()
	m.Mkdir("testdir")
	fi, _, err := m.stat("testdir")
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if !fi.IsDir() {
		t.Fatal("expected is dir")
	}
}
