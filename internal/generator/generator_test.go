package generator

import (
	"bytes"
	"strings"
	"testing"

	"devopsgen.dev/devopsgen/internal/config"
)

func TestGenerate_Go(t *testing.T) {
	cfg := config.Config{
		Stack:    config.StackGo,
		AppType:  config.AppTypeBinary,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryGHCR,
		Image:    "bangkah/my-app",
		Port:     8080,
		BuildCmd: "go build -o server .",
		TestCmd:  "go test ./...",
		StartCmd: "./server",
		Tag:      "v1",
	}
	g := New()
	arts, err := g.Generate(cfg)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(arts) < 2 {
		t.Fatalf("expected at least Dockerfile + workflow, got %d", len(arts))
	}
	var dockerfile []byte
	for _, a := range arts {
		if a.Path == "Dockerfile" {
			dockerfile = a.Content
		}
	}
	if dockerfile == nil {
		t.Fatal("Dockerfile not generated")
	}
	if !bytes.Contains(dockerfile, []byte("EXPOSE 8080")) {
		t.Fatal("Dockerfile missing EXPOSE 8080")
	}
	if !bytes.Contains(dockerfile, []byte("go build")) {
		t.Fatal("Dockerfile missing build command")
	}
}

func TestGenerate_NodeExpress(t *testing.T) {
	cfg := config.Config{
		Stack:    config.StackNode,
		AppType:  config.AppTypeExpress,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryDockerHub,
		Image:    "bangkah/api",
		Port:     3000,
		StartCmd: "npm start",
	}
	arts, err := New().Generate(cfg)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	var dockerfile []byte
	for _, a := range arts {
		if a.Path == "Dockerfile" {
			dockerfile = a.Content
		}
	}
	if dockerfile == nil {
		t.Fatal("Dockerfile not generated")
	}
	if !bytes.Contains(dockerfile, []byte("npm start")) {
		t.Fatal("Dockerfile missing npm start")
	}
}

func TestGenerate_TestCmdSkipped(t *testing.T) {
	cfg := config.Config{
		Stack:    config.StackGo,
		AppType:  config.AppTypeBinary,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryDockerHub,
		Image:    "bangkah/app",
		Port:     8080,
	}
	arts, err := New().Generate(cfg)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	var wf []byte
	for _, a := range arts {
		if strings.HasSuffix(a.Path, "ci.yml") {
			wf = a.Content
		}
	}
	if wf == nil {
		t.Fatal("workflow not generated")
	}
	if !bytes.Contains(wf, []byte("Test stage is not configured")) {
		t.Fatal("workflow should state test stage is skipped")
	}
	if bytes.Contains(wf, []byte("- name: Test\n        run:")) {
		t.Fatal("workflow should not have a real test step")
	}
}

func TestGenerate_Deterministic(t *testing.T) {
	cfg := config.Config{
		Stack:    config.StackGo,
		AppType:  config.AppTypeBinary,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryGHCR,
		Image:    "bangkah/my-app",
		Port:     8080,
		BuildCmd: "go build",
		TestCmd:  "go test",
	}
	first, err := New().Generate(cfg)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := New().Generate(cfg)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if len(first) != len(second) {
		t.Fatalf("artifact count differs: %d vs %d", len(first), len(second))
	}
	for i := range first {
		if !bytes.Equal(first[i].Content, second[i].Content) {
			t.Fatalf("artifact %d content differs", i)
		}
		if first[i].Path != second[i].Path {
			t.Fatalf("artifact %d path differs: %q vs %q", i, first[i].Path, second[i].Path)
		}
	}
}

func TestGenerate_NoK8s(t *testing.T) {
	cfg := config.Config{
		Stack:    config.StackGo,
		AppType:  config.AppTypeBinary,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryDockerHub,
		Image:    "bangkah/app",
		Port:     8080,
		K8s:      config.K8sOptions{Enabled: false},
	}
	arts, err := New().Generate(cfg)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	for _, a := range arts {
		if strings.HasPrefix(a.Path, "kubernetes/") {
			t.Fatalf("k8s artifact generated when disabled: %s", a.Path)
		}
	}
}

func TestGenerate_K8sWithIngress(t *testing.T) {
	cfg := config.Config{
		Stack:    config.StackGo,
		AppType:  config.AppTypeBinary,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryDockerHub,
		Image:    "bangkah/app",
		Port:     8080,
		K8s: config.K8sOptions{
			Enabled:     true,
			Ingress:     true,
			IngressHost: "app.example.com",
			HealthPath:  "/health",
		},
	}
	arts, err := New().Generate(cfg)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	var dep []byte
	for _, a := range arts {
		if strings.HasSuffix(a.Path, "deployment.yaml") {
			dep = a.Content
		}
	}
	if dep == nil {
		t.Fatal("deployment not generated")
	}
	if !bytes.Contains(dep, []byte("app.example.com")) {
		t.Fatal("deployment missing ingress host")
	}
	if !bytes.Contains(dep, []byte("/health")) {
		t.Fatal("deployment missing health probe")
	}
}

func TestTemplateData_Helpers(t *testing.T) {
	cfg := config.Config{
		Stack:    config.StackGo,
		AppType:  config.AppTypeBinary,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryGHCR,
		Image:    "org/my-service",
		Port:     8080,
		Tag:      "v2",
		K8s: config.K8sOptions{
			Enabled:     true,
			Ingress:     true,
			IngressHost: "svc.example.com",
			HealthPath:  "/live",
		},
	}
	td := templateData{Config: cfg}
	if td.Name() != "my-service" {
		t.Fatalf("expected my-service, got %q", td.Name())
	}
	if td.ImageRef() != "ghcr.io/org/my-service" {
		t.Fatalf("expected image ref ghcr.io/org/my-service, got %q", td.ImageRef())
	}
	if td.Tag() != "v2" {
		t.Fatalf("expected v2, got %q", td.Tag())
	}
	tdDefaultTag := templateData{Config: config.Config{}}
	if tdDefaultTag.Tag() != "latest" {
		t.Fatalf("expected latest, got %q", tdDefaultTag.Tag())
	}
	if td.IngressHost() != "svc.example.com" {
		t.Fatalf("expected ingress host, got %q", td.IngressHost())
	}
	if td.HealthPath() != "/live" {
		t.Fatalf("expected /live, got %q", td.HealthPath())
	}
	if !td.Ingress() {
		t.Fatal("expected Ingress true")
	}
	if !td.K8sEnabled() {
		t.Fatal("expected K8sEnabled true")
	}

	// Test lastSegment helper
	if lastSegment("simple") != "simple" {
		t.Fatalf("expected simple, got %q", lastSegment("simple"))
	}
	if lastSegment("a/b/c/") != "c" {
		t.Fatalf("expected c, got %q", lastSegment("a/b/c/"))
	}
}
