package wizard

import (
	"testing"

	"devopsgen.dev/devopsgen/internal/config"
	"devopsgen.dev/devopsgen/internal/detector"
)

// walkToPort drives a fresh model from the detect step to the port step
// using the natural enter-key flow.
func walkToPort(t *testing.T, stack config.StackType, appType config.AppType) Model {
	t.Helper()
	m := New()
	detection := detectionDone{
		result: detector.DetectionResult{
			Stack:          string(stack),
			Indicators:     []string{"go.mod"},
			AppCandidates:  []string{string(appType)},
			BuildCmd:       "go build",
			TestCmd:        "go test ./...",
			StartCmd:       "./server",
			PackageManager: "npm",
		},
	}
	next, _ := m.Update(detection)
	m = next.(Model)
	// Choose the detected stack and leave the commands step.
	// go:   stack -> commands -> port           => 2 presses
	// node: stack -> app-type -> commands -> port => 3 presses
	presses := 2
	if stack == config.StackNode {
		presses = 3
	}
	for i := 0; i < presses; i++ {
		next, _ = m.Update(keyMsg("enter"))
		m = next.(Model)
	}
	if m.step != StepPort {
		t.Fatalf("expected port step after natural flow, got %v", m.step)
	}
	return m
}

func TestWizard_CharLimitIsStepScoped(t *testing.T) {
	m := walkToPort(t, config.StackGo, config.AppTypeBinary)

	// The port input keeps its tight character limit; commands -> port reset.
	if m.input.CharLimit != 5 {
		t.Fatalf("port input must have CharLimit 5, got %d", m.input.CharLimit)
	}

	// A long registry name must be accepted after leaving the port step.
	m.input.SetValue("8080")
	next, _ := m.Update(keyMsg("enter"))
	m = next.(Model)
	if m.step != StepRegistry {
		t.Fatalf("expected registry step, got %v", m.step)
	}
	if m.input.CharLimit != 0 {
		t.Fatalf("port CharLimit leaked into registry step: %d", m.input.CharLimit)
	}
	m.input.SetValue("ghcr")
	next, _ = m.Update(keyMsg("enter"))
	m = next.(Model)
	if m.step != StepImage {
		t.Fatalf("expected image step, got %v", m.step)
	}
	if m.err != nil {
		t.Fatalf("unexpected error after registry: %v", m.err)
	}

	m.input.SetValue("bangkah/my-app")
	next, _ = m.Update(keyMsg("enter"))
	m = next.(Model)
	if m.step != StepK8s {
		t.Fatalf("expected k8s step, got %v", m.step)
	}
	next, _ = m.Update(keyMsg("enter"))
	m = next.(Model)
	if m.step != StepSummary {
		t.Fatalf("expected summary step, got %v", m.step)
	}
	if m.cfg.Image != "bangkah/my-app" {
		t.Fatalf("expected image to survive later steps, got %q", m.cfg.Image)
	}
	if m.cfg.Port != 8080 {
		t.Fatalf("expected port 8080, got %d", m.cfg.Port)
	}
	if m.cfg.Registry != config.RegistryGHCR {
		t.Fatalf("expected registry ghcr, got %s", m.cfg.Registry)
	}
}

func TestWizard_PortNonNumericRejected(t *testing.T) {
	m := walkToPort(t, config.StackGo, config.AppTypeBinary)
	m.input.SetValue("81a2")
	next, _ := m.Update(keyMsg("enter"))
	m = next.(Model)
	if m.step != StepPort {
		t.Fatalf("non-numeric port must stay on port step, got %v", m.step)
	}
	if m.err == nil {
		t.Fatal("expected error for non-numeric port")
	}
	if m.cfg.Port != 0 {
		t.Fatalf("port must not be applied on invalid input, got %d", m.cfg.Port)
	}
}
