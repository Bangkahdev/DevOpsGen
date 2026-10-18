package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInit_EndToEnd(t *testing.T) {
	dir := t.TempDir()
	flags := &InitFlags{
		Stack:     "go",
		AppType:   "binary",
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "bangkah/my-app",
		Port:      8080,
		BuildCmd:  "go build -o server .",
		TestCmd:   "go test ./...",
		Tag:       "v1",
		K8s:       true,
		OutputDir: dir,
	}
	if err := RunInit(flags); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	for _, p := range []string{"Dockerfile", ".github/workflows/ci.yml", "kubernetes/deployment.yaml"} {
		if _, err := os.Stat(filepath.Join(dir, p)); err != nil {
			t.Fatalf("expected artifact %s: %v", p, err)
		}
	}
}

func TestRunInit_ConflictExit3(t *testing.T) {
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
	}
	err := RunInit(flags)
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if !strings.Contains(err.Error(), "conflict") {
		t.Fatalf("error should mention conflict: %v", err)
	}
}

func TestRunInit_ForceOverwrites(t *testing.T) {
	dir := t.TempDir()
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
		Force:     true,
	}
	if err := RunInit(flags); err != nil {
		t.Fatalf("runInit with force: %v", err)
	}
	c, _ := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if !strings.Contains(string(c), "FROM golang") {
		t.Fatalf("expected overwritten Dockerfile, got %q", string(c)[:40])
	}
}

func TestRunInit_DryRunNoWrites(t *testing.T) {
	dir := t.TempDir()
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
		t.Fatalf("runInit dry-run: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "Dockerfile")); err == nil {
		t.Fatal("dry-run must not write files")
	}
}

func TestRunInit_MissingStack(t *testing.T) {
	flags := &InitFlags{
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "a/b",
		Port:      8080,
		OutputDir: t.TempDir(),
	}
	err := RunInit(flags)
	if err == nil {
		t.Fatal("expected error for missing stack")
	}
}

func TestRunInit_BadPort(t *testing.T) {
	flags := &InitFlags{
		Stack:     "go",
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "a/b",
		Port:      99999,
		OutputDir: t.TempDir(),
	}
	err := RunInit(flags)
	if err == nil {
		t.Fatal("expected error for bad port")
	}
}

func TestRunInit_ExitCodes(t *testing.T) {
	// Verify exit codes are surfaced via newExitError without building
	// a separate binary (cross-version Windows exe execution is flaky).
	cases := []struct {
		name string
		fn   func() error
		want int
	}{
		{"missing stack", func() error {
			flags := &InitFlags{
				CI:        "github-actions",
				Registry:  "ghcr",
				Image:     "a/b",
				Port:      8080,
				OutputDir: t.TempDir(),
			}
			return RunInit(flags)
		}, 2},
		{"bad port", func() error {
			flags := &InitFlags{
				Stack:     "go",
				CI:        "github-actions",
				Registry:  "ghcr",
				Image:     "a/b",
				Port:      99999,
				OutputDir: t.TempDir(),
			}
			return RunInit(flags)
		}, 2},
		{"conflict exit 3", func() error {
			dir := t.TempDir()
			_ = os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("old"), 0o644)
			flags := &InitFlags{
				Stack:     "go",
				AppType:   "binary",
				CI:        "github-actions",
				Registry:  "ghcr",
				Image:     "bangkah/my-app",
				Port:      8080,
				OutputDir: dir,
			}
			return RunInit(flags)
		}, 3},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.fn()
			if err == nil {
				t.Fatalf("expected exit code %d, got nil", c.want)
			}
			var ee *exitError
			if !errors.As(err, &ee) {
				t.Fatalf("expected *exitError, got %T", err)
			}
			if ee.ExitCode() != c.want {
				t.Fatalf("exit code: got %d want %d", ee.ExitCode(), c.want)
			}
		})
	}
}
