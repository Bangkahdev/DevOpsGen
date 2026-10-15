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

	"github.com/spf13/cobra"
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

// exitCoder allows domain errors to declare an explicit exit code.
type exitCoder interface {
	ExitCode() int
}

// Domain error sentinels. Concrete types may wrap these with context.
var (
	errConflict       = errors.New("file conflict without overwrite consent")
	errValidation    = errors.New("artifact validation failed")
	errAmbiguousStack = errors.New("ambiguous or unsupported stack")
)

// newInitCommand wires the init command. The implementation lives in a
// separate file so this file stays focused on the command tree.
func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Generate DevOps configuration for the project",
		Long: `Initialize DevOps configuration.

In an interactive terminal with no complete configuration, the TUI wizard
is launched. When all required parameters are supplied via flags, the
process runs headless and never waits for stdin.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO (Phase 6): connect to core engine.
			_, _ = fmt.Fprintf(cmd.OutOrStderr(), "devopsgen init: not yet implemented (Phase 6)\n")
			return nil
		},
	}
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
			_, _ = fmt.Fprintf(cmd.OutOrStderr(), "devopsgen detect: not yet implemented (Phase 5)\n")
			return nil
		},
	}
}

func newValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate generated artifacts",
		Long: `Validate generated artifacts in the working directory (or --dir).

Checks required files, YAML syntax, Kubernetes field consistency,
selector/label alignment, image references, Service/Ingress relations,
GitHub Actions structure, and port/health probe consistency.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _ = fmt.Fprintf(cmd.OutOrStderr(), "devopsgen validate: not yet implemented (Phase 5)\n")
			return nil
		},
	}
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
			_, _ = fmt.Fprintf(cmd.OutOrStderr(), "devopsgen list: not yet implemented (Phase 6)\n")
			return nil
		},
	}
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long: `Print the application version and the bundled template version.

On the MVP the template version tracks the binary version. Development
builds may report "dev" when a release version is unavailable.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "devopsgen dev (template: dev)\n")
			return nil
		},
	}
}