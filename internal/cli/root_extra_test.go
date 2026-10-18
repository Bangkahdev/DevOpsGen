package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devopsgen.dev/devopsgen/internal/output"
)

func TestRunValidate_Recursive(t *testing.T) {
	dir := t.TempDir()
	// Create artifacts in subdirectories.
	if err := os.MkdirAll(filepath.Join(dir, ".github/workflows"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "kubernetes"), 0o755); err != nil {
		t.Fatal(err)
	}
	arts := []struct {
		path    string
		content string
	}{
		{"Dockerfile", "FROM golang:1.22\nEXPOSE 8080"},
		{".github/workflows/ci.yml", "name: CI\non:\n  push:\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n      - run: docker push x:${{ github.sha }}"},
		{"kubernetes/deployment.yaml", "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: app\nspec:\n  selector:\n    matchLabels:\n      app: app\n  template:\n    metadata:\n      labels:\n        app: app\n    spec:\n      containers:\n        - name: app\n          image: ghcr.io/x/app:latest"},
		{"kubernetes/service.yaml", "apiVersion: v1\nkind: Service\nmetadata:\n  name: app\nspec:\n  selector:\n    app: app\n  ports:\n    - port: 8080\n      targetPort: 8080"},
	}
	for _, a := range arts {
		fullPath := filepath.Join(dir, a.path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(a.content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := runValidate(dir, os.Stdout); err != nil {
		t.Fatalf("runValidate recursive: %v", err)
	}
}

func TestRunValidate_MissingRequiredArtifact(t *testing.T) {
	dir := t.TempDir()
	// Only create a Dockerfile, missing workflow.
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM golang:1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := runValidate(dir, os.Stdout)
	if err == nil {
		t.Fatal("expected error for missing required artifact")
	}
	var ee *exitError
	if !errors.As(err, &ee) || ee.ExitCode() != 4 {
		t.Fatalf("expected exit 4, got %v", err)
	}
}

func TestCollectArtifacts_Recursive(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "kubernetes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM scratch"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "kubernetes", "deployment.yaml"), []byte("kind: Deployment"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	arts, err := collectArtifacts(dir)
	if err != nil {
		t.Fatalf("collectArtifacts: %v", err)
	}
	if len(arts) != 2 {
		t.Fatalf("expected 2 artifacts, got %d", len(arts))
	}
	paths := map[string]bool{}
	for _, a := range arts {
		paths[a.Path] = true
	}
	if !paths["Dockerfile"] {
		t.Fatal("expected Dockerfile artifact")
	}
	if !paths["kubernetes/deployment.yaml"] {
		t.Fatal("expected kubernetes/deployment.yaml artifact")
	}
}

func TestCollectArtifacts_SkipsHiddenDirs(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".git", "config"), []byte("[core]"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM scratch"), 0o644); err != nil {
		t.Fatal(err)
	}
	arts, err := collectArtifacts(dir)
	if err != nil {
		t.Fatalf("collectArtifacts: %v", err)
	}
	for _, a := range arts {
		if strings.Contains(a.Path, ".git") {
			t.Fatal("should not include .git directory artifacts")
		}
	}
}

func TestMapCommandError_ExitError(t *testing.T) {
	if got := MapExitCode(newExitError(4, errors.New("validate"))); got != 4 {
		t.Fatalf("expected 4, got %d", got)
	}
}

var _ = output.Artifact{}
