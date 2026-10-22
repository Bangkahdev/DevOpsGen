// Package generator renders templates into in-memory artifacts.
//
// The generator is pure: it performs no filesystem I/O, no network access,
// and no credential handling.
package generator

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
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
// Artifacts are returned in deterministic (lexicographic) order.
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
	// Sort by path for deterministic output ordering (NFR-04).
	sort.Slice(out, func(i, j int) bool {
		return out[i].Path < out[j].Path
	})
	return out, nil
}

// specs builds the rendering plan for the configuration.
func (g *engine) specs(cfg config.Config) []templateSpec {
	data := templateData{Config: cfg}
	specs := []templateSpec{
		{name: ".dockerignore", file: "docker/dockerignore.tmpl", delims: [2]string{"{{", "}}"}, data: data},
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
		specs = append(specs, []templateSpec{
			{name: "kubernetes/deployment.yaml", file: "kubernetes/deployment.yaml.tmpl", delims: [2]string{"{{", "}}"}, data: data},
			{name: "kubernetes/service.yaml", file: "kubernetes/service.yaml.tmpl", delims: [2]string{"{{", "}}"}, data: data},
		}...)
		if cfg.K8s.Ingress {
			specs = append(specs, templateSpec{
				name:   "kubernetes/ingress.yaml",
				file:   "kubernetes/ingress.yaml.tmpl",
				delims: [2]string{"{{", "}}"},
				data:   data,
			})
		}
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
// MVP supports only Go and Node.js stacks.
func dockerfileFor(cfg config.Config) string {
	switch {
	case cfg.Stack == config.StackGo:
		return "docker/Dockerfile.go.tmpl"
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
// It is the last path segment of the configured image, sanitized for
// Kubernetes naming rules (DNS-1035 label: lowercase alphanumeric and '-',
// starting with a letter, max 63 chars).
func (t templateData) Name() string {
	return sanitizeK8sName(lastSegment(t.Config.Image))
}

// sanitizeK8sName converts a name into a valid Kubernetes DNS-1035 label:
// lowercase, alphanumerics and hyphens only (dots become hyphens), starting
// with a letter, ending alphanumeric, at most 63 characters.
func sanitizeK8sName(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	// DNS-1035 labels must start with a letter.
	if len(out) > 0 && (out[0] < 'a' || out[0] > 'z') {
		out = "app-" + out
	}
	if len(out) > 63 {
		out = out[:63]
	}
	out = strings.TrimRight(out, "-")
	if out == "" {
		out = "app"
	}
	return out
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

// Tags returns all tags to publish: alias + commit SHA.
func (t templateData) Tags() []string {
	return t.Config.TagList()
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

// BuildOrDefault returns the build command: explicit value, detected
// default for the app type, or empty string.
func (t templateData) BuildOrDefault() string {
	if t.Config.BuildCmd != "" {
		return t.Config.BuildCmd
	}
	pm := t.packageManager()
	switch t.Config.AppType {
	case config.AppTypeNextJS:
		return pm + " run build"
	case config.AppTypeVite:
		return pm + " run build"
	case config.AppTypeNestJS:
		return pm + " run build"
	case config.AppTypeExpress:
		return ""
	case config.AppTypeGeneric:
		return ""
	case config.AppTypeBinary:
		return fmt.Sprintf("go build -ldflags=\"-s -w\" -o /out/server %s", t.buildPath())
	default:
		return fmt.Sprintf("go build -ldflags=\"-s -w\" -o /out/server %s", t.buildPath())
	}
}

// buildPath returns the configured build path, defaulting to the project root.
func (t templateData) buildPath() string {
	if t.Config.BuildPath != "" {
		return t.Config.BuildPath
	}
	return "."
}

// StartOrDefault returns the startup command: explicit value or default.
func (t templateData) StartOrDefault() string {
	if t.Config.StartCmd != "" {
		return t.Config.StartCmd
	}
	pm := t.packageManager()
	switch t.Config.AppType {
	case config.AppTypeExpress:
		return pm + " run start"
	case config.AppTypeNextJS:
		return "npx next start -p " + fmt.Sprintf("%d", t.Config.Port)
	case config.AppTypeVite:
		// Use Vite's built-in preview server instead of the `serve` package,
		// which is not guaranteed to be a project dependency.
		return pm + " run preview -- --port " + fmt.Sprintf("%d", t.Config.Port)
	case config.AppTypeNestJS:
		return "node dist/main.js"
	case config.AppTypeGeneric:
		return pm + " run start"
	case config.AppTypeBinary:
		return "/server"
	default:
		return "/server"
	}
}

// packageManager returns the package manager invocation prefix for Node.js.
// Defaults to npm if none was detected or configured.
func (t templateData) packageManager() string {
	switch t.Config.PackageManager {
	case "pnpm":
		return "pnpm"
	case "yarn":
		return "yarn"
	default:
		return "npm"
	}
}

// InstallCmd returns the dependency-installation command for the Node.js
// builder stage. pnpm/yarn workspaces get their own toolchain; the lockfile
// is always honored when present.
func (t templateData) InstallCmd() string {
	switch t.packageManager() {
	case "pnpm":
		return "npm install -g pnpm@10 && pnpm install --frozen-lockfile"
	case "yarn":
		return "yarn install --frozen-lockfile"
	default:
		return "if [ -f package-lock.json ]; then npm ci; else npm install; fi"
	}
}

// PruneCmd returns the command that strips development dependencies before
// the runtime stage.
func (t templateData) PruneCmd() string {
	switch t.packageManager() {
	case "pnpm":
		return "pnpm install --prod"
	case "yarn":
		return "yarn install --production"
	default:
		return "npm prune --omit=dev"
	}
}

// TestCmdForWorkflow returns the test command as-is.
// Both Go and Node.js test commands are passed through unchanged.
func (t templateData) TestCmdForWorkflow() string {
	if t.Config.TestCmd == "" {
		return ""
	}
	return t.Config.TestCmd
}

// lastSegment returns the final path segment of a slash-separated name.
func lastSegment(s string) string {
	s = strings.TrimSuffix(s, "/")
	if i := strings.LastIndex(s, "/"); i >= 0 {
		return s[i+1:]
	}
	return s
}
