package wizard

import (
	"strings"
	"testing"

	"devopsgen.dev/devopsgen/internal/config"
)

func TestRun_NonInteractive(t *testing.T) {
	// Force the interactivity check to fail so Run refuses to launch
	// the wizard rather than block on input.
	orig := interactiveFn
	defer func() { interactiveFn = orig }()
	interactiveFn = func() bool { return false }
	_, err := Run()
	if err == nil {
		t.Fatal("expected error in non-interactive environment")
	}
}

func TestRun_InteractiveNoIndicators(t *testing.T) {
	// In an interactive terminal with no project indicators, Run still
	// returns a usable (empty-stack) config rather than erroring.
	orig := interactiveFn
	defer func() { interactiveFn = orig }()
	interactiveFn = func() bool { return true }
	cfg, err := Run()
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if cfg.CI != config.CIGitHubActions {
		t.Fatalf("CI: %s", cfg.CI)
	}
}

func TestIsInteractive(t *testing.T) {
	// The value is environment-dependent; just ensure it returns a bool.
	_ = isInteractive()
}

func TestNewModel(t *testing.T) {
	m := New()
	if m.step != StepDetect {
		t.Fatalf("step: %v", m.step)
	}
	if len(m.choices) != 3 {
		t.Fatalf("choices: %v", m.choices)
	}
	if m.input.Placeholder != "port" {
		t.Fatalf("placeholder: %q", m.input.Placeholder)
	}
}

func TestViewDetect(t *testing.T) {
	m := New()
	m.step = StepDetect
	v := m.view()
	if v == "" {
		t.Fatal("view should not be empty")
	}
}

func TestViewSummary(t *testing.T) {
	m := New()
	m.step = StepSummary
	m.cfg = config.Config{
		Stack:    config.StackGo,
		AppType:  config.AppTypeBinary,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryDockerHub,
		Image:    "bangkah/my-app",
		Port:     8080,
		BuildCmd: "go build",
		TestCmd:  "go test",
		StartCmd: "./server",
		Tag:      "v1",
		K8s:      config.K8sOptions{Enabled: true},
	}
	v := m.view()
	for _, want := range []string{"stack:", "go", "app-type:", "binary", "image:", "bangkah/my-app", "port:", "8080"} {
		if !strings.Contains(v, want) {
			t.Fatalf("summary view missing %q in:\n%s", want, v)
		}
	}
}

func TestViewPort(t *testing.T) {
	m := New()
	m.step = StepPort
	v := m.view()
	if !strings.Contains(v, "port") {
		t.Fatalf("port view should mention port:\n%s", v)
	}
}

func TestViewStack(t *testing.T) {
	m := New()
	m.step = StepStack
	m.cfg.Stack = config.StackGo
	v := m.view()
	if !strings.Contains(v, "go") || !strings.Contains(v, "nodejs") || !strings.Contains(v, "python") {
		t.Fatalf("stack view should list stacks:\n%s", v)
	}
}
