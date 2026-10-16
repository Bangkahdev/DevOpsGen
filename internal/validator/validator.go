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

	"go.yaml.in/yaml/v4"

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

// parseYAML parses a single YAML document into a map[string]any.
func parseYAML(content []byte) (map[string]any, error) {
	var doc map[string]any
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}
	if doc == nil {
		return map[string]any{}, nil
	}
	return doc, nil
}

// parseYAMLMulti parses all YAML documents in a multi-document file.
func parseYAMLMulti(content []byte) ([]map[string]any, error) {
	var docs []map[string]any
	dec := yaml.NewDecoder(strings.NewReader(string(content)))
	for {
		var doc map[string]any
		if err := dec.Decode(&doc); err != nil {
			break
		}
		if doc != nil {
			docs = append(docs, doc)
		}
	}
	if len(docs) == 0 {
		return docs, nil
	}
	return docs, nil
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
	docs, err := parseYAMLMulti(a.Content)
	if err != nil {
		return err
	}
	if len(docs) == 0 {
		return fmt.Errorf("validator: %s is empty", a.Path)
	}
	doc := docs[0]

	// apiVersion
	apiVersion, ok := doc["apiVersion"]
	if !ok || apiVersion == "" {
		return fmt.Errorf("validator: %s missing apiVersion", a.Path)
	}
	if _, ok := apiVersion.(string); !ok {
		return fmt.Errorf("validator: %s apiVersion must be a string", a.Path)
	}

	// kind
	kind, ok := doc["kind"]
	if !ok || kind != "Deployment" {
		return fmt.Errorf("validator: %s missing or invalid kind: expected Deployment", a.Path)
	}

	// metadata.name
	metadata, ok := doc["metadata"].(map[string]any)
	if !ok {
		return fmt.Errorf("validator: %s missing metadata", a.Path)
	}
	name, ok := metadata["name"]
	if !ok || name == "" {
		return fmt.Errorf("validator: %s missing metadata.name", a.Path)
	}
	if _, ok := name.(string); !ok {
		return fmt.Errorf("validator: %s metadata.name must be a string", a.Path)
	}

	// spec
	spec, ok := doc["spec"].(map[string]any)
	if !ok {
		return fmt.Errorf("validator: %s missing spec", a.Path)
	}

	// spec.selector.matchLabels
	selector, ok := spec["selector"].(map[string]any)
	if !ok {
		return fmt.Errorf("validator: %s missing spec.selector", a.Path)
	}
	matchLabels, ok := selector["matchLabels"].(map[string]any)
	if !ok || len(matchLabels) == 0 {
		return fmt.Errorf("validator: %s missing spec.selector.matchLabels", a.Path)
	}

	// spec.template
	template, ok := spec["template"].(map[string]any)
	if !ok {
		return fmt.Errorf("validator: %s missing spec.template", a.Path)
	}
	tmplMetadata, ok := template["metadata"].(map[string]any)
	if !ok {
		return fmt.Errorf("validator: %s missing spec.template.metadata", a.Path)
	}
	templateLabels, ok := tmplMetadata["labels"].(map[string]any)
	if !ok || len(templateLabels) == 0 {
		return fmt.Errorf("validator: %s missing spec.template.metadata.labels", a.Path)
	}

	// Selector matchLabels must be consistent with template labels.
	for k, v := range matchLabels {
		if tv, ok := templateLabels[k]; !ok || tv != v {
			return fmt.Errorf("validator: %s selector.matchLabels %q not consistent with template labels", a.Path, k)
		}
	}

	// spec.template.spec.containers
	tmplSpec, ok := template["spec"].(map[string]any)
	if !ok {
		return fmt.Errorf("validator: %s missing spec.template.spec", a.Path)
	}
	containers, ok := tmplSpec["containers"].([]any)
	if !ok || len(containers) == 0 {
		return fmt.Errorf("validator: %s missing spec.template.spec.containers", a.Path)
	}

	containerNames := map[string]bool{}
	for i, c := range containers {
		cmap, ok := c.(map[string]any)
		if !ok {
			return fmt.Errorf("validator: %s container %d is not a mapping", a.Path, i)
		}
		cname, ok := cmap["name"]
		if !ok || cname == "" {
			return fmt.Errorf("validator: %s container %d missing name", a.Path, i)
		}
		if _, ok := cname.(string); !ok {
			return fmt.Errorf("validator: %s container %d name must be a string", a.Path, i)
		}
		if containerNames[cname.(string)] {
			return fmt.Errorf("validator: %s duplicate container name %q", a.Path, cname)
		}
		containerNames[cname.(string)] = true

		// Image container not empty.
		image, ok := cmap["image"]
		if !ok || image == "" {
			return fmt.Errorf("validator: %s container %s missing image", a.Path, cname)
		}
		if _, ok := image.(string); !ok {
			return fmt.Errorf("validator: %s container %s image must be a string", a.Path, cname)
		}
	}
	return nil
}

func (e *engine) validateService(a output.Artifact) error {
	docs, err := parseYAMLMulti(a.Content)
	if err != nil {
		return err
	}
	if len(docs) == 0 {
		return fmt.Errorf("validator: %s is empty", a.Path)
	}
	doc := docs[0]

	apiVersion, ok := doc["apiVersion"]
	if !ok || apiVersion == "" {
		return fmt.Errorf("validator: %s missing apiVersion", a.Path)
	}
	if _, ok := apiVersion.(string); !ok {
		return fmt.Errorf("validator: %s apiVersion must be a string", a.Path)
	}

	kind, ok := doc["kind"]
	if !ok || kind != "Service" {
		return fmt.Errorf("validator: %s missing or invalid kind: expected Service", a.Path)
	}

	metadata, ok := doc["metadata"].(map[string]any)
	if !ok {
		return fmt.Errorf("validator: %s missing metadata", a.Path)
	}
	name, ok := metadata["name"]
	if !ok || name == "" {
		return fmt.Errorf("validator: %s missing metadata.name", a.Path)
	}

	spec, ok := doc["spec"].(map[string]any)
	if !ok {
		return fmt.Errorf("validator: %s missing spec", a.Path)
	}

	selector, ok := spec["selector"].(map[string]any)
	if !ok || len(selector) == 0 {
		return fmt.Errorf("validator: %s missing spec.selector", a.Path)
	}

	ports, ok := spec["ports"].([]any)
	if !ok || len(ports) == 0 {
		return fmt.Errorf("validator: %s missing spec.ports", a.Path)
	}
	for i, p := range ports {
		pmap, ok := p.(map[string]any)
		if !ok {
			return fmt.Errorf("validator: %s spec.ports[%d] is not a mapping", a.Path, i)
		}
		if _, ok := pmap["port"]; !ok {
			return fmt.Errorf("validator: %s spec.ports[%d] missing port", a.Path, i)
		}
		if _, ok := pmap["targetPort"]; !ok {
			return fmt.Errorf("validator: %s spec.ports[%d] missing targetPort", a.Path, i)
		}
	}
	return nil
}

func (e *engine) validateIngress(a output.Artifact) error {
	docs, err := parseYAMLMulti(a.Content)
	if err != nil {
		return err
	}
	if len(docs) == 0 {
		return fmt.Errorf("validator: %s is empty", a.Path)
	}
	doc := docs[0]

	apiVersion, ok := doc["apiVersion"]
	if !ok || apiVersion == "" {
		return fmt.Errorf("validator: %s missing apiVersion", a.Path)
	}
	if apiVersion != "networking.k8s.io/v1" {
		return fmt.Errorf("validator: %s apiVersion must be networking.k8s.io/v1", a.Path)
	}

	kind, ok := doc["kind"]
	if !ok || kind != "Ingress" {
		return fmt.Errorf("validator: %s missing or invalid kind: expected Ingress", a.Path)
	}

	metadata, ok := doc["metadata"].(map[string]any)
	if !ok {
		return fmt.Errorf("validator: %s missing metadata", a.Path)
	}
	if _, ok := metadata["name"]; !ok {
		return fmt.Errorf("validator: %s missing metadata.name", a.Path)
	}

	spec, ok := doc["spec"].(map[string]any)
	if !ok {
		return fmt.Errorf("validator: %s missing spec", a.Path)
	}
	rules, ok := spec["rules"].([]any)
	if !ok || len(rules) == 0 {
		return fmt.Errorf("validator: %s missing spec.rules", a.Path)
	}
	for i, r := range rules {
		rmap, ok := r.(map[string]any)
		if !ok {
			return fmt.Errorf("validator: %s spec.rules[%d] is not a mapping", a.Path, i)
		}
		host, _ := rmap["host"].(string)
		if host == "" {
			return fmt.Errorf("validator: %s spec.rules[%d] missing host", a.Path, i)
		}
		http, ok := rmap["http"].(map[string]any)
		if !ok {
			return fmt.Errorf("validator: %s spec.rules[%d] missing http", a.Path, i)
		}
		paths, ok := http["paths"].([]any)
		if !ok || len(paths) == 0 {
			return fmt.Errorf("validator: %s spec.rules[%d] missing http.paths", a.Path, i)
		}
		for j, p := range paths {
			pmap, ok := p.(map[string]any)
			if !ok {
				return fmt.Errorf("validator: %s spec.rules[%d].http.paths[%d] is not a mapping", a.Path, i, j)
			}
			if _, ok := pmap["pathType"]; !ok {
				return fmt.Errorf("validator: %s spec.rules[%d].http.paths[%d] missing pathType", a.Path, i, j)
			}
			backend, ok := pmap["backend"].(map[string]any)
			if !ok {
				return fmt.Errorf("validator: %s spec.rules[%d].http.paths[%d] missing backend", a.Path, i, j)
			}
			svc, ok := backend["service"].(map[string]any)
			if !ok {
				return fmt.Errorf("validator: %s spec.rules[%d].http.paths[%d] missing backend.service", a.Path, i, j)
			}
			if _, ok := svc["name"]; !ok {
				return fmt.Errorf("validator: %s spec.rules[%d].http.paths[%d] missing backend.service.name", a.Path, i, j)
			}
			if _, ok := svc["port"]; !ok {
				return fmt.Errorf("validator: %s spec.rules[%d].http.paths[%d] missing backend.service.port", a.Path, i, j)
			}
		}
	}
	return nil
}

func (e *engine) validateWorkflow(a output.Artifact) error {
	s := string(a.Content)
	if !strings.Contains(s, "name: CI") {
		return fmt.Errorf("validator: %s missing workflow name", a.Path)
	}
	// Check for `on:` as a YAML key (start of line), not as substring of runs-on.
	hasOnKey := false
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "on:" || strings.HasPrefix(trimmed, "on:") {
			hasOnKey = true
			break
		}
	}
	if !hasOnKey {
		return fmt.Errorf("validator: %s missing on trigger", a.Path)
	}
	if !strings.Contains(s, "jobs:") {
		return fmt.Errorf("validator: %s missing jobs", a.Path)
	}
	if !strings.Contains(s, "steps:") {
		return fmt.Errorf("validator: %s missing steps", a.Path)
	}
	if !strings.Contains(s, "uses: actions/checkout") {
		return fmt.Errorf("validator: %s missing checkout step", a.Path)
	}
	if strings.Contains(s, "{{.") {
		return fmt.Errorf("validator: %s contains unrendered template syntax", a.Path)
	}
	if !strings.Contains(s, "github.sha") {
		return fmt.Errorf("validator: %s does not publish commit SHA tag", a.Path)
	}
	return nil
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
