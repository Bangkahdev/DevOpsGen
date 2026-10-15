package generator

import (
	"testing"

	"devopsgen.dev/devopsgen/internal/config"
)

// BenchmarkGenerate measures core engine rendering under the NFR-01
// reference conditions: <=20 artifacts, <=256 KiB total output, no
// network access, no external processes.
func BenchmarkGenerate(b *testing.B) {
	cfg := config.Config{
		Stack:    config.StackGo,
		AppType:  config.AppTypeBinary,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryDockerHub,
		Image:    "bangkah/my-app",
		Port:     8080,
		BuildCmd: "go build -o server .",
		TestCmd:  "go test ./...",
		Tag:      "v1",
		K8s:      config.K8sOptions{Enabled: true, Ingress: true, IngressHost: "app.example.com", HealthPath: "/health"},
	}
	g := New()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := g.Generate(cfg); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerate_Node measures rendering for a Node.js Express stack.
func BenchmarkGenerate_Node(b *testing.B) {
	cfg := config.Config{
		Stack:    config.StackNode,
		AppType:  config.AppTypeExpress,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryGHCR,
		Image:    "bangkah/api",
		Port:     3000,
		StartCmd: "npm start",
	}
	g := New()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := g.Generate(cfg); err != nil {
			b.Fatal(err)
		}
	}
}

// TestGenerate_NFROneBounds asserts the NFR-01 reference conditions:
// <=20 artifacts, <=256 KiB total output.
func TestGenerate_NFROneBounds(t *testing.T) {
	cfg := config.Config{
		Stack:    config.StackGo,
		AppType:  config.AppTypeBinary,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryDockerHub,
		Image:    "bangkah/my-app",
		Port:     8080,
		BuildCmd: "go build -o server .",
		TestCmd:  "go test ./...",
		Tag:      "v1",
		K8s:      config.K8sOptions{Enabled: true, Ingress: true, IngressHost: "app.example.com", HealthPath: "/health"},
	}
	arts, err := New().Generate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(arts) > 20 {
		t.Fatalf("artifact count %d exceeds NFR-01 bound of 20", len(arts))
	}
	var total int
	for _, a := range arts {
		total += len(a.Content)
	}
	if total > 256*1024 {
		t.Fatalf("total output %d bytes exceeds NFR-01 bound of 256 KiB", total)
	}
}
