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

// TestRunInit_NodeStack ensures Node.js stack generates Dockerfile + workflow.
func TestRunInit_NodeStack(t *testing.T) {
	dir := t.TempDir()
	flags := &InitFlags{
		Stack:     "nodejs",
		AppType:   "express",
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "bangkah/my-app",
		Port:      3000,
		StartCmd:  "node server.js",
		OutputDir: dir,
	}
	if err := RunInit(flags); err != nil {
		t.Fatalf("runInit nodejs: %v", err)
	}
	for _, p := range []string{"Dockerfile", ".github/workflows/ci.yml"} {
		if _, err := os.Stat(filepath.Join(dir, p)); err != nil {
			t.Fatalf("expected artifact %s: %v", p, err)
		}
	}
	df, _ := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if !strings.Contains(string(df), "node:20-alpine") {
		t.Fatalf("nodejs Dockerfile should use node:20-alpine, got %q", string(df)[:200])
	}
	wf, _ := os.ReadFile(filepath.Join(dir, ".github/workflows/ci.yml"))
	if !strings.Contains(string(wf), "setup-node") {
		t.Fatalf("nodejs workflow should have setup-node step, got %q", string(wf)[:200])
	}
}
