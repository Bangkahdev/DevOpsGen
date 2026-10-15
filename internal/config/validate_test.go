package config

import (
	"testing"
)

func TestValidateConfig_RequiredFields(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
	}{
		{"empty", Config{}},
		{"bad stack", Config{Stack: "python"}},
		{"bad ci", Config{Stack: StackGo, CI: "jenkins"}},
		{"bad registry", Config{Stack: StackGo, CI: CIGitHubActions, Registry: "registry.gitlab"}},
		{"empty image", Config{Stack: StackGo, CI: CIGitHubActions, Registry: RegistryDockerHub, Port: 8080}},
		{"bad image host", Config{Stack: StackGo, CI: CIGitHubActions, Registry: RegistryDockerHub, Image: "registry.example.com/a/b", Port: 8080}},
		{"bad image tag", Config{Stack: StackGo, CI: CIGitHubActions, Registry: RegistryDockerHub, Image: "a/b:latest", Port: 8080}},
		{"port too low", Config{Stack: StackGo, CI: CIGitHubActions, Registry: RegistryDockerHub, Image: "a/b", Port: 0}},
		{"port too high", Config{Stack: StackGo, CI: CIGitHubActions, Registry: RegistryDockerHub, Image: "a/b", Port: 70000}},
		{"bad app type for go", Config{Stack: StackGo, CI: CIGitHubActions, Registry: RegistryDockerHub, Image: "a/b", Port: 8080, AppType: AppTypeExpress}},
		{"bad app type for node", Config{Stack: StackNode, CI: CIGitHubActions, Registry: RegistryDockerHub, Image: "a/b", Port: 8080, AppType: AppTypeBinary}},
		{"ingress host missing", Config{Stack: StackGo, CI: CIGitHubActions, Registry: RegistryDockerHub, Image: "a/b", Port: 8080, K8s: K8sOptions{Ingress: true}}},
		{"bad ingress host", Config{Stack: StackGo, CI: CIGitHubActions, Registry: RegistryDockerHub, Image: "a/b", Port: 8080, K8s: K8sOptions{Ingress: true, IngressHost: "Bad_Host!"}}},
		{"bad health path", Config{Stack: StackGo, CI: CIGitHubActions, Registry: RegistryDockerHub, Image: "a/b", Port: 8080, K8s: K8sOptions{HealthPath: "health"}}},
		{"command with newline", Config{Stack: StackGo, CI: CIGitHubActions, Registry: RegistryDockerHub, Image: "a/b", Port: 8080, BuildCmd: "make\nbuild"}},
		{"command with control char", Config{Stack: StackGo, CI: CIGitHubActions, Registry: RegistryDockerHub, Image: "a/b", Port: 8080, StartCmd: "run\x01"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := ValidateConfig(c.cfg); err == nil {
				t.Fatalf("expected validation error for %+v", c.cfg)
			}
		})
	}
}

func TestValidateConfig_Valid(t *testing.T) {
	cases := []Config{
		{Stack: StackGo, CI: CIGitHubActions, Registry: RegistryDockerHub, Image: "bangkah/my-app", Port: 8080},
		{Stack: StackGo, CI: CIGitHubActions, Registry: RegistryDockerHub, Image: "bangkah/my-app", Port: 8080, AppType: AppTypeBinary},
		{Stack: StackNode, CI: CIGitHubActions, Registry: RegistryGHCR, Image: "bangkah/api", Port: 3000, AppType: AppTypeExpress, StartCmd: "npm start"},
		{Stack: StackNode, CI: CIGitHubActions, Registry: RegistryDockerHub, Image: "bangkah/site", Port: 8080, AppType: AppTypeVite, K8s: K8sOptions{Ingress: true, IngressHost: "example.com", HealthPath: "/health"}},
	}
	for i, cfg := range cases {
		t.Run("ok", func(t *testing.T) {
			if err := ValidateConfig(cfg); err != nil {
				t.Fatalf("case %d: %v", i, err)
			}
		})
	}
}

func TestValidateOutputDir(t *testing.T) {
	good := []string{"out", "deploy/k8s", "."}
	for _, d := range good {
		if err := ValidateOutputDir(d); err != nil {
			t.Fatalf("ValidateOutputDir(%q) err: %v", d, err)
		}
	}
	bad := []string{"", "/etc", "../escape", "out/../../etc", "/abs/path"}
	for _, d := range bad {
		if err := ValidateOutputDir(d); err == nil {
			t.Fatalf("expected error for %q", d)
		}
	}
}

func TestImageRef(t *testing.T) {
	if c := (Config{Registry: RegistryDockerHub, Image: "a/b"}).ImageRef(); c != "docker.io/a/b" {
		t.Fatalf("dockerhub ref: %s", c)
	}
	if c := (Config{Registry: RegistryGHCR, Image: "a/b"}).ImageRef(); c != "ghcr.io/a/b" {
		t.Fatalf("ghcr ref: %s", c)
	}
}

func TestTagList(t *testing.T) {
	tags := (Config{Tag: "v1"}).TagList()
	if len(tags) != 2 || tags[0] != "v1" || tags[1] != "${{ github.sha }}" {
		t.Fatalf("tag list: %v", tags)
	}
	tags = (Config{}).TagList()
	if len(tags) != 2 || tags[0] != "latest" || tags[1] != "${{ github.sha }}" {
		t.Fatalf("default tag list: %v", tags)
	}
}