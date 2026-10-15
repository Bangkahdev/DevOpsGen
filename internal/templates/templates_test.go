package templates

import (
	"strings"
	"testing"
)

func TestRead_Dockerfile(t *testing.T) {
	b, err := Read("docker/Dockerfile.go.tmpl")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(b), "EXPOSE") {
		t.Fatal("Dockerfile.go.tmpl missing EXPOSE")
	}
}

func TestRead_AllTemplates(t *testing.T) {
	for _, f := range []string{
		"docker/Dockerfile.go.tmpl",
		"docker/Dockerfile.express.tmpl",
		"docker/Dockerfile.nextjs.tmpl",
		"docker/Dockerfile.vite.tmpl",
		"docker/Dockerfile.nestjs.tmpl",
		"docker/Dockerfile.generic.tmpl",
		"github-actions/ci-workflow.yaml.tmpl",
		"kubernetes/deployment.yaml.tmpl",
	} {
		b, err := Read(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		if len(b) == 0 {
			t.Fatalf("empty template: %s", f)
		}
	}
}

func TestRead_TraversalRejected(t *testing.T) {
	bad := []string{
		"../embed.go",
		"../../go.mod",
		"/etc/passwd",
		"..",
		".",
		"docker/../..",
	}
	for _, p := range bad {
		if _, err := Read(p); err == nil {
			t.Fatalf("expected traversal error for %q", p)
		}
	}
}

func TestExists(t *testing.T) {
	if !Exists("docker/Dockerfile.go.tmpl") {
		t.Fatal("Exists should find Dockerfile.go.tmpl")
	}
	if Exists("does/not/exist.tmpl") {
		t.Fatal("Exists should not find missing template")
	}
}

func TestList(t *testing.T) {
	paths, err := List("docker")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(paths) < 6 {
		t.Fatalf("expected at least 6 docker templates, got %d", len(paths))
	}
}
