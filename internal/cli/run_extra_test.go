package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devopsgen.dev/devopsgen/internal/config"
	"devopsgen.dev/devopsgen/internal/detector"
)

func TestRunInit_StackFromDetection(t *testing.T) {
	// Create a temp dir with go.mod so detection finds Go.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	// Non-interactive: detection should work and stack should be resolved.
	flags := &InitFlags{
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "a/b",
		Port:      8080,
		OutputDir: t.TempDir(),
	}
	err := RunInit(flags)
	if err != nil {
		t.Fatalf("expected success with detected stack, got: %v", err)
	}
}

func TestRunInit_StackDetectionFailsNonInteractive(t *testing.T) {
	// Empty temp dir with no indicators.
	flags := &InitFlags{
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "a/b",
		Port:      8080,
		OutputDir: t.TempDir(),
	}
	err := RunInit(flags)
	if err == nil {
		t.Fatal("expected error when stack cannot be detected")
	}
	var ee *exitError
	if !errors.As(err, &ee) || ee.ExitCode() != 2 {
		t.Fatalf("expected exit 2, got %v", err)
	}
}

func TestRunInit_AmbiguousDetectionExit5(t *testing.T) {
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
	flags := &InitFlags{
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "a/b",
		Port:      8080,
		OutputDir: t.TempDir(),
	}
	err := RunInit(flags)
	if err == nil {
		t.Fatal("expected ambiguity error")
	}
	var ee *exitError
	if !errors.As(err, &ee) || ee.ExitCode() != 5 {
		t.Fatalf("expected exit 5, got %v", err)
	}
}

func TestRunInit_ValidateTag(t *testing.T) {
	flags := &InitFlags{
		Stack:     "go",
		AppType:   "binary",
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "a/b",
		Port:      8080,
		Tag:       "v1;rm -rf /",
		OutputDir: t.TempDir(),
	}
	err := RunInit(flags)
	if err == nil {
		t.Fatal("expected error for invalid tag")
	}
}

func TestRunInit_ValidTag(t *testing.T) {
	dir := t.TempDir()
	flags := &InitFlags{
		Stack:     "go",
		AppType:   "binary",
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "a/b",
		Port:      8080,
		Tag:       "v1.2.3",
		OutputDir: dir,
	}
	if err := RunInit(flags); err != nil {
		t.Fatalf("expected success with valid tag, got: %v", err)
	}
}

func TestResolveConfig_DetectStack(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		CI:       config.CIGitHubActions,
		Registry: config.RegistryDockerHub,
		Image:    "a/b",
		Port:     8080,
	}
	d := detector.New()
	result, err := ResolveConfig(cfg, d)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if result.Stack != config.StackGo {
		t.Fatalf("expected stack go, got %s", result.Stack)
	}
}

func TestResolveConfig_AmbiguousExit(t *testing.T) {
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
	cfg := config.Config{
		CI:       config.CIGitHubActions,
		Registry: config.RegistryDockerHub,
		Image:    "a/b",
		Port:     8080,
	}
	d := detector.New()
	_, err := ResolveConfig(cfg, d)
	if err == nil {
		t.Fatal("expected ambiguity error")
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("expected ambiguous error, got: %v", err)
	}
}
