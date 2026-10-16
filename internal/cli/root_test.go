package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDetect(t *testing.T) {
	dir := t.TempDir()
	// write a go.mod so detection succeeds
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	err := runDetect(os.Stdout)
	if err != nil {
		t.Fatalf("runDetect: %v", err)
	}
}

func TestRunDetect_Ambiguous(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"scripts":{"start":"node s.js"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	err := runDetect(os.Stdout)
	if err == nil {
		t.Fatal("expected ambiguity error")
	}
	var ee *exitError
	if !errors.As(err, &ee) || ee.ExitCode() != 5 {
		t.Fatalf("expected exit 5, got %v", err)
	}
}

func TestRunValidate_Pass(t *testing.T) {
	dir := t.TempDir()
	arts := []string{
		"Dockerfile:FROM golang:1.22\nEXPOSE 8080",
		".github/workflows/ci.yml:name: CI\non:\n  push:\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n      - run: docker push x:${{ github.sha }}",
	}
	for _, a := range arts {
		parts := strings.SplitN(a, ":", 2)
		if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(parts[0])), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, parts[0]), []byte(parts[1]), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := runValidate(dir, os.Stdout); err != nil {
		t.Fatalf("runValidate: %v", err)
	}
}

func TestRunValidate_Fail(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("EXPOSE 8080"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := runValidate(dir, os.Stdout)
	if err == nil {
		t.Fatal("expected validation failure")
	}
	var ee *exitError
	if !errors.As(err, &ee) || ee.ExitCode() != 4 {
		t.Fatalf("expected exit 4, got %v", err)
	}
}

func TestRunValidate_Empty(t *testing.T) {
	dir := t.TempDir()
	err := runValidate(dir, os.Stdout)
	if err == nil {
		t.Fatal("expected error for no artifacts")
	}
	var ee *exitError
	if !errors.As(err, &ee) || ee.ExitCode() != 4 {
		t.Fatalf("expected exit 4, got %v", err)
	}
}

func TestRunValidate_MissingDir(t *testing.T) {
	err := runValidate("/nonexistent-dir-xyz", os.Stdout)
	if err == nil {
		t.Fatal("expected error for missing dir")
	}
	var ee *exitError
	if !errors.As(err, &ee) || ee.ExitCode() != 1 {
		t.Fatalf("expected exit 1, got %v", err)
	}
}

func TestListCommand(t *testing.T) {
	cmd := newListCommand()
	if err := cmd.Flags().Parse(nil); err != nil {
		t.Fatal(err)
	}
}

func TestVersionCommand(t *testing.T) {
	cmd := newVersionCommand()
	if err := cmd.Flags().Parse(nil); err != nil {
		t.Fatal(err)
	}
}

func TestDetectCommandFlags(t *testing.T) {
	cmd := newDetectCommand()
	if err := cmd.Flags().Parse(nil); err != nil {
		t.Fatal(err)
	}
}

func TestValidateCommandFlags(t *testing.T) {
	cmd := newValidateCommand()
	if err := cmd.Flags().Parse([]string{"--dir", "/tmp"}); err != nil {
		t.Fatal(err)
	}
	v, _ := cmd.Flags().GetString("dir")
	if v != "/tmp" {
		t.Fatalf("dir flag: %q", v)
	}
}

func TestInitCommandFlags(t *testing.T) {
	cmd := newInitCommand()
	args := []string{
		"--stack", "go", "--ci", "github-actions", "--registry", "ghcr",
		"--image", "a/b", "--port", "8080", "--tag", "v1",
		"--k8s", "--ingress", "--ingress-host", "app.example.com",
		"--health-path", "/health", "--output-dir", "/tmp",
		"--force", "--dry-run",
	}
	if err := cmd.Flags().Parse(args); err != nil {
		t.Fatal(err)
	}
	flags, err := readInitFlags(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if flags.Stack != "go" || flags.Port != 8080 || !flags.DryRun || !flags.Force {
		t.Fatalf("flags: %+v", flags)
	}
}

func TestMapExitCode_AllCases(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, 0},
		{"conflict", errConflict, 3},
		{"validation", errValidation, 4},
		{"ambiguous", errAmbiguousStack, 5},
		{"exitError2", newExitError(2, errors.New("x")), 2},
		{"exitError1", newExitError(1, errors.New("x")), 1},
		{"generic", errors.New("boom"), 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := MapExitCode(c.err); got != c.want {
				t.Fatalf("MapExitCode(%v)=%d want %d", c.err, got, c.want)
			}
		})
	}
}

func TestExitError(t *testing.T) {
	ee := newExitError(3, errors.New("conflict"))
	if ee.Error() != "conflict" || ee.ExitCode() != 3 {
		t.Fatalf("exitError: %v %d", ee.Error(), ee.ExitCode())
	}
}

func TestNewRootCommand(t *testing.T) {
	root := NewRootCommand()
	if root == nil {
		t.Fatal("NewRootCommand returned nil")
	}

	// Test executing "version" subcommand through root
	buf := new(strings.Builder)
	root.SetOut(buf)
	root.SetArgs([]string{"version"})
	if err := root.Execute(); err != nil {
		t.Fatalf("root execute version: %v", err)
	}
	if !strings.Contains(buf.String(), "devopsgen") {
		t.Fatalf("expected version output, got %q", buf.String())
	}

	// Test executing "list" subcommand through root
	buf.Reset()
	root.SetOut(buf)
	root.SetArgs([]string{"list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("root execute list: %v", err)
	}
	if !strings.Contains(buf.String(), "go") {
		t.Fatalf("expected list output, got %q", buf.String())
	}
}

func TestExitVersion(t *testing.T) {
	// Call ExitVersion to exercise stdout printing helper
	ExitVersion()
}

func TestIsInteractive(t *testing.T) {
	// Exercise isInteractive helper
	_ = isInteractive()
}
