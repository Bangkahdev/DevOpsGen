package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInit_DryRunConflictExit3(t *testing.T) {
	outDir := t.TempDir()
	// A pre-existing Dockerfile conflicts with the plan; dry-run must report
	// exit 3 without touching it.
	conflict := filepath.Join(outDir, "Dockerfile")
	if err := os.WriteFile(conflict, []byte("FROM scratch\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	flags := &InitFlags{
		Stack:     "go",
		AppType:   "binary",
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "a/b",
		Port:      8080,
		TestCmd:   "go test ./...",
		OutputDir: outDir,
		DryRun:    true,
	}
	err := RunInit(flags)
	if err == nil {
		t.Fatal("expected dry-run with conflict to fail")
	}
	var ee *exitError
	if !errors.As(err, &ee) || ee.ExitCode() != 3 {
		t.Fatalf("expected exit 3, got %v", err)
	}
	got, err := os.ReadFile(conflict)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "FROM scratch\n" {
		t.Fatal("dry-run must not modify the conflicting file")
	}
	if _, err := os.Stat(filepath.Join(outDir, ".github", "workflows", "ci.yml")); !os.IsNotExist(err) {
		t.Fatal("dry-run must not write any artifacts")
	}
}

func TestRunInit_DryRunForceNoConflict(t *testing.T) {
	outDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(outDir, "Dockerfile"), []byte("FROM scratch\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	flags := &InitFlags{
		Stack:     "go",
		AppType:   "binary",
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "a/b",
		Port:      8080,
		TestCmd:   "go test ./...",
		OutputDir: outDir,
		DryRun:    true,
		Force:     true,
	}
	if err := RunInit(flags); err != nil {
		t.Fatalf("dry-run with --force should be conflict-free, got: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(outDir, "Dockerfile"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "FROM scratch\n" {
		t.Fatal("dry-run with --force must still not modify files")
	}
}

func TestRunInit_HarvestRespectsExplicitBuildCmd(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	outDir := t.TempDir()
	flags := &InitFlags{
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "a/b",
		Port:      8080,
		BuildCmd:  "go build ./cmd/devopsgen",
		OutputDir: outDir,
	}
	if err := RunInit(flags); err != nil {
		t.Fatalf("expected harvest to run detection but keep explicit build-cmd, got: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(outDir, "Dockerfile"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "./cmd/devopsgen") {
		t.Fatalf("explicit build-cmd must win over detection, Dockerfile:\n%s", content)
	}
}

func TestRunInit_NonInteractiveRequiresStackExit2(t *testing.T) {
	flags := &InitFlags{
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "a/b",
		Port:      8080,
		OutputDir: t.TempDir(),
	}
	err := RunInit(flags)
	if err == nil {
		t.Fatal("stack is required in non-interactive mode")
	}
	var ee *exitError
	if !errors.As(err, &ee) || ee.ExitCode() != 2 {
		t.Fatalf("expected exit 2, got %v", err)
	}
	if !strings.Contains(err.Error(), "--stack") {
		t.Fatalf("error should tell the user to provide --stack, got: %v", err)
	}
}

func TestRootCommand_FlagParseErrorExit2(t *testing.T) {
	root := NewRootCommand()
	root.SetArgs([]string{"init", "--port", "not-a-number"})
	err := root.Execute()
	code := mapCommandError(err)
	if code != 2 {
		t.Fatalf("expected exit 2 for flag parse error, got %v (%v)", code, err)
	}
	if !strings.Contains(err.Error(), "port") {
		t.Fatalf("error should mention the offending flag: %v", err)
	}
}

func TestRootCommand_UnknownCommandExit2(t *testing.T) {
	root := NewRootCommand()
	root.SetArgs([]string{"frobnicate"})
	err := root.Execute()
	code := mapCommandError(err)
	if code != 2 {
		t.Fatalf("expected exit 2 for unknown command, got %v (%v)", code, err)
	}
}

func TestRootCommand_UnknownFlagExit2(t *testing.T) {
	root := NewRootCommand()
	root.SetArgs([]string{"init", "--kraft"})
	err := root.Execute()
	code := mapCommandError(err)
	if code != 2 {
		t.Fatalf("expected exit 2 for unknown flag, got %v (%v)", code, err)
	}
}

func TestRootCommand_VersionTakesNoArgs(t *testing.T) {
	root := NewRootCommand()
	root.SetArgs([]string{"version", "unexpected", "args"})
	err := root.Execute()
	code := mapCommandError(err)
	if code != 2 {
		t.Fatalf("expected exit 2 for version with args, got %v (%v)", code, err)
	}
}
