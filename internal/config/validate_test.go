package config

import (
	"strings"
	"testing"
)

func TestValidateConfig_RequiredFields(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
	}{
		{"empty", Config{}},
		{"bad stack", Config{Stack: "ruby"}},
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
		{"tag too long", Config{Stack: StackGo, CI: CIGitHubActions, Registry: RegistryDockerHub, Image: "a/b", Port: 8080, Tag: strings.Repeat("a", 129)}},
		{"tag with newline", Config{Stack: StackGo, CI: CIGitHubActions, Registry: RegistryDockerHub, Image: "a/b", Port: 8080, Tag: "v1\nrm -rf /"}},
		{"tag with shell metachar", Config{Stack: StackGo, CI: CIGitHubActions, Registry: RegistryDockerHub, Image: "a/b", Port: 8080, Tag: "v1; rm -rf /"}},
		{"tag with space", Config{Stack: StackGo, CI: CIGitHubActions, Registry: RegistryDockerHub, Image: "a/b", Port: 8080, Tag: "v1 tag"}},
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
		{Stack: StackNode, CI: CIGitHubActions, Registry: RegistryDockerHub, Image: "a/b", Port: 8080, Tag: "v1.2.3"},
		{Stack: StackNode, CI: CIGitHubActions, Registry: RegistryDockerHub, Image: "a/b", Port: 8080, Tag: "latest"},
	}
	for i, cfg := range cases {
		t.Run("ok", func(t *testing.T) {
			if err := ValidateConfig(cfg); err != nil {
				t.Fatalf("case %d: %v", i, err)
			}
		})
	}
}

func TestValidateTag(t *testing.T) {
	good := []string{"", "latest", "v1", "v1.2.3", "my-tag", "my_tag", "my.tag", "sha-abc123"}
	for _, tag := range good {
		if err := ValidateTag(tag); err != nil {
			t.Fatalf("ValidateTag(%q) unexpected error: %v", tag, err)
		}
	}
	bad := []string{"v1\n", "v1;rm", "v1 tag", "v1$()", "v1`x`", "v1|x", "v1&x", "v1<x", "v1>x", "v1!x", "v1#x", "v1%x", "v1@x", "v1:x", "v1\\x", strings.Repeat("a", 129)}
	for _, tag := range bad {
		if err := ValidateTag(tag); err == nil {
			t.Fatalf("expected error for tag %q", tag)
		}
	}
	if len(bad[0]) > 0 {
		_ = bad[0]
	}
}

func TestValidateOutputDir(t *testing.T) {
	good := []string{"out", "deploy/k8s", "."}
	for _, d := range good {
		if err := ValidateOutputDir(d); err != nil {
			t.Fatalf("ValidateOutputDir(%q) err: %v", d, err)
		}
	}
	bad := []string{"", "../escape", "out/../../etc", ".."}
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

func TestRequiredFieldsMissing(t *testing.T) {
	cfg := Config{}
	missing := RequiredFieldsMissing(cfg)
	if len(missing) == 0 {
		t.Fatal("expected missing fields for empty config")
	}
	foundStack := false
	foundImage := false
	foundPort := false
	for _, m := range missing {
		if m == "--stack" {
			foundStack = true
		}
		if m == "--image" {
			foundImage = true
		}
		if m == "--port" {
			foundPort = true
		}
	}
	if !foundStack || !foundImage || !foundPort {
		t.Fatalf("missing fields incomplete: %v", missing)
	}

	// Node.js without start-cmd
	cfgNode := Config{Stack: StackNode, Image: "a/b", Port: 8080}
	missing = RequiredFieldsMissing(cfgNode)
	foundStart := false
	for _, m := range missing {
		if m == "--start-cmd" {
			foundStart = true
		}
	}
	if !foundStart {
		t.Fatalf("expected --start-cmd missing for nodejs, got %v", missing)
	}
}
