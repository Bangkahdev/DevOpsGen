package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunInit_DryRunWithConflict ensures --dry-run does not require --force
// when target files already exist. Dry-run only reports the plan and never
// writes, so conflicts must not block it.
func TestRunInit_DryRunWithConflict(t *testing.T) {
	dir := t.TempDir()
	// Pre-create Dockerfile to force a conflict.
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	flags := &InitFlags{
		Stack:     "go",
		AppType:   "binary",
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "bangkah/my-app",
		Port:      8080,
		OutputDir: dir,
		DryRun:    true,
	}
	if err := RunInit(flags); err != nil {
		t.Fatalf("dry-run with conflict should not error: %v", err)
	}
	// Dry-run must not write anything.
	if c, _ := os.ReadFile(filepath.Join(dir, "Dockerfile")); string(c) != "old" {
		t.Fatal("dry-run must not overwrite existing files")
	}
}

// TestRunInit_PythonStack ensures Python stack generates Dockerfile + workflow.
func TestRunInit_PythonStack(t *testing.T) {
	dir := t.TempDir()
	flags := &InitFlags{
		Stack:     "python",
		AppType:   "fastapi",
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "bangkah/my-app",
		Port:      8000,
		StartCmd:  "uvicorn app:app --host 0.0.0.0 --port 8000",
		OutputDir: dir,
	}
	if err := RunInit(flags); err != nil {
		t.Fatalf("runInit python: %v", err)
	}
	for _, p := range []string{"Dockerfile", ".github/workflows/ci.yml"} {
		if _, err := os.Stat(filepath.Join(dir, p)); err != nil {
			t.Fatalf("expected artifact %s: %v", p, err)
		}
	}
	df, _ := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if !strings.Contains(string(df), "/opt/venv") {
		t.Fatalf("python Dockerfile should use /opt/venv, got %q", string(df)[:200])
	}
	wf, _ := os.ReadFile(filepath.Join(dir, ".github/workflows/ci.yml"))
	if !strings.Contains(string(wf), "setup-python") {
		t.Fatalf("python workflow should have setup-python step, got %q", string(wf)[:200])
	}
}
