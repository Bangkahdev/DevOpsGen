// Package cli exposes the Cobra command tree for DevOpsGen.
//
// The CLI is responsible for parsing flags, mapping command failures to the
// exit codes defined in SRS FR-16, and delegating to the core engine.
// This package must not import the TUI (Bubble Tea) or any filesystem
// implementation directly; it only orchestrates.
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"devopsgen.dev/devopsgen/internal/config"
	"devopsgen.dev/devopsgen/internal/detector"
	"devopsgen.dev/devopsgen/internal/output"
	"devopsgen.dev/devopsgen/internal/templates"
	"devopsgen.dev/devopsgen/internal/validator"
	"devopsgen.dev/devopsgen/internal/wizard"
)

// NewRootCommand builds the root command and registers the MVP subcommands.
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "devopsgen",
		Short: "Developer-oriented DevOps scaffolding tool",
		Long: `DevOpsGen generates deterministic Dockerfile, GitHub Actions,
and Kubernetes manifests from a small, explicit configuration.

It supports interactive (TUI) and non-interactive (headless) modes.
All core operations can be executed without the TUI.`,
		SilenceUsage: true,
	}

	root.AddCommand(
		newInitCommand(),
		newDetectCommand(),
		newValidateCommand(),
		newListCommand(),
		newVersionCommand(),
	)

	return root
}

// mapCommandError translates command-level failures into exit codes.
//
// Exit codes follow SRS FR-16:
//
//	0 success
//	1 general / I/O / rollback
//	2 invalid input, flag, or configuration
//	3 file conflict without explicit overwrite
//	4 artifact validation failure
//	5 ambiguous or unsupported stack
func mapCommandError(err error) int {
	if err == nil {
		return 0
	}
	var exitErr exitCoder
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	if errors.Is(err, errConflict) {
		return 3
	}
	if errors.Is(err, errValidation) {
		return 4
	}
	if errors.Is(err, errAmbiguousStack) {
		return 5
	}
	return 1
}

// MapExitCode is the exported form used by main to translate errors.
func MapExitCode(err error) int {
	return mapCommandError(err)
}

// exitCoder allows domain errors to declare an explicit exit code.
type exitCoder interface {
	ExitCode() int
}

// Domain error sentinels. Concrete types may wrap these with context.
var (
	errConflict       = errors.New("file conflict without overwrite consent")
	errValidation     = errors.New("artifact validation failed")
	errAmbiguousStack = errors.New("ambiguous or unsupported stack")
)

// newInitCommand wires the init command.
func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate DevOps configuration for the project",
		Long: `Initialize DevOps configuration.

In an interactive terminal with no complete configuration, the TUI wizard
is launched. When all required parameters are supplied via flags, the
process runs headless and never waits for stdin.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags, err := readInitFlags(cmd)
			if err != nil {
				return newExitError(2, err)
			}
			// In an interactive terminal with incomplete configuration,
			// launch the TUI wizard before generation.
			if flags.Stack == "" && isInteractive() {
				wcfg, werr := wizard.Run()
				if werr != nil {
					return newExitError(2, werr)
				}
				flags.Stack = string(wcfg.Stack)
				flags.AppType = string(wcfg.AppType)
				flags.BuildCmd = wcfg.BuildCmd
				flags.TestCmd = wcfg.TestCmd
				flags.StartCmd = wcfg.StartCmd
				if wcfg.Image != "" {
					flags.Image = wcfg.Image
				}
				if wcfg.Port != 0 {
					flags.Port = wcfg.Port
				}
				if wcfg.Tag != "" {
					flags.Tag = wcfg.Tag
				}
				if wcfg.Registry != "" {
					flags.Registry = string(wcfg.Registry)
				}
				if wcfg.CI != "" {
					flags.CI = string(wcfg.CI)
				}
				// Merge flag-based k8s settings into wizard config.
				wcfg.K8s.Enabled = flags.K8s
				wcfg.K8s.Ingress = flags.Ingress
				wcfg.K8s.IngressHost = flags.IngressHost
				wcfg.K8s.HealthPath = flags.HealthPath
				wcfg.OutputDir = flags.OutputDir
				wcfg.Generation = config.GenerationOptions{Force: flags.Force, DryRun: flags.DryRun}
				// Read back merged k8s options into flags for RunInit.
				flags.K8s = wcfg.K8s.Enabled
				flags.Ingress = wcfg.K8s.Ingress
				flags.IngressHost = wcfg.K8s.IngressHost
				flags.HealthPath = wcfg.K8s.HealthPath
			}
			return RunInit(flags)
		},
	}
	_ = AddInitFlags(cmd)
	return cmd
}

// isInteractive reports whether stdin is a terminal.
func isInteractive() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func newDetectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "detect",
		Short: "Detect candidate project stack",
		Long: `Detect candidate project stack without modifying the project.

Reports candidate stack, detected indicators, candidate app-type, build /
test / start commands, lockfile or package manager, and unresolved
ambiguities. Detection is read-only.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDetect(cmd.OutOrStdout())
		},
	}
}

func runDetect(out io.Writer) error {
	root, _ := os.Getwd()
	d := detector.New()
	r, err := d.Detect(root)
	if err != nil {
		return newExitError(5, fmt.Errorf("detect: %w", err))
	}
	fmt.Fprintf(out, "stack: %s\n", r.Stack)
	fmt.Fprintf(out, "indicators: %s\n", strings.Join(r.Indicators, ", "))
	if len(r.Warnings) > 0 {
		fmt.Fprintf(out, "warnings: %s\n", strings.Join(r.Warnings, "; "))
	}
	if len(r.AppCandidates) > 0 {
		fmt.Fprintf(out, "app-type candidates: %s\n", strings.Join(r.AppCandidates, ", "))
	}
	if r.BuildCmd != "" {
		fmt.Fprintf(out, "build-cmd: %s\n", r.BuildCmd)
	}
	if r.TestCmd != "" {
		fmt.Fprintf(out, "test-cmd: %s\n", r.TestCmd)
	}
	if r.StartCmd != "" {
		fmt.Fprintf(out, "start-cmd: %s\n", r.StartCmd)
	}
	if r.Lockfile != "" {
		fmt.Fprintf(out, "lockfile: %s\n", r.Lockfile)
	}
	if r.PackageManager != "" {
		fmt.Fprintf(out, "package-manager: %s\n", r.PackageManager)
	}
	if r.IsAmbiguous() {
		return newExitError(5, fmt.Errorf("ambiguities: %s", strings.Join(r.Ambiguities, "; ")))
	}
	return nil
}

func newValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate generated artifacts",
		Long: `Validate generated artifacts in the working directory (or --dir).

Checks required files, YAML syntax, Kubernetes field consistency,
selector/label alignment, image references, Service/Ingress relations,
GitHub Actions structure, and port/health probe consistency.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			return runValidate(dir, cmd.OutOrStderr())
		},
	}
	_ = cmd.Flags().String("dir", ".", "directory containing artifacts to validate")
	return cmd
}

func runValidate(dir string, out io.Writer) error {
	arts, err := collectArtifacts(dir)
	if err != nil {
		return newExitError(1, fmt.Errorf("read dir: %w", err))
	}
	if len(arts) == 0 {
		return newExitError(4, fmt.Errorf("no artifacts found in %s", dir))
	}
	// Check for required MVP artifacts.
	required := map[string]bool{"Dockerfile": false, ".github/workflows/ci.yml": false}
	for _, a := range arts {
		if a.Path == "Dockerfile" {
			required["Dockerfile"] = true
		}
		if a.Path == ".github/workflows/ci.yml" || strings.HasSuffix(a.Path, "ci.yml") {
			required[".github/workflows/ci.yml"] = true
		}
	}
	var missing []string
	for k, v := range required {
		if !v {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return newExitError(4, fmt.Errorf("missing required artifacts: %s", strings.Join(missing, ", ")))
	}
	r, err := validator.New().Validate(arts)
	if err != nil {
		return newExitError(4, fmt.Errorf("validate: %w", err))
	}
	if r != validator.ResultValid {
		return newExitError(4, fmt.Errorf("validation failed: %s", r))
	}
	fmt.Fprintf(out, "Internal validation passed\n")
	return nil
}

// collectArtifacts recursively finds supported artifact files in dir.
func collectArtifacts(dir string) ([]output.Artifact, error) {
	var arts []output.Artifact
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip hidden directories and common non-artifact dirs.
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == ".devopsgen" {
				return filepath.SkipDir
			}
			return nil
		}
		if !isArtifact(d.Name()) {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		arts = append(arts, output.Artifact{Path: filepath.ToSlash(rel), Content: b})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return arts, nil
}

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List supported stacks, app types, registries, and templates",
		Long: `List supported stacks, app types, registries, CI providers, and
deployment templates bundled in the running binary. The list reflects
actual bundled capability, not a future roadmap.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "Stacks:\n  go, nodejs\n")
			fmt.Fprintf(cmd.OutOrStdout(), "App types:\n  binary (Go), express, nextjs, vite, nestjs, generic (Node.js)\n")
			fmt.Fprintf(cmd.OutOrStdout(), "CI providers:\n  github-actions\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Registries:\n  dockerhub, ghcr\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Templates: docker, github-actions, kubernetes\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Template version: %s\n", templates.Version)
			return nil
		},
	}
}

func newVersionCommand() *cobra.Command {
	return NewVersionCommand()
}

func isArtifact(name string) bool {
	return strings.HasSuffix(name, "Dockerfile") ||
		strings.HasSuffix(name, ".yaml") ||
		strings.HasSuffix(name, ".yml")
}
