package wizard

import (
	"testing"

	"devopsgen.dev/devopsgen/internal/config"
)

func TestUpdate_RegistryStep(t *testing.T) {
	m := New()
	m.step = StepRegistry
	m.input.SetValue("ghcr")
	mm, _ := m.Update(keyMsg("enter"))
	m = mm.(Model)
	if m.cfg.Registry != config.RegistryGHCR {
		t.Fatalf("expected ghcr, got %s", m.cfg.Registry)
	}
}

func TestUpdate_ImageStep(t *testing.T) {
	m := New()
	m.step = StepImage
	m.input.SetValue("bangkah/my-app")
	mm, _ := m.Update(keyMsg("enter"))
	m = mm.(Model)
	if m.cfg.Image != "bangkah/my-app" {
		t.Fatalf("expected bangkah/my-app, got %s", m.cfg.Image)
	}
}

func TestUpdate_K8sStep(t *testing.T) {
	m := New()
	m.step = StepK8s
	m.input.SetValue("/health")
	mm, _ := m.Update(keyMsg("enter"))
	m = mm.(Model)
	if m.cfg.K8s.HealthPath != "/health" {
		t.Fatalf("expected /health, got %s", m.cfg.K8s.HealthPath)
	}
}

func TestUpdate_PortEmpty(t *testing.T) {
	m := New()
	m.step = StepPort
	m.input.SetValue("")
	mm, _ := m.Update(keyMsg("enter"))
	m = mm.(Model)
	if m.cfg.Port != 0 {
		t.Fatalf("expected port to remain 0, got %d", m.cfg.Port)
	}
}
