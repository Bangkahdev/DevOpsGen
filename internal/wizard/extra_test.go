package wizard

import (
	"errors"
	"strings"
	"testing"

	"devopsgen.dev/devopsgen/internal/config"
	"devopsgen.dev/devopsgen/internal/detector"
)

func TestUpdate_Registry(t *testing.T) {
	m := New()
	m.step = StepRegistry
	m.input.SetValue("ghcr")
	mm, _ := m.Update(keyMsg("enter"))
	m = mm.(Model)
	if m.cfg.Registry != config.RegistryGHCR {
		t.Fatalf("expected registry ghcr, got %s", m.cfg.Registry)
	}
	if m.step != StepImage {
		t.Fatalf("expected step Image, got %v", m.step)
	}
}

func TestUpdate_RegistryInvalid(t *testing.T) {
	m := New()
	m.step = StepRegistry
	m.input.SetValue("registry.gitlab")
	mm, _ := m.Update(keyMsg("enter"))
	m = mm.(Model)
	if m.step != StepRegistry {
		t.Fatalf("expected step to remain Registry, got %v", m.step)
	}
	if m.err == nil {
		t.Fatal("expected error for invalid registry")
	}
}

func TestUpdate_Image(t *testing.T) {
	m := New()
	m.step = StepImage
	m.input.SetValue("bangkah/my-app")
	mm, _ := m.Update(keyMsg("enter"))
	m = mm.(Model)
	if m.cfg.Image != "bangkah/my-app" {
		t.Fatalf("expected image bangkah/my-app, got %s", m.cfg.Image)
	}
	if m.step != StepK8s {
		t.Fatalf("expected step K8s, got %v", m.step)
	}
}

func TestUpdate_ImageInvalid(t *testing.T) {
	m := New()
	m.step = StepImage
	m.input.SetValue("registry.example.com/a/b")
	mm, _ := m.Update(keyMsg("enter"))
	m = mm.(Model)
	if m.step != StepImage {
		t.Fatalf("expected step to remain Image, got %v", m.step)
	}
	if m.err == nil {
		t.Fatal("expected error for invalid image")
	}
}

func TestUpdate_K8s(t *testing.T) {
	m := New()
	m.step = StepK8s
	m.input.SetValue("/health")
	mm, _ := m.Update(keyMsg("enter"))
	m = mm.(Model)
	if m.cfg.K8s.HealthPath != "/health" {
		t.Fatalf("expected health-path /health, got %s", m.cfg.K8s.HealthPath)
	}
	if m.step != StepSummary {
		t.Fatalf("expected step Summary, got %v", m.step)
	}
}

func TestUpdate_Summary_Valid(t *testing.T) {
	m := New()
	m.step = StepSummary
	m.cfg = config.Config{
		Stack:    config.StackGo,
		AppType:  config.AppTypeBinary,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryDockerHub,
		Image:    "a/b",
		Port:     8080,
		Tag:      "v1",
		K8s:      config.K8sOptions{Enabled: true},
	}
	_, cmd := m.Update(keyMsg("enter"))
	if cmd == nil {
		t.Fatal("expected tea.Quit command on valid summary")
	}
}

func TestUpdate_Summary_Invalid(t *testing.T) {
	m := New()
	m.step = StepSummary
	m.cfg = config.Config{
		Stack:    config.StackGo,
		AppType:  config.AppTypeBinary,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryDockerHub,
		Image:    "a/b",
		Port:     8080,
		Tag:      "v1;rm",
		K8s:      config.K8sOptions{Enabled: true},
	}
	mm, _ := m.Update(keyMsg("enter"))
	m = mm.(Model)
	if m.quitting {
		t.Fatal("should not quit on invalid config")
	}
	if m.err == nil {
		t.Fatal("expected error for invalid tag in summary")
	}
}

func TestUpdate_CommandsStep(t *testing.T) {
	m := New()
	m.step = StepCommands
	m.cfg.BuildCmd = "npm run build"
	m.cfg.TestCmd = "jest"
	m.cfg.StartCmd = "node server.js"
	mm, _ := m.Update(keyMsg("enter"))
	m = mm.(Model)
	if m.step != StepPort {
		t.Fatalf("expected step Port, got %v", m.step)
	}
	if m.cfg.BuildCmd != "npm run build" {
		t.Fatalf("build-cmd should be preserved: %s", m.cfg.BuildCmd)
	}
}

func TestView_Commands(t *testing.T) {
	m := New()
	m.step = StepCommands
	m.detect.BuildCmd = "npm run build"
	m.detect.TestCmd = "jest"
	v := m.View()
	if !stringsContains(v, "npm run build") || !stringsContains(v, "jest") {
		t.Fatalf("commands view should show detected commands:\n%s", v)
	}
}

func stringsContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestView_DetectWithError(t *testing.T) {
	m := New()
	m.step = StepDetect
	m.err = errors.New("detection failed")
	v := m.View()
	if !strings.Contains(v, "detection failed") {
		t.Fatalf("detect view should show error:\n%s", v)
	}
}

func TestKeyMsg_String(t *testing.T) {
	km := keyMsg("enter")
	if km.String() != "enter" {
		t.Fatalf("expected 'enter', got %q", km.String())
	}
}

func TestUpdate_DetectionDone_NoStack(t *testing.T) {
	m := New()
	msg := detectionDone{
		result: detector.DetectionResult{},
	}
	mm, _ := m.Update(msg)
	m = mm.(Model)
	if m.step != StepStack {
		t.Fatalf("expected step Stack, got %v", m.step)
	}
}

func TestUpdate_PortEmptyExtra(t *testing.T) {
	m := New()
	m.step = StepPort
	m.input.SetValue("")
	mm, _ := m.Update(keyMsg("enter"))
	m = mm.(Model)
	if m.step != StepRegistry {
		t.Fatalf("expected step Registry, got %v", m.step)
	}
	if m.cfg.Port != 0 {
		t.Fatalf("expected port to remain 0, got %d", m.cfg.Port)
	}
}

func TestUpdate_ImageEmpty(t *testing.T) {
	m := New()
	m.step = StepImage
	m.input.SetValue("")
	mm, _ := m.Update(keyMsg("enter"))
	m = mm.(Model)
	if m.step != StepK8s {
		t.Fatalf("expected step K8s, got %v", m.step)
	}
}

func TestInit_ReturnsCmd(t *testing.T) {
	m := New()
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init should return a non-nil command")
	}
}

func TestView_DetectNotDetected(t *testing.T) {
	m := New()
	m.step = StepDetect
	v := m.View()
	if !strings.Contains(v, "Detecting") {
		t.Fatalf("detect view should show detecting message:\n%s", v)
	}
}

func TestView_K8sStep(t *testing.T) {
	m := New()
	m.step = StepK8s
	v := m.View()
	if !strings.Contains(v, "health") {
		t.Fatalf("k8s view should mention health:\n%s", v)
	}
}

func TestView_SummaryWithIngress(t *testing.T) {
	m := New()
	m.step = StepSummary
	m.cfg = config.Config{
		Stack:    config.StackNode,
		AppType:  config.AppTypeExpress,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryGHCR,
		Image:    "a/b",
		Port:     3000,
		K8s:      config.K8sOptions{Enabled: true, Ingress: true, IngressHost: "app.example.com"},
	}
	v := m.View()
	if !strings.Contains(v, "app.example.com") {
		t.Fatalf("summary should show ingress host:\n%s", v)
	}
}

func TestRunDetection(t *testing.T) {
	msg := runDetection()
	if msg == nil {
		t.Fatal("runDetection should return a non-nil message")
	}
}

func TestIsInteractive_True(t *testing.T) {
	orig := interactiveFn
	defer func() { interactiveFn = orig }()
	interactiveFn = func() bool { return true }
	if !isInteractive() {
		t.Fatal("expected interactive to be true")
	}
}

func TestUpdate_KeyNavigation(t *testing.T) {
	m := New()
	m.step = StepStack
	m.choiceIdx = 0
	// Press down to move to nodejs.
	mm, _ := m.Update(keyMsg("down"))
	m = mm.(Model)
	if m.choiceIdx != 1 {
		t.Fatalf("expected choiceIdx 1, got %d", m.choiceIdx)
	}
	// Press up to move back to go.
	mm, _ = m.Update(keyMsg("up"))
	m = mm.(Model)
	if m.choiceIdx != 0 {
		t.Fatalf("expected choiceIdx 0, got %d", m.choiceIdx)
	}
}
