package wizard

import (
	"errors"
	"strings"
	"testing"

	"devopsgen.dev/devopsgen/internal/config"
	"devopsgen.dev/devopsgen/internal/detector"
)

func TestRun_NonInteractive(t *testing.T) {
	orig := interactiveFn
	defer func() { interactiveFn = orig }()
	interactiveFn = func() bool { return false }
	_, err := Run()
	if err == nil {
		t.Fatal("expected error in non-interactive environment")
	}
	if !strings.Contains(err.Error(), "not interactive") {
		t.Fatalf("expected not-interactive error, got: %v", err)
	}
}

func TestIsInteractive(t *testing.T) {
	_ = isInteractive()
}

func TestNewModel(t *testing.T) {
	m := New()
	if m.step != StepDetect {
		t.Fatalf("step: %v", m.step)
	}
	if len(m.choices) != 2 {
		t.Fatalf("choices: %v", m.choices)
	}
	if m.input.Placeholder != "port" {
		t.Fatalf("placeholder: %q", m.input.Placeholder)
	}
	if m.cfg.CI != config.CIGitHubActions {
		t.Fatalf("CI: %s", m.cfg.CI)
	}
	if m.cfg.Registry != config.RegistryDockerHub {
		t.Fatalf("Registry: %s", m.cfg.Registry)
	}
	if m.cfg.Tag != config.DefaultTag {
		t.Fatalf("Tag: %s", m.cfg.Tag)
	}
	if !m.cfg.K8s.Enabled {
		t.Fatal("K8s should be enabled by default")
	}
}

func TestViewDetect(t *testing.T) {
	m := New()
	m.step = StepDetect
	m.detected = true
	m.detect.Stack = "go"
	m.detect.Indicators = []string{"go.mod"}
	v := m.View()
	if v == "" {
		t.Fatal("view should not be empty")
	}
	if !strings.Contains(v, "go") {
		t.Fatalf("view should mention stack: %s", v)
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
	v := m.View()
	for _, want := range []string{"stack:", "go", "app-type:", "binary", "image:", "bangkah/my-app", "port:", "8080"} {
		if !strings.Contains(v, want) {
			t.Fatalf("summary view missing %q in:\n%s", want, v)
		}
	}
}

func TestViewPort(t *testing.T) {
	m := New()
	m.step = StepPort
	v := m.View()
	if !strings.Contains(v, "port") {
		t.Fatalf("port view should mention port:\n%s", v)
	}
}

func TestViewStack(t *testing.T) {
	m := New()
	m.step = StepStack
	m.cfg.Stack = config.StackGo
	v := m.View()
	if !strings.Contains(v, "go") || !strings.Contains(v, "nodejs") {
		t.Fatalf("stack view should list stacks:\n%s", v)
	}
}

func TestViewAppType(t *testing.T) {
	m := New()
	m.step = StepAppType
	m.cfg.Stack = config.StackNode
	m.choices = []string{"express", "nextjs", "vite", "nestjs", "generic"}
	m.choiceIdx = 0
	v := m.View()
	if !strings.Contains(v, "express") || !strings.Contains(v, "nextjs") {
		t.Fatalf("app-type view should list Node.js types:\n%s", v)
	}
}

func TestUpdate_Quit(t *testing.T) {
	m := New()
	mm, _ := m.Update(keyMsg("q"))
	m = mm.(Model)
	if !m.quitting {
		t.Fatal("pressing q should set quitting")
	}
}

func TestUpdate_StackSelection(t *testing.T) {
	m := New()
	m.step = StepStack
	m.choiceIdx = 1 // nodejs
	mm, _ := m.Update(keyMsg("enter"))
	m = mm.(Model)
	if m.cfg.Stack != config.StackNode {
		t.Fatalf("expected stack nodejs, got %s", m.cfg.Stack)
	}
	if m.step != StepAppType {
		t.Fatalf("expected step AppType, got %v", m.step)
	}
}

func TestUpdate_AppTypeSelection(t *testing.T) {
	m := New()
	m.step = StepAppType
	m.cfg.Stack = config.StackNode
	m.choices = []string{"express", "nextjs", "vite", "nestjs", "generic"}
	m.choiceIdx = 2 // vite
	mm, _ := m.Update(keyMsg("enter"))
	m = mm.(Model)
	if m.cfg.AppType != config.AppTypeVite {
		t.Fatalf("expected app-type vite, got %s", m.cfg.AppType)
	}
	if m.step != StepCommands {
		t.Fatalf("expected step Commands, got %v", m.step)
	}
}

func TestUpdate_Port(t *testing.T) {
	m := New()
	m.step = StepPort
	m.input.SetValue("8080")
	mm, _ := m.Update(keyMsg("enter"))
	m = mm.(Model)
	if m.cfg.Port != 8080 {
		t.Fatalf("expected port 8080, got %d", m.cfg.Port)
	}
}

func TestUpdate_PortInvalid(t *testing.T) {
	m := New()
	m.step = StepPort
	m.input.SetValue("99999")
	mm, _ := m.Update(keyMsg("enter"))
	m = mm.(Model)
	if m.cfg.Port != 0 {
		t.Fatalf("expected port to remain 0 on invalid input, got %d", m.cfg.Port)
	}
	if m.err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestUpdate_DetectionDone(t *testing.T) {
	m := New()
	msg := detectionDone{
		result: detector.DetectionResult{
			Stack:          "nodejs",
			Indicators:     []string{"package.json"},
			AppCandidates:  []string{"express"},
			BuildCmd:       "npm run build",
			TestCmd:        "jest",
			StartCmd:       "node server.js",
			Lockfile:       "package-lock.json",
			PackageManager: "npm",
		},
	}
	mm, _ := m.Update(msg)
	m = mm.(Model)
	if m.cfg.Stack != config.StackNode {
		t.Fatalf("expected stack nodejs, got %s", m.cfg.Stack)
	}
	if m.cfg.AppType != config.AppTypeExpress {
		t.Fatalf("expected app-type express, got %s", m.cfg.AppType)
	}
	if m.cfg.BuildCmd != "npm run build" {
		t.Fatalf("expected build-cmd, got %s", m.cfg.BuildCmd)
	}
}

func TestUpdate_DetectionError(t *testing.T) {
	m := New()
	msg := detectionDone{
		result: detector.DetectionResult{},
		err:    errors.New("no indicators"),
	}
	mm, _ := m.Update(msg)
	m = mm.(Model)
	if m.err == nil {
		t.Fatal("expected error to be set")
	}
}
