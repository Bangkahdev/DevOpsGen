// Package cli wires the init command to the core engine.
//
// It composes config validation, template rendering, filesystem planning,
// and commit into a single deterministic flow per SRS 6.6.
package cli

import (
	"fmt"
	"os"

	"devopsgen.dev/devopsgen/internal/filesystem"
	"devopsgen.dev/devopsgen/internal/generator"
)

// exitError wraps an error with an explicit exit code.
type exitError struct {
	code int
	err  error
}

func newExitError(code int, err error) *exitError {
	return &exitError{code: code, err: err}
}

func (e *exitError) Error() string { return e.err.Error() }
func (e *exitError) ExitCode() int  { return e.code }

// RunInit executes the full generation flow for the init command.
func RunInit(flags *InitFlags) error {
	cfg, err := flags.ToConfig()
	if err != nil {
		return newExitError(2, fmt.Errorf("config: %w", err))
	}

	// Stage 4: render artifacts in memory.
	arts, err := generator.New().Generate(cfg)
	if err != nil {
		return newExitError(4, fmt.Errorf("generate: %w", err))
	}

	// Stage 5: filesystem plan and conflict check.
	fs := &filesystem.RealFS{Root: cfg.OutputDir}
	artifacts := map[string][]byte{}
	for _, a := range arts {
		artifacts[a.Path] = a.Content
	}
	plan, err := fs.PlanBuild(artifacts)
	if err != nil {
		return newExitError(2, fmt.Errorf("plan: %w", err))
	}

	// Conflict without --force -> exit 3.
	if len(plan.Conflicts) > 0 && !cfg.Generation.Force {
		return newExitError(3, fmt.Errorf("conflict: %d file(s) would be overwritten without --force", len(plan.Conflicts)))
	}

	// Dry-run stops here and reports the plan.
	if cfg.Generation.DryRun {
		fmt.Fprintf(os.Stderr, "dry-run: would create %d file(s), replace %d file(s)\n",
			len(plan.Create), len(plan.Replace))
		for _, p := range plan.Create {
			fmt.Fprintf(os.Stderr, "  create %s\n", p)
		}
		for _, p := range plan.Replace {
			fmt.Fprintf(os.Stderr, "  replace %s\n", p)
		}
		return nil
	}

	// Stage 6: backup and commit.
	if err := fs.Commit(plan); err != nil {
		return newExitError(1, fmt.Errorf("commit: %w", err))
	}

	fmt.Fprintf(os.Stderr, "created %d file(s), replaced %d file(s) in %s\n",
		len(plan.Create), len(plan.Replace), cfg.OutputDir)
	return nil
}