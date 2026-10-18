// Package cli provides flag definitions and parsing for the init command.
//
// Flag parsing maps user-supplied values onto the config.Config model and
// runs the same validation the core engine uses. It never touches the
// filesystem, network, or credentials.
package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"devopsgen.dev/devopsgen/internal/config"
)

// InitFlags holds every user-supplied parameter for `devopsgen init`.
type InitFlags struct {
	Stack       string
	AppType     string
	CI          string
	Registry    string
	Image       string
	Port        int
	BuildCmd    string
	BuildPath   string
	TestCmd     string
	StartCmd    string
	Tag         string
	K8s         bool
	Ingress     bool
	IngressHost string
	HealthPath  string
	OutputDir   string
	Force       bool
	DryRun      bool
}

// AddInitFlags registers all init flags on the given command.
func AddInitFlags(cmd *cobra.Command) *InitFlags {
	f := cmd.Flags()
	flags := &InitFlags{}
	f.StringVar(&flags.Stack, "stack", "", "project stack: go or nodejs")
	f.StringVar(&flags.AppType, "app-type", "", "application type (see devopsgen list)")
	f.StringVar(&flags.CI, "ci", "github-actions", "CI provider")
	f.StringVar(&flags.Registry, "registry", "dockerhub", "container registry: dockerhub or ghcr")
	f.StringVar(&flags.Image, "image", "", "image repository as namespace/repository")
	f.IntVar(&flags.Port, "port", 0, "application port (1-65535)")
	f.StringVar(&flags.BuildCmd, "build-cmd", "", "build command")
	f.StringVar(&flags.BuildPath, "build-path", ".", "Go package path to build inside the container (default .)")
	f.StringVar(&flags.TestCmd, "test-cmd", "", "test command")
	f.StringVar(&flags.StartCmd, "start-cmd", "", "startup command")
	f.StringVar(&flags.Tag, "tag", "", "additional image tag alias (default latest)")
	f.BoolVar(&flags.K8s, "k8s", true, "generate Kubernetes manifests")
	f.BoolVar(&flags.Ingress, "ingress", false, "enable Kubernetes Ingress")
	f.StringVar(&flags.IngressHost, "ingress-host", "", "Ingress hostname (RFC 1123)")
	f.StringVar(&flags.HealthPath, "health-path", "", "HTTP health probe path")
	f.StringVar(&flags.OutputDir, "output-dir", ".", "output directory")
	f.BoolVar(&flags.Force, "force", false, "overwrite existing regular files")
	f.BoolVar(&flags.DryRun, "dry-run", false, "simulate without writing")
	return flags
}

// ToConfig converts parsed flags into a Config without requiring stack.
// Stack resolution (from flags or detection) happens in RunInit per SRS 4.2.
func (f *InitFlags) ToConfig() (config.Config, error) {
	cfg := config.Config{
		Stack:      config.StackType(f.Stack),
		AppType:    config.AppType(f.AppType),
		CI:         config.CIProvider(f.CI),
		Registry:   config.Registry(f.Registry),
		Image:      f.Image,
		Port:       f.Port,
		BuildCmd:   f.BuildCmd,
		BuildPath:  f.BuildPath,
		TestCmd:    f.TestCmd,
		StartCmd:   f.StartCmd,
		Tag:        f.Tag,
		OutputDir:  f.OutputDir,
		Generation: config.GenerationOptions{Force: f.Force, DryRun: f.DryRun},
	}
	cfg.K8s = config.K8sOptions{
		Enabled:     f.K8s,
		Ingress:     f.Ingress,
		IngressHost: f.IngressHost,
		HealthPath:  f.HealthPath,
	}

	// Validate CI, registry, image, port, tag, app-type, k8s options.
	// Stack is allowed to be empty here; it is resolved later.
	if cfg.Stack != "" && cfg.Stack != config.StackGo && cfg.Stack != config.StackNode {
		return cfg, fmt.Errorf("stack must be 'go' or 'nodejs'")
	}
	if cfg.CI != config.CIGitHubActions {
		return cfg, fmt.Errorf("ci must be 'github-actions'")
	}
	if cfg.Registry != config.RegistryDockerHub && cfg.Registry != config.RegistryGHCR {
		return cfg, fmt.Errorf("registry must be 'dockerhub' or 'ghcr'")
	}
	if strings.TrimSpace(cfg.Image) == "" {
		return cfg, fmt.Errorf("image is required")
	}
	if !config.ImageRe.MatchString(cfg.Image) {
		return cfg, fmt.Errorf("image must be 'namespace/repository' without host or tag")
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return cfg, fmt.Errorf("port must be in range 1-65535")
	}
	if cfg.AppType != "" && !config.ValidAppTypeForStack(cfg.Stack, cfg.AppType) {
		return cfg, fmt.Errorf("app-type is not supported for the selected stack")
	}
	if err := config.ValidateTag(cfg.Tag); err != nil {
		return cfg, fmt.Errorf("tag: %w", err)
	}
	if cfg.K8s.Ingress && cfg.K8s.IngressHost == "" {
		return cfg, fmt.Errorf("ingress-host is required when ingress is enabled")
	}
	if cfg.K8s.IngressHost != "" {
		if err := config.ValidateHostname(cfg.K8s.IngressHost); err != nil {
			return cfg, err
		}
	}
	if cfg.K8s.HealthPath != "" && !config.HealthPathRe.MatchString(cfg.K8s.HealthPath) {
		return cfg, fmt.Errorf("health-path must be a valid HTTP path")
	}
	if err := config.ValidateCommand(cfg.BuildCmd); err != nil {
		return cfg, fmt.Errorf("build-cmd: %w", err)
	}
	if err := config.ValidateCommand(cfg.TestCmd); err != nil {
		return cfg, fmt.Errorf("test-cmd: %w", err)
	}
	if err := config.ValidateCommand(cfg.StartCmd); err != nil {
		return cfg, fmt.Errorf("start-cmd: %w", err)
	}
	if err := config.ValidateOutputDir(cfg.OutputDir); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// ParsePort is a helper used by tests to exercise the flag binding path.
func ParsePort(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}

// readInitFlags collects the parsed init flag values from the command.
func readInitFlags(cmd *cobra.Command) (*InitFlags, error) {
	flags := &InitFlags{}
	f := cmd.Flags()
	flags.Stack, _ = f.GetString("stack")
	flags.AppType, _ = f.GetString("app-type")
	flags.CI, _ = f.GetString("ci")
	flags.Registry, _ = f.GetString("registry")
	flags.Image, _ = f.GetString("image")
	flags.Port, _ = f.GetInt("port")
	flags.BuildCmd, _ = f.GetString("build-cmd")
	flags.BuildPath, _ = f.GetString("build-path")
	flags.TestCmd, _ = f.GetString("test-cmd")
	flags.StartCmd, _ = f.GetString("start-cmd")
	flags.Tag, _ = f.GetString("tag")
	flags.K8s, _ = f.GetBool("k8s")
	flags.Ingress, _ = f.GetBool("ingress")
	flags.IngressHost, _ = f.GetString("ingress-host")
	flags.HealthPath, _ = f.GetString("health-path")
	flags.OutputDir, _ = f.GetString("output-dir")
	flags.Force, _ = f.GetBool("force")
	flags.DryRun, _ = f.GetBool("dry-run")
	return flags, nil
}
