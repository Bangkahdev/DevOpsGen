// Package generator renders templates into in-memory artifacts.
//
// The generator is pure: it performs no filesystem I/O, no network access,
// and no credential handling.
package generator

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"devopsgen.dev/devopsgen/internal/config"
	"devopsgen.dev/devopsgen/internal/output"
	"devopsgen.dev/devopsgen/internal/templates"
)

// Generator is the core rendering contract.
type Generator interface {
	Generate(cfg config.Config) ([]output.Artifact, error)
}

// engine implements Generator using embedded templates.
type engine struct{}

// New returns a Generator backed by embedded templates.
func New() Generator {
	return &engine{}
}

// templateSpec describes how to render one artifact.
type templateSpec struct {
	name   string
	file   string
	delims [2]string
	data   interface{}
}

// Generate renders all artifacts for the given configuration.
func (g *engine) Generate(cfg config.Config) ([]output.Artifact, error) {
	specs := g.specs(cfg)
	out := make([]output.Artifact, 0, len(specs))
	for _, s := range specs {
		art, err := g.render(s)
		if err != nil {
			return nil, err
		}
		out = append(out, art)
	}
	return out, nil
}

// specs builds the rendering plan for the configuration.
func (g *engine) specs(cfg config.Config) []templateSpec {
	data := templateData{Config: cfg}
	specs := []templateSpec{
		{name: "Dockerfile", file: dockerfileFor(cfg), delims: [2]string{"{{", "}}"}, data: data},
	}
	if cfg.CI == config.CIGitHubActions {
		specs = append(specs, templateSpec{
			name:   ".github/workflows/ci.yml",
			file:   "github-actions/ci-workflow.yaml.tmpl",
			delims: [2]string{"[[", "]]"},
			data:   data,
		})
	}
	if cfg.K8s.Enabled {
		specs = append(specs, templateSpec{
			name:   "kubernetes/deployment.yaml",
			file:   "kubernetes/deployment.yaml.tmpl",
			delims: [2]string{"{{", "}}"},
			data:   data,
		})
	}
	return specs
}

func (g *engine) render(s templateSpec) (output.Artifact, error) {
	src, err := templates.Read(s.file)
	if err != nil {
		return output.Artifact{}, fmt.Errorf("template %s: %w", s.file, err)
	}
	t, err := template.New(s.name).Delims(s.delims[0], s.delims[1]).Parse(string(src))
	if err != nil {
		return output.Artifact{}, fmt.Errorf("template %s parse: %w", s.file, err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, s.data); err != nil {
		return output.Artifact{}, fmt.Errorf("template %s render: %w", s.file, err)
	}
	return output.Artifact{
		Path:    filepath.ToSlash(filepath.Clean(s.name)),
		Content: buf.Bytes(),
	}, nil
}

// dockerfileFor selects the Dockerfile template for the configuration.
func dockerfileFor(cfg config.Config) string {
	switch {
	case cfg.Stack == config.StackGo:
		return "docker/Dockerfile.go.tmpl"
	case cfg.Stack == config.StackPy:
		return "docker/Dockerfile.python.tmpl"
	case cfg.AppType == config.AppTypeExpress:
		return "docker/Dockerfile.express.tmpl"
	case cfg.AppType == config.AppTypeNextJS:
		return "docker/Dockerfile.nextjs.tmpl"
	case cfg.AppType == config.AppTypeVite:
		return "docker/Dockerfile.vite.tmpl"
	case cfg.AppType == config.AppTypeNestJS:
		return "docker/Dockerfile.nestjs.tmpl"
	default:
		return "docker/Dockerfile.generic.tmpl"
	}
}

// templateData is the data passed to templates.
type templateData struct {
	config.Config
}

// Name derives a Kubernetes resource name from the image repository.
// It is the last path segment of the configured image.
func (t templateData) Name() string {
	return lastSegment(t.Config.Image)
}

// ImageRef is exposed to templates as a top-level field for convenience.
func (t templateData) ImageRef() string {
	return t.Config.ImageRef()
}

// Tag resolves the tag used by manifests: the configured alias or "latest".
func (t templateData) Tag() string {
	if t.Config.Tag != "" {
		return t.Config.Tag
	}
	return "latest"
}

// IngressHost unwraps K8sOptions for templates.
func (t templateData) IngressHost() string {
	return t.Config.K8s.IngressHost
}

// HealthPath unwraps K8sOptions for templates.
func (t templateData) HealthPath() string {
	return t.Config.K8s.HealthPath
}

// Ingress reports whether the Ingress manifest should be generated.
func (t templateData) Ingress() bool {
	return t.Config.K8s.Ingress
}

// K8sEnabled reports whether Kubernetes manifests should be generated.
func (t templateData) K8sEnabled() bool {
	return t.Config.K8s.Enabled
}

// lastSegment returns the final path segment of a slash-separated name.
func lastSegment(s string) string {
	s = strings.TrimSuffix(s, "/")
	if i := strings.LastIndex(s, "/"); i >= 0 {
		return s[i+1:]
	}
	return s
}
