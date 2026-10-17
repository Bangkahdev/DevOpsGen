package validator

import (
	"testing"
)

func TestParseYAML_Valid(t *testing.T) {
	doc, err := parseYAML([]byte("key: value\nlist:\n  - a\n  - b\n"))
	if err != nil {
		t.Fatalf("parseYAML: %v", err)
	}
	if doc["key"] != "value" {
		t.Fatalf("expected key=value, got %v", doc["key"])
	}
}

func TestParseYAML_Empty(t *testing.T) {
	doc, err := parseYAML([]byte(""))
	if err != nil {
		t.Fatalf("parseYAML empty: %v", err)
	}
	if len(doc) != 0 {
		t.Fatalf("expected empty doc, got %v", doc)
	}
}

func TestParseYAML_Invalid(t *testing.T) {
	_, err := parseYAML([]byte(": : : invalid yaml ::"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestDeploymentPodLabels_Normal(t *testing.T) {
	doc := map[string]any{
		"spec": map[string]any{
			"template": map[string]any{
				"metadata": map[string]any{
					"labels": map[string]any{"app": "myapp", "env": "prod"},
				},
			},
		},
	}
	labels := deploymentPodLabels(doc)
	if labels["app"] != "myapp" || labels["env"] != "prod" {
		t.Fatalf("unexpected labels: %v", labels)
	}
}

func TestDeploymentPodLabels_Missing(t *testing.T) {
	if labels := deploymentPodLabels(map[string]any{}); labels != nil {
		t.Fatalf("expected nil labels, got %v", labels)
	}
	if labels := deploymentPodLabels(map[string]any{"spec": map[string]any{}}); labels != nil {
		t.Fatalf("expected nil labels, got %v", labels)
	}
}

func TestDeploymentImage_Normal(t *testing.T) {
	doc := map[string]any{
		"spec": map[string]any{
			"template": map[string]any{
				"spec": map[string]any{
					"containers": []any{
						map[string]any{"image": "ghcr.io/x/app:v1"},
					},
				},
			},
		},
	}
	if img := deploymentImage(doc); img != "ghcr.io/x/app:v1" {
		t.Fatalf("expected image, got %q", img)
	}
}

func TestDeploymentImage_Missing(t *testing.T) {
	if img := deploymentImage(map[string]any{}); img != "" {
		t.Fatalf("expected empty image, got %q", img)
	}
	if img := deploymentImage(map[string]any{"spec": map[string]any{}}); img != "" {
		t.Fatalf("expected empty image, got %q", img)
	}
}

func TestDeploymentContainerPort_Missing(t *testing.T) {
	if port := deploymentContainerPort(map[string]any{}); port != 0 {
		t.Fatalf("expected port 0, got %d", port)
	}
}

func TestValidateCrossResource_Empty(t *testing.T) {
	eng := New().(*engine)
	if err := eng.validateCrossResource(nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestResultString(t *testing.T) {
	if ResultValid.String() != "valid" {
		t.Fatalf("expected valid")
	}
	if ResultInvalid.String() != "invalid" {
		t.Fatalf("expected invalid")
	}
	if ResultExternalSkipped.String() != "external_skipped" {
		t.Fatalf("expected external_skipped")
	}
	if Result(99).String() != "unknown" {
		t.Fatalf("expected unknown")
	}
}
