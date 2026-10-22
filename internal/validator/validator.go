// Package validator performs internal artifact validation.
//
// Validation is always internal and runs even when external validators
// (e.g. kubeconform) are unavailable. External validation is additive and
// is reported separately from internal results.
package validator

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
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
//
// Per-artifact validation runs first, followed by cross-resource
// consistency checks (Service selector vs Deployment labels, Ingress
// backend vs Service name, port consistency).
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

	// Cross-resource consistency checks.
	if err := e.validateCrossResource(artifacts); err != nil {
		firstErr = errors.Join(firstErr, err)
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
// A decode error in ANY document is returned; only io.EOF (end of stream)
// terminates the loop normally, so a broken trailing document can no longer
// hide behind a valid leading document (FR-13).
func parseYAMLMulti(content []byte) ([]map[string]any, error) {
	var docs []map[string]any
	dec := yaml.NewDecoder(bytes.NewReader(content))
	for {
		var doc map[string]any
		err := dec.Decode(&doc)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("invalid YAML: %w", err)
		}
		if doc != nil {
			docs = append(docs, doc)
		}
	}
	return docs, nil
}

// validateKindDocs validates every document in the file against its own
// kind, using the file's primary-checker docs as the full contract. A
// document whose kind differs from the file's usual kind is still checked
// with that kind's own rules, so a malformed extra document is never
// silently skipped. Kinds outside MVP scope pass through unchecked.
func (e *engine) validateKindDocs(a output.Artifact, want string, check func(path string, doc map[string]any) error) error {
	docs, err := parseYAMLMulti(a.Content)
	if err != nil {
		return fmt.Errorf("validator: %s %w", a.Path, err)
	}
	if len(docs) == 0 {
		return fmt.Errorf("validator: %s is empty", a.Path)
	}
	for i, doc := range docs {
		path := fmt.Sprintf("%s (document %d)", a.Path, i)
		kind, _ := doc["kind"].(string)
		switch {
		case kind == want:
			if err := check(path, doc); err != nil {
				return err
			}
		case kind == "Deployment":
			if err := e.validateDeploymentDoc(path, doc); err != nil {
				return err
			}
		case kind == "Service":
			if err := e.validateServiceDoc(path, doc); err != nil {
				return err
			}
		case kind == "Ingress":
			if err := e.validateIngressDoc(path, doc); err != nil {
				return err
			}
		}
	}
	return nil
}

// findArtifact returns the artifact with the given path, or nil.
func findArtifact(artifacts []output.Artifact, path string) *output.Artifact {
	for i := range artifacts {
		if artifacts[i].Path == path {
			return &artifacts[i]
		}
	}
	return nil
}

// deploymentSelectorLabels returns the matchLabels from a Deployment artifact.
func deploymentSelectorLabels(doc map[string]any) map[string]any {
	spec, _ := doc["spec"].(map[string]any)
	selector, _ := spec["selector"].(map[string]any)
	matchLabels, _ := selector["matchLabels"].(map[string]any)
	return matchLabels
}

// deploymentPodLabels returns the template labels from a Deployment artifact.
func deploymentPodLabels(doc map[string]any) map[string]any {
	spec, _ := doc["spec"].(map[string]any)
	template, _ := spec["template"].(map[string]any)
	tmplMetadata, _ := template["metadata"].(map[string]any)
	labels, _ := tmplMetadata["labels"].(map[string]any)
	return labels
}

// deploymentContainerPort returns the containerPort from a Deployment artifact.
func deploymentContainerPort(doc map[string]any) int {
	spec, _ := doc["spec"].(map[string]any)
	template, _ := spec["template"].(map[string]any)
	tmplSpec, _ := template["spec"].(map[string]any)
	containers, _ := tmplSpec["containers"].([]any)
	if len(containers) == 0 {
		return 0
	}
	cmap, _ := containers[0].(map[string]any)
	ports, _ := cmap["ports"].([]any)
	if len(ports) == 0 {
		return 0
	}
	portMap, _ := ports[0].(map[string]any)
	containerPort, _ := portMap["containerPort"].(int)
	return containerPort
}

// deploymentImage returns the container image from a Deployment artifact.
func deploymentImage(doc map[string]any) string {
	spec, _ := doc["spec"].(map[string]any)
	template, _ := spec["template"].(map[string]any)
	tmplSpec, _ := template["spec"].(map[string]any)
	containers, _ := tmplSpec["containers"].([]any)
	if len(containers) == 0 {
		return ""
	}
	cmap, _ := containers[0].(map[string]any)
	image, _ := cmap["image"].(string)
	return image
}

// serviceSelector returns the selector from a Service artifact.
func serviceSelector(doc map[string]any) map[string]any {
	spec, _ := doc["spec"].(map[string]any)
	selector, _ := spec["selector"].(map[string]any)
	return selector
}

// servicePorts returns the ports from a Service artifact.
func servicePorts(doc map[string]any) []any {
	spec, _ := doc["spec"].(map[string]any)
	ports, _ := spec["ports"].([]any)
	return ports
}

// serviceTargetPort returns the first targetPort from a Service artifact.
func serviceTargetPort(doc map[string]any) int {
	ports := servicePorts(doc)
	if len(ports) == 0 {
		return 0
	}
	portMap, _ := ports[0].(map[string]any)
	targetPort, _ := portMap["targetPort"].(int)
	return targetPort
}

// ingressBackendServiceName returns the backend service name from an Ingress artifact.
func ingressBackendServiceName(doc map[string]any) string {
	spec, _ := doc["spec"].(map[string]any)
	rules, _ := spec["rules"].([]any)
	if len(rules) == 0 {
		return ""
	}
	rule, _ := rules[0].(map[string]any)
	http, _ := rule["http"].(map[string]any)
	paths, _ := http["paths"].([]any)
	if len(paths) == 0 {
		return ""
	}
	path, _ := paths[0].(map[string]any)
	backend, _ := path["backend"].(map[string]any)
	svc, _ := backend["service"].(map[string]any)
	name, _ := svc["name"].(string)
	return name
}

// findKindDoc parses the artifact at path and returns the first document of
// the wanted kind.
func findKindDoc(artifacts []output.Artifact, path, want string) (map[string]any, error) {
	a := findArtifact(artifacts, path)
	if a == nil {
		return nil, nil
	}
	docs, err := parseYAMLMulti(a.Content)
	if err != nil {
		return nil, fmt.Errorf("validator: %s %w", a.Path, err)
	}
	for _, doc := range docs {
		if kind, _ := doc["kind"].(string); kind == want {
			return doc, nil
		}
	}
	return nil, nil
}

// serviceTargetPortRef returns the targetPort of the first Service port as
// either an integer or a named port reference.
func serviceTargetPortRef(doc map[string]any) (int, string) {
	ports := servicePorts(doc)
	if len(ports) == 0 {
		return 0, ""
	}
	portMap, _ := ports[0].(map[string]any)
	tp, _ := portMap["targetPort"].(int)
	name, _ := portMap["targetPort"].(string)
	return tp, name
}

// deploymentContainerPortRef returns the containerPort of the first container
// port as either an integer or a named port reference.
func deploymentContainerPortRef(doc map[string]any) (int, string) {
	spec, _ := doc["spec"].(map[string]any)
	template, _ := spec["template"].(map[string]any)
	tmplSpec, _ := template["spec"].(map[string]any)
	containers, _ := tmplSpec["containers"].([]any)
	if len(containers) == 0 {
		return 0, ""
	}
	cmap, _ := containers[0].(map[string]any)
	ports, _ := cmap["ports"].([]any)
	if len(ports) == 0 {
		return 0, ""
	}
	portMap, _ := ports[0].(map[string]any)
	cp, _ := portMap["containerPort"].(int)
	name, _ := portMap["name"].(string)
	return cp, name
}

// ingressBackendServicePortRef returns the backend service port reference of
// the first Ingress path as either a port number or a port name.
func ingressBackendServicePortRef(doc map[string]any) (int, string) {
	spec, _ := doc["spec"].(map[string]any)
	rules, _ := spec["rules"].([]any)
	if len(rules) == 0 {
		return 0, ""
	}
	rule, _ := rules[0].(map[string]any)
	http, _ := rule["http"].(map[string]any)
	paths, _ := http["paths"].([]any)
	if len(paths) == 0 {
		return 0, ""
	}
	path, _ := paths[0].(map[string]any)
	backend, _ := path["backend"].(map[string]any)
	svc, _ := backend["service"].(map[string]any)
	port, _ := svc["port"].(map[string]any)
	number, _ := port["number"].(int)
	name, _ := port["name"].(string)
	return number, name
}

// validateCrossResource checks consistency between Kubernetes artifacts.
func (e *engine) validateCrossResource(artifacts []output.Artifact) error {
	var firstErr error

	depDoc, err := findKindDoc(artifacts, "kubernetes/deployment.yaml", "Deployment")
	if err != nil {
		return err
	}
	svcDoc, err := findKindDoc(artifacts, "kubernetes/service.yaml", "Service")
	if err != nil {
		return err
	}

	if depDoc != nil && svcDoc != nil {
		// Service selector must cover Deployment matchLabels.
		svcSelector := serviceSelector(svcDoc)
		depSelector := deploymentSelectorLabels(depDoc)
		for k, v := range depSelector {
			if sv, ok := svcSelector[k]; !ok || sv != v {
				firstErr = errors.Join(firstErr, fmt.Errorf("validator: Service selector %q does not match Deployment matchLabels %q", k, v))
			}
		}

		// and must only use keys the pod template actually carries; a selector
		// key absent from the pod labels selects zero Pods.
		podLabels := deploymentPodLabels(depDoc)
		for k, v := range svcSelector {
			if pv, ok := podLabels[k]; !ok || pv != v {
				firstErr = errors.Join(firstErr, fmt.Errorf("validator: Service selector %q does not match pod template labels %q", k, v))
			}
		}

		// Service targetPort must match Deployment containerPort (integer or
		// named port). When either side carries no port information the check
		// is skipped rather than assumed (an int-typed named targetPort reads
		// as 0, so both channels are consulted).
		svcTPI, svcTPName := serviceTargetPortRef(svcDoc)
		if svcTPI > 0 || svcTPName != "" {
			depI, depName := deploymentContainerPortRef(depDoc)
			switch {
			case svcTPI > 0 && depI > 0 && svcTPI != depI:
				firstErr = errors.Join(firstErr, fmt.Errorf("validator: Service targetPort %d does not match Deployment containerPort %d", svcTPI, depI))
			case svcTPName != "" && depName != "" && svcTPName != depName:
				firstErr = errors.Join(firstErr, fmt.Errorf("validator: Service targetPort %q does not match Deployment container port name %q", svcTPName, depName))
			}
		}
	}

	// Ingress backend service must match the Service name and publish one of
	// its ports. This holds even when no Deployment artifact is present.
	if svcDoc != nil {
		ingDoc, err := findKindDoc(artifacts, "kubernetes/ingress.yaml", "Ingress")
		if err != nil {
			return err
		}
		if ingDoc != nil {
			ingBackend := ingressBackendServiceName(ingDoc)
			svcMetadata, _ := svcDoc["metadata"].(map[string]any)
			svcName, _ := svcMetadata["name"].(string)
			if ingBackend != "" && ingBackend != svcName {
				firstErr = errors.Join(firstErr, fmt.Errorf("validator: Ingress backend service %q does not match Service name %q", ingBackend, svcName))
			}
			if number, name := ingressBackendServicePortRef(ingDoc); number > 0 || name != "" {
				matched := false
				for _, p := range servicePorts(svcDoc) {
					pmap, _ := p.(map[string]any)
					if pmap == nil {
						continue
					}
					if sp, _ := pmap["port"].(int); number > 0 && sp == number {
						matched = true
					}
					if pn, _ := pmap["name"].(string); name != "" && pn == name {
						matched = true
					}
				}
				if !matched {
					firstErr = errors.Join(firstErr, fmt.Errorf("validator: Ingress backend port (%d, %q) does not match any Service port", number, name))
				}
			}
		}
	}

	return firstErr
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
	return e.validateKindDocs(a, "Deployment", e.validateDeploymentDoc)
}

func (e *engine) validateDeploymentDoc(path string, doc map[string]any) error {
	// apiVersion
	apiVersion, ok := doc["apiVersion"]
	if !ok || apiVersion == "" {
		return fmt.Errorf("validator: %s missing apiVersion", path)
	}
	if _, ok := apiVersion.(string); !ok {
		return fmt.Errorf("validator: %s apiVersion must be a string", path)
	}

	// kind
	kind, ok := doc["kind"]
	if !ok || kind != "Deployment" {
		return fmt.Errorf("validator: %s missing or invalid kind: expected Deployment", path)
	}

	// metadata.name
	metadata, ok := doc["metadata"].(map[string]any)
	if !ok {
		return fmt.Errorf("validator: %s missing metadata", path)
	}
	name, ok := metadata["name"]
	if !ok || name == "" {
		return fmt.Errorf("validator: %s missing metadata.name", path)
	}
	if _, ok := name.(string); !ok {
		return fmt.Errorf("validator: %s metadata.name must be a string", path)
	}

	// spec
	spec, ok := doc["spec"].(map[string]any)
	if !ok {
		return fmt.Errorf("validator: %s missing spec", path)
	}

	// spec.selector.matchLabels
	selector, ok := spec["selector"].(map[string]any)
	if !ok {
		return fmt.Errorf("validator: %s missing spec.selector", path)
	}
	matchLabels, ok := selector["matchLabels"].(map[string]any)
	if !ok || len(matchLabels) == 0 {
		return fmt.Errorf("validator: %s missing spec.selector.matchLabels", path)
	}

	// spec.template
	template, ok := spec["template"].(map[string]any)
	if !ok {
		return fmt.Errorf("validator: %s missing spec.template", path)
	}
	tmplMetadata, ok := template["metadata"].(map[string]any)
	if !ok {
		return fmt.Errorf("validator: %s missing spec.template.metadata", path)
	}
	templateLabels, ok := tmplMetadata["labels"].(map[string]any)
	if !ok || len(templateLabels) == 0 {
		return fmt.Errorf("validator: %s missing spec.template.metadata.labels", path)
	}

	// Selector matchLabels must be consistent with template labels.
	for k, v := range matchLabels {
		if tv, ok := templateLabels[k]; !ok || tv != v {
			return fmt.Errorf("validator: %s selector.matchLabels %q not consistent with template labels", path, k)
		}
	}

	// spec.template.spec.containers
	tmplSpec, ok := template["spec"].(map[string]any)
	if !ok {
		return fmt.Errorf("validator: %s missing spec.template.spec", path)
	}
	containers, ok := tmplSpec["containers"].([]any)
	if !ok || len(containers) == 0 {
		return fmt.Errorf("validator: %s missing spec.template.spec.containers", path)
	}

	containerNames := map[string]bool{}
	for i, c := range containers {
		cmap, ok := c.(map[string]any)
		if !ok {
			return fmt.Errorf("validator: %s container %d is not a mapping", path, i)
		}
		cname, ok := cmap["name"]
		if !ok || cname == "" {
			return fmt.Errorf("validator: %s container %d missing name", path, i)
		}
		if _, ok := cname.(string); !ok {
			return fmt.Errorf("validator: %s container %d name must be a string", path, i)
		}
		if containerNames[cname.(string)] {
			return fmt.Errorf("validator: %s duplicate container name %q", path, cname)
		}
		containerNames[cname.(string)] = true

		// Image container not empty.
		image, ok := cmap["image"]
		if !ok || image == "" {
			return fmt.Errorf("validator: %s container %s missing image", path, cname)
		}
		if _, ok := image.(string); !ok {
			return fmt.Errorf("validator: %s container %s image must be a string", path, cname)
		}

		// Container ports must be integers within the valid range.
		cports, ok := cmap["ports"].([]any)
		if !ok {
			continue // ports without any entry are allowed
		}
		for pi, pm := range cports {
			pmap, ok := pm.(map[string]any)
			if !ok {
				return fmt.Errorf("validator: %s container %s ports[%d] is not a mapping", path, cname, pi)
			}
			cp, ok := pmap["containerPort"].(int)
			if !ok {
				return fmt.Errorf("validator: %s container %s ports[%d] containerPort must be an integer", path, cname, pi)
			}
			if cp < 1 || cp > 65535 {
				return fmt.Errorf("validator: %s container %s ports[%d] containerPort %d out of range 1-65535", path, cname, pi, cp)
			}
		}
	}
	return nil
}

func (e *engine) validateService(a output.Artifact) error {
	return e.validateKindDocs(a, "Service", e.validateServiceDoc)
}

func (e *engine) validateServiceDoc(path string, doc map[string]any) error {
	apiVersion, ok := doc["apiVersion"]
	if !ok || apiVersion == "" {
		return fmt.Errorf("validator: %s missing apiVersion", path)
	}
	if _, ok := apiVersion.(string); !ok {
		return fmt.Errorf("validator: %s apiVersion must be a string", path)
	}

	kind, ok := doc["kind"]
	if !ok || kind != "Service" {
		return fmt.Errorf("validator: %s missing or invalid kind: expected Service", path)
	}

	metadata, ok := doc["metadata"].(map[string]any)
	if !ok {
		return fmt.Errorf("validator: %s missing metadata", path)
	}
	name, ok := metadata["name"]
	if !ok || name == "" {
		return fmt.Errorf("validator: %s missing metadata.name", path)
	}

	spec, ok := doc["spec"].(map[string]any)
	if !ok {
		return fmt.Errorf("validator: %s missing spec", path)
	}

	selector, ok := spec["selector"].(map[string]any)
	if !ok || len(selector) == 0 {
		return fmt.Errorf("validator: %s missing spec.selector", path)
	}

	ports, ok := spec["ports"].([]any)
	if !ok || len(ports) == 0 {
		return fmt.Errorf("validator: %s missing spec.ports", path)
	}
	for i, p := range ports {
		pmap, ok := p.(map[string]any)
		if !ok {
			return fmt.Errorf("validator: %s spec.ports[%d] is not a mapping", path, i)
		}
		if _, ok := pmap["port"]; !ok {
			return fmt.Errorf("validator: %s spec.ports[%d] missing port", path, i)
		}
		if _, ok := pmap["targetPort"]; !ok {
			return fmt.Errorf("validator: %s spec.ports[%d] missing targetPort", path, i)
		}
	}
	return nil
}

func (e *engine) validateIngress(a output.Artifact) error {
	return e.validateKindDocs(a, "Ingress", e.validateIngressDoc)
}

func (e *engine) validateIngressDoc(path string, doc map[string]any) error {
	apiVersion, ok := doc["apiVersion"]
	if !ok || apiVersion == "" {
		return fmt.Errorf("validator: %s missing apiVersion", path)
	}
	if apiVersion != "networking.k8s.io/v1" {
		return fmt.Errorf("validator: %s apiVersion must be networking.k8s.io/v1", path)
	}

	kind, ok := doc["kind"]
	if !ok || kind != "Ingress" {
		return fmt.Errorf("validator: %s missing or invalid kind: expected Ingress", path)
	}

	metadata, ok := doc["metadata"].(map[string]any)
	if !ok {
		return fmt.Errorf("validator: %s missing metadata", path)
	}
	if _, ok := metadata["name"]; !ok {
		return fmt.Errorf("validator: %s missing metadata.name", path)
	}

	spec, ok := doc["spec"].(map[string]any)
	if !ok {
		return fmt.Errorf("validator: %s missing spec", path)
	}
	rules, ok := spec["rules"].([]any)
	if !ok || len(rules) == 0 {
		return fmt.Errorf("validator: %s missing spec.rules", path)
	}
	for i, r := range rules {
		rmap, ok := r.(map[string]any)
		if !ok {
			return fmt.Errorf("validator: %s spec.rules[%d] is not a mapping", path, i)
		}
		host, _ := rmap["host"].(string)
		if host == "" {
			return fmt.Errorf("validator: %s spec.rules[%d] missing host", path, i)
		}
		http, ok := rmap["http"].(map[string]any)
		if !ok {
			return fmt.Errorf("validator: %s spec.rules[%d] missing http", path, i)
		}
		paths, ok := http["paths"].([]any)
		if !ok || len(paths) == 0 {
			return fmt.Errorf("validator: %s spec.rules[%d] missing http.paths", path, i)
		}
		for j, p := range paths {
			pmap, ok := p.(map[string]any)
			if !ok {
				return fmt.Errorf("validator: %s spec.rules[%d].http.paths[%d] is not a mapping", path, i, j)
			}
			pt, ok := pmap["pathType"].(string)
			switch {
			case !ok:
				return fmt.Errorf("validator: %s spec.rules[%d].http.paths[%d] missing or invalid pathType", path, i, j)
			case pt != "Exact" && pt != "Prefix" && pt != "ImplementationSpecific":
				return fmt.Errorf("validator: %s spec.rules[%d].http.paths[%d] pathType %q must be Exact, Prefix, or ImplementationSpecific", path, i, j, pt)
			}
			backend, ok := pmap["backend"].(map[string]any)
			if !ok {
				return fmt.Errorf("validator: %s spec.rules[%d].http.paths[%d] missing backend", path, i, j)
			}
			svc, ok := backend["service"].(map[string]any)
			if !ok {
				return fmt.Errorf("validator: %s spec.rules[%d].http.paths[%d] missing backend.service", path, i, j)
			}
			if _, ok := svc["name"]; !ok {
				return fmt.Errorf("validator: %s spec.rules[%d].http.paths[%d] missing backend.service.name", path, i, j)
			}
			if _, ok := svc["port"]; !ok {
				return fmt.Errorf("validator: %s spec.rules[%d].http.paths[%d] missing backend.service.port", path, i, j)
			}
		}
	}
	return nil
}

// unrenderedTemplateRe matches unrendered Go template actions that can
// never appear in a rendered artifact.
var unrenderedTemplateRe = regexp.MustCompile(`\{\{\s*\.`)

// validateWorkflow validates the GitHub Actions workflow structurally:
// the file must parse as exactly one YAML document exposing name, trigger,
// jobs with steps, a checkout step, and a commit-SHA image publish
// (FR-13: structural YAML parsing, not text matching).
func (e *engine) validateWorkflow(a output.Artifact) error {
	s := string(a.Content)
	if unrenderedTemplateRe.MatchString(s) {
		return fmt.Errorf("validator: %s contains unrendered template syntax", a.Path)
	}
	docs, err := parseYAMLMulti(a.Content)
	if err != nil {
		return fmt.Errorf("validator: %s %w", a.Path, err)
	}
	if len(docs) != 1 {
		return fmt.Errorf("validator: %s must contain exactly one workflow document", a.Path)
	}
	doc := docs[0]

	name, ok := doc["name"].(string)
	if !ok || name == "" {
		return fmt.Errorf("validator: %s missing workflow name", a.Path)
	}
	// yaml.v4 follows YAML 1.2: the unquoted `on:` trigger always decodes as
	// the string key "on".
	if _, hasOn := doc["on"]; !hasOn {
		return fmt.Errorf("validator: %s missing on trigger", a.Path)
	}
	jobs, ok := doc["jobs"].(map[string]any)
	if !ok || len(jobs) == 0 {
		return fmt.Errorf("validator: %s missing jobs", a.Path)
	}
	hasCheckout := false
	hasSHA := false
	for _, jv := range jobs {
		job, ok := jv.(map[string]any)
		if !ok {
			continue
		}
		steps, ok := job["steps"].([]any)
		if !ok || len(steps) == 0 {
			return fmt.Errorf("validator: %s missing steps", a.Path)
		}
		for i, sv := range steps {
			step, ok := sv.(map[string]any)
			if !ok {
				return fmt.Errorf("validator: %s steps[%d] is not a mapping", a.Path, i)
			}
			if uses, _ := step["uses"].(string); strings.Contains(uses, "actions/checkout") {
				hasCheckout = true
			}
			if run, _ := step["run"].(string); strings.Contains(run, "github.sha") {
				hasSHA = true
			}
		}
	}
	if !hasCheckout {
		return fmt.Errorf("validator: %s missing checkout step", a.Path)
	}
	if !hasSHA {
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
