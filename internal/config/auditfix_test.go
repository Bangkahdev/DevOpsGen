package config

import "testing"

func TestValidateConfig_RejectsGitHubActionsExpressions(t *testing.T) {
	for _, field := range []func(*Config) string{
		func(c *Config) string { c.BuildCmd = "${{ secrets.CMD }} go build"; return "build-cmd" },
		func(c *Config) string { c.TestCmd = "go test ${{ github.sha }}"; return "test-cmd" },
		func(c *Config) string { c.StartCmd = "${{ secrets.CMD }} & ./server"; return "start-cmd" },
	} {
		cfg := Config{
			Stack:    StackGo,
			CI:       CIGitHubActions,
			Registry: RegistryDockerHub,
			Image:    "bangkah/my-app",
			Port:     8080,
		}
		what := field(&cfg)
		if err := ValidateConfig(cfg); err == nil {
			t.Fatalf("%s with GitHub Actions expression must be rejected", what)
		}
	}
}

func TestValidateConfig_HealthPathTraversalRejected(t *testing.T) {
	cfg := Config{
		Stack:    StackGo,
		CI:       CIGitHubActions,
		Registry: RegistryDockerHub,
		Image:    "bangkah/my-app",
		Port:     8080,
		K8s:      K8sOptions{Enabled: true, HealthPath: "/../admin"},
	}
	if err := ValidateConfig(cfg); err == nil {
		t.Fatal("health path containing .. must be rejected")
	}
}

func TestValidateConfig_TagLeadingHyphenRejected(t *testing.T) {
	for _, tag := range []string{"-rc", "...", "-v1"} {
		cfg := Config{
			Stack:    StackGo,
			CI:       CIGitHubActions,
			Registry: RegistryDockerHub,
			Image:    "bangkah/my-app",
			Port:     8080,
			Tag:      tag,
		}
		if err := ValidateConfig(cfg); err == nil {
			t.Fatalf("tag %q must be rejected", tag)
		}
	}
}

func TestValidateConfig_IngressRequiresK8s(t *testing.T) {
	cfg := Config{
		Stack:    StackGo,
		CI:       CIGitHubActions,
		Registry: RegistryDockerHub,
		Image:    "bangkah/my-app",
		Port:     8080,
		K8s:      K8sOptions{Ingress: true, IngressHost: "app.example.com"},
	}
	if err := ValidateConfig(cfg); err == nil {
		t.Fatal("ingress without k8s generation must be rejected")
	}

	cfg.K8s.Enabled = true
	if err := ValidateConfig(cfg); err != nil {
		t.Fatalf("ingress with k8s generation should be valid: %v", err)
	}
}

func TestValidateConfig_IngressHostRequired(t *testing.T) {
	cfg := Config{
		Stack:    StackGo,
		CI:       CIGitHubActions,
		Registry: RegistryDockerHub,
		Image:    "bangkah/my-app",
		Port:     8080,
		K8s:      K8sOptions{Enabled: true, Ingress: true},
	}
	if err := ValidateConfig(cfg); err == nil {
		t.Fatal("ingress without ingress host must be rejected")
	}
}
