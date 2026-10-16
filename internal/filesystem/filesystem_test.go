package filesystem

import (
	"errors"
	"strings"
	"testing"
)

func TestPlanBuild_NewFiles(t *testing.T) {
	m := NewMemFS()
	p, err := m.PlanBuild(map[string][]byte{
		"Dockerfile":               []byte("FROM scratch"),
		".github/workflows/ci.yml": []byte("name: CI"),
	})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if len(p.Create) != 2 {
		t.Fatalf("expected 2 create, got %d (%v)", len(p.Create), p.Create)
	}
	if len(p.Conflicts) != 0 {
		t.Fatalf("expected no conflicts, got %v", p.Conflicts)
	}
}

func TestPlanBuild_ConflictWithoutForce(t *testing.T) {
	m := NewMemFS()
	m.Put("Dockerfile", []byte("old"))
	p, err := m.PlanBuild(map[string][]byte{"Dockerfile": []byte("new")})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if len(p.Conflicts) != 1 || p.Conflicts[0] != "Dockerfile" {
		t.Fatalf("expected conflict on Dockerfile, got %v", p.Conflicts)
	}
	if len(p.Backup) != 1 {
		t.Fatalf("expected backup entry, got %d", len(p.Backup))
	}
}

func TestPlanBuild_SymlinkRejected(t *testing.T) {
	m := NewMemFS()
	m.Symlink("Dockerfile", "target")
	_, err := m.PlanBuild(map[string][]byte{"Dockerfile": []byte("new")})
	if err == nil {
		t.Fatal("expected symlink rejection")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("error should mention symlink: %v", err)
	}
}

func TestPlanBuild_DirectoryRejected(t *testing.T) {
	m := NewMemFS()
	m.Mkdir("k8s")
	_, err := m.PlanBuild(map[string][]byte{"k8s": []byte("x")})
	if err == nil {
		t.Fatal("expected directory overwrite rejection")
	}
}

func TestPlanBuild_TraversalRejected(t *testing.T) {
	m := NewMemFS()
	bad := []string{"../escape", "/etc/passwd", "out/../../etc", ".."}
	for _, p := range bad {
		_, err := m.PlanBuild(map[string][]byte{p: []byte("x")})
		if err == nil {
			t.Fatalf("expected traversal error for %q", p)
		}
	}
}

func TestPlanBuild_BackupCollision(t *testing.T) {
	m := NewMemFS()
	m.Put("Dockerfile", []byte("old"))
	m.Put("Dockerfile.devopsgen.bak", []byte("existing backup"))
	_, err := m.PlanBuild(map[string][]byte{"Dockerfile": []byte("new")})
	if err == nil {
		t.Fatal("expected backup collision error")
	}
	if !strings.Contains(err.Error(), "backup") {
		t.Fatalf("error should mention backup: %v", err)
	}
}

func TestPlanBuild_TracksDirs(t *testing.T) {
	m := NewMemFS()
	p, err := m.PlanBuild(map[string][]byte{
		"kubernetes/deployment.yaml": []byte("apiVersion: apps/v1"),
	})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if len(p.Dirs) == 0 {
		t.Fatal("expected directories to be tracked for creation")
	}
	found := false
	for _, d := range p.Dirs {
		if d == "kubernetes" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected kubernetes dir in Dirs, got %v", p.Dirs)
	}
}

func TestCommit_NewFiles(t *testing.T) {
	m := NewMemFS()
	p, _ := m.PlanBuild(map[string][]byte{
		"Dockerfile":               []byte("FROM scratch"),
		".github/workflows/ci.yml": []byte("name: CI"),
	})
	if err := m.Commit(p); err != nil {
		t.Fatalf("commit: %v", err)
	}
	c, ok := m.Get("Dockerfile")
	if !ok || string(c) != "FROM scratch" {
		t.Fatalf("Dockerfile not written: %v %q", ok, c)
	}
	c, ok = m.Get(".github/workflows/ci.yml")
	if !ok || string(c) != "name: CI" {
		t.Fatalf("workflow not written: %v %q", ok, c)
	}
}

func TestCommit_ReplaceWithBackup(t *testing.T) {
	m := NewMemFS()
	m.Put("Dockerfile", []byte("old"))
	p, _ := m.PlanBuild(map[string][]byte{"Dockerfile": []byte("new")})
	if err := m.Commit(p); err != nil {
		t.Fatalf("commit: %v", err)
	}
	c, _ := m.Get("Dockerfile")
	if string(c) != "new" {
		t.Fatalf("expected new content, got %q", c)
	}
	if _, ok := m.Get("Dockerfile.devopsgen.bak"); ok {
		t.Fatal("backup should be removed after successful commit")
	}
}

func TestCommit_WriteFailureRollsBack(t *testing.T) {
	m := NewMemFS()
	m.WriteErr = errors.New("simulated write failure")
	p, _ := m.PlanBuild(map[string][]byte{
		"Dockerfile": []byte("new"),
	})
	err := m.Commit(p)
	if err == nil {
		t.Fatal("expected commit error")
	}
	if !strings.Contains(err.Error(), "write") {
		t.Fatalf("error should mention write: %v", err)
	}
	if !strings.Contains(err.Error(), "rollback") {
		t.Fatalf("error should mention rollback: %v", err)
	}
	// New file should have been cleaned up.
	if _, ok := m.Get("Dockerfile"); ok {
		t.Fatal("file should have been rolled back")
	}
}

func TestCommit_BackupFailureRollsBack(t *testing.T) {
	m := NewMemFS()
	m.Put("Dockerfile", []byte("old"))
	m.BackupErr = errors.New("simulated backup failure")
	p, _ := m.PlanBuild(map[string][]byte{"Dockerfile": []byte("new")})
	err := m.Commit(p)
	if err == nil {
		t.Fatal("expected commit error")
	}
	if !strings.Contains(err.Error(), "backup") {
		t.Fatalf("error should mention backup: %v", err)
	}
	// Original file should be restored.
	c, ok := m.Get("Dockerfile")
	if !ok || string(c) != "old" {
		t.Fatalf("original file should be restored after backup failure: %v %q", ok, c)
	}
}

func TestCommit_RestoreFailure(t *testing.T) {
	m := NewMemFS()
	m.Put("Dockerfile", []byte("old"))
	m.RestoreErr = errors.New("simulated restore failure")
	p, _ := m.PlanBuild(map[string][]byte{"Dockerfile": []byte("new")})
	err := m.Commit(p)
	if err == nil {
		t.Fatal("expected commit error")
	}
	if !strings.Contains(err.Error(), "restore") {
		t.Fatalf("error should mention restore: %v", err)
	}
}

func TestDryRun_NoWrites(t *testing.T) {
	m := NewMemFS()
	p, _ := m.PlanBuild(map[string][]byte{"Dockerfile": []byte("new")})
	_, err := m.DryRun(p)
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if _, ok := m.Get("Dockerfile"); ok {
		t.Fatal("dry-run must not write files")
	}
}

func TestMemFS_CommitEmpty(t *testing.T) {
	m := NewMemFS()
	p := Plan{}
	if err := m.Commit(p); err != nil {
		t.Fatalf("empty commit should succeed: %v", err)
	}
}
