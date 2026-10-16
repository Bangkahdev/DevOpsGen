package cli

import (
	"os"
	"path/filepath"
	"testing"

	"devopsgen.dev/devopsgen/internal/config"
)

func TestRunInit_OutputDirInvalid(t *testing.T) {
	flags := &InitFlags{
		Stack:     "go",
		AppType:   "binary",
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "a/b",
		Port:      8080,
		OutputDir: "../escape",
	}
	err := RunInit(flags)
	if err == nil {
		t.Fatal("expected error for invalid output dir")
	}
}

func TestRunInit_TagInvalid(t *testing.T) {
	dir := t.TempDir()
	flags := &InitFlags{
		Stack:     "go",
		AppType:   "binary",
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "a/b",
		Port:      8080,
		Tag:       "v1;rm",
		OutputDir: dir,
	}
	err := RunInit(flags)
	if err == nil {
		t.Fatal("expected error for invalid tag")
	}
}

func TestRunInit_GoBuildCmdUsed(t *testing.T) {
	dir := t.TempDir()
	flags := &InitFlags{
		Stack:     "go",
		AppType:   "binary",
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "a/b",
		Port:      8080,
		BuildCmd:  "go build -o custom .",
		OutputDir: dir,
	}
	if err := RunInit(flags); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	df, err := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if err != nil {
		t.Fatal(err)
	}
	if !contains(string(df), "go build -o custom") {
		t.Fatalf("Dockerfile should use custom build cmd:\n%s", string(df))
	}
}

func TestRunInit_NodeStartCmdDefault(t *testing.T) {
	dir := t.TempDir()
	flags := &InitFlags{
		Stack:     "nodejs",
		AppType:   "express",
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "a/b",
		Port:      3000,
		OutputDir: dir,
	}
	if err := RunInit(flags); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	df, err := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if err != nil {
		t.Fatal(err)
	}
	if !contains(string(df), "node server.js") {
		t.Fatalf("Dockerfile should default to node server.js:\n%s", string(df))
	}
}

func TestRunInit_ViteStartCmdDefault(t *testing.T) {
	dir := t.TempDir()
	flags := &InitFlags{
		Stack:     "nodejs",
		AppType:   "vite",
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "a/b",
		Port:      8080,
		OutputDir: dir,
	}
	if err := RunInit(flags); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	df, err := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if err != nil {
		t.Fatal(err)
	}
	if !contains(string(df), "npx serve -s dist -l 8080") {
		t.Fatalf("Dockerfile should use npx serve:\n%s", string(df))
	}
}

func TestRunInit_NextJSDefaults(t *testing.T) {
	dir := t.TempDir()
	flags := &InitFlags{
		Stack:     "nodejs",
		AppType:   "nextjs",
		CI:        "github-actions",
		Registry:  "ghcr",
		Image:     "a/b",
		Port:      3000,
		OutputDir: dir,
	}
	if err := RunInit(flags); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	df, err := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if err != nil {
		t.Fatal(err)
	}
	if !contains(string(df), "npm run build") || !contains(string(df), "npx next start -p 3000") {
		t.Fatalf("Dockerfile should use Next.js defaults:\n%s", string(df))
	}
}

func TestValidateConfig_StartCmdForGo(t *testing.T) {
	// Go ignores start-cmd
	cfg := config.Config{
		Stack:    config.StackGo,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryDockerHub,
		Image:    "a/b",
		Port:     8080,
		StartCmd: "echo hello",
	}
	if err := config.ValidateConfig(cfg); err != nil {
		t.Fatalf("Go should accept start-cmd: %v", err)
	}
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
