// Package validator performs internal artifact validation.
//
// Validation is always internal and runs even when external validators
// (e.g. kubeconform) are unavailable. External validation is additive and
// is reported separately from internal results.
package validator

import (
	"errors"
	"fmt"
	"strings"

	"devopsgen.dev/devopsgen/internal/output"
)

// Result categorizes a validation outcome.
type Result int

const (
	ResultValid Result = iota
	ResultInvalid
	ResultExternalSkipped
)

// Validator is the artifact validation contract.
type Validator interface {
	Validate(artifacts []output.Artifact) (Result, error)
}

// engine implements Validator.
type engine struct{}

// New returns a Validator.
func New() Validator {
	return &engine{}
}

// Validate runs all internal checks over the artifacts.
func (e *engine) Validate(artifacts []output.Artifact) (Result, error) {
	if len(artifacts) == 0 {
		return ResultInvalid, errors.New("validator: no artifacts to validate")
	}
	var firstErr error
	for _, a := range artifacts {
		if err := e.validateArtifact(a); err != nil {
			firstErr = errors.Join(firstErr, err)
		}
	}
	if firstErr != nil {
		return ResultInvalid, firstErr
	}
	return ResultValid, nil
}

func (e *engine) validateArtifact(a output.Artifact) error {
	switch {
	case a.Path == "Dockerfile":
		return e.validateDockerfile(a)
	case strings.HasSuffix(a.Path, "deployment.yaml"):
		return e.validateDeployment(a)
	case strings.HasSuffix(a.Path, "service.yaml"):
		return e.validateService(a)
	case strings.HasSuffix(a.Path, "ingress.yaml"):
		return e.validateIngress(a)
	case strings.HasSuffix(a.Path, "ci.yml") || strings.HasSuffix(a.Path, "ci.yaml"):
		return e.validateWorkflow(a)
	default:
		return nil
	}
}

func (e *engine) validateDockerfile(a output.Artifact) error {
	s := string(a.Content)
	if !strings.Contains(s, "FROM ") {
		return fmt.Errorf("validator: %s missing FROM instruction", a.Path)
	}
	if strings.Contains(s, "docker login") {
		return fmt.Errorf("validator: %s contains credential login in image build", a.Path)
	}
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if (strings.HasPrefix(trimmed, "COPY ") || strings.HasPrefix(trimmed, "ADD ")) &&
			(strings.Contains(trimmed, "2>/dev/null") || strings.Contains(trimmed, "||") || strings.Contains(trimmed, "&&")) {
			return fmt.Errorf("validator: %s contains shell syntax in %s instruction", a.Path, strings.Fields(trimmed)[0])
		}
	}
	return nil
}

func (e *engine) validateDeployment(a output.Artifact) error {
	s := string(a.Content)
	for _, want := range []string{"kind: Deployment", "apiVersion:", "spec:", "selector:", "template:", "containers:", "image:"} {
		if !strings.Contains(s, want) {
			return fmt.Errorf("validator: %s missing %q", a.Path, want)
		}
	}
	return nil
}

func (e *engine) validateService(a output.Artifact) error {
	s := string(a.Content)
	for _, want := range []string{"kind: Service", "ports:", "targetPort:"} {
		if !strings.Contains(s, want) {
			return fmt.Errorf("validator: %s missing %q", a.Path, want)
		}
	}
	return nil
}

func (e *engine) validateIngress(a output.Artifact) error {
	s := string(a.Content)
	for _, want := range []string{"kind: Ingress", "networking.k8s.io/v1", "pathType"} {
		if !strings.Contains(s, want) {
			return fmt.Errorf("validator: %s missing %q", a.Path, want)
		}
	}
	return nil
}

func (e *engine) validateWorkflow(a output.Artifact) error {
	s := string(a.Content)
	for _, want := range []string{"name: CI", "on:", "jobs:", "steps:", "uses: actions/checkout"} {
		if !strings.Contains(s, want) {
			return fmt.Errorf("validator: %s missing %q", a.Path, want)
		}
	}
	if strings.Contains(s, "[[") || strings.Contains(s, "{{.") {
		return fmt.Errorf("validator: %s contains unrendered template syntax", a.Path)
	}
	if strings.Contains(s, "github.sha") {
		return nil
	}
	return fmt.Errorf("validator: %s does not publish commit SHA tag", a.Path)
}

// String returns a human-readable validation result.
func (r Result) String() string {
	switch r {
	case ResultValid:
		return "valid"
	case ResultInvalid:
		return "invalid"
	case ResultExternalSkipped:
		return "external_skipped"
	default:
		return "unknown"
	}
}
