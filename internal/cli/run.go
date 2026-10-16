// Package cli wires the init command to the core engine.
//
// It composes config validation, template rendering, filesystem planning,
// and commit into a single deterministic flow per SRS 6.6.
package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"devopsgen.dev/devopsgen/internal/config"
	"devopsgen.dev/devopsgen/internal/detector"
	"devopsgen.dev/devopsgen/internal/filesystem"
	"devopsgen.dev/devopsgen/internal/generator"
	"devopsgen.dev/devopsgen/internal/validator"
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
func (e *exitError) ExitCode() int { return e.code }

// RunInit executes the full generation flow for the init command.
//
// Stage order (SRS 6.6, FR-16):
//  1. Parsing flags and validasi nilai eksplisit          -> exit 2
//  2. Deteksi stack bila diperlukan                       -> exit 5
//  3. Penyelesaian + validasi config lengkap              -> exit 2
//  4. Rendering seluruh artefak di memori                 -> (error -> 4)
//  5. Validasi seluruh output                             -> exit 4
//  6. Validasi path + rencana + konflik                   -> exit 2/3
//  7. Backup + commit filesystem                          -> exit 1
func RunInit(flags *InitFlags) error {
	// Stage 1: flag-based config (stack may be empty; resolved later).
	cfg, err := flags.ToConfig()
	if err != nil {
		return newExitError(2, fmt.Errorf("config: %w", err))
	}

	// Stage 2: stack detection when not provided via flags.
	if cfg.Stack == "" {
		if !isInteractive() {
			// Non-interactive: cannot prompt, fail with actionable error.
			return newExitError(2, errors.New("stack is required; provide --stack or run in an interactive terminal"))
		}
		d := detector.New()
		result, err := d.Detect(".")
		if err != nil {
			return newExitError(2, fmt.Errorf("stack detection failed: %w", err))
		}
		if result.IsAmbiguous() {
			return newExitError(5, fmt.Errorf("ambiguous stack: %s", strings.Join(result.Ambiguities, "; ")))
		}
		cfg.Stack = config.StackType(result.Stack)
		if len(result.AppCandidates) == 1 {
			cfg.AppType = config.AppType(result.AppCandidates[0])
		}
		cfg.BuildCmd = result.BuildCmd
		cfg.TestCmd = result.TestCmd
		cfg.StartCmd = result.StartCmd
	}

	// Stage 3: full configuration validation after detection.
	if err := config.ValidateConfig(cfg); err != nil {
		return newExitError(2, fmt.Errorf("config: %w", err))
	}

	// Stage 4: render artifacts in memory.
	arts, err := generator.New().Generate(cfg)
	if err != nil {
		return newExitError(4, fmt.Errorf("generate: %w", err))
	}

	// Stage 5: validate all output artifacts.
	result, err := validator.New().Validate(arts)
	if err != nil {
		return newExitError(4, fmt.Errorf("validate: %w", err))
	}
	if result != validator.ResultValid {
		return newExitError(4, fmt.Errorf("validate: result %s", result))
	}

	// Stage 6: filesystem plan and conflict check.
	fs := &filesystem.RealFS{Root: cfg.OutputDir}
	artifactMap := map[string][]byte{}
	for _, a := range arts {
		artifactMap[a.Path] = a.Content
	}
	plan, err := fs.PlanBuild(artifactMap)
	if err != nil {
		return newExitError(2, fmt.Errorf("plan: %w", err))
	}

	// Dry-run stops here and reports the plan. Never writes.
	if cfg.Generation.DryRun {
		fmt.Fprintf(os.Stderr, "dry-run: would create %d file(s), replace %d file(s)\n",
			len(plan.Create), len(plan.Replace))
		for _, p := range plan.Create {
			fmt.Fprintf(os.Stderr, "  create %s\n", p)
		}
		for _, p := range plan.Replace {
			fmt.Fprintf(os.Stderr, "  replace %s\n", p)
		}
		for _, p := range plan.Conflicts {
			fmt.Fprintf(os.Stderr, "  conflict %s (would be replaced)\n", p)
		}
		return nil
	}

	// Conflict without --force -> exit 3.
	if len(plan.Conflicts) > 0 && !cfg.Generation.Force {
		return newExitError(3, fmt.Errorf("conflict: %d file(s) would be overwritten without --force", len(plan.Conflicts)))
	}

	// Stage 7: backup and commit.
	if err := fs.Commit(plan); err != nil {
		return newExitError(1, fmt.Errorf("commit: %w", err))
	}

	fmt.Fprintf(os.Stderr, "created %d file(s), replaced %d file(s) in %s\n",
		len(plan.Create), len(plan.Replace), cfg.OutputDir)
	return nil
}

// ResolveConfig applies the SRS 4.2 resolution priority:
// flags -> detection -> template defaults -> error.
func ResolveConfig(cfg config.Config, det detector.Detector) (config.Config, error) {
	if cfg.Stack == "" {
		result, err := det.Detect(".")
		if err != nil {
			return cfg, err
		}
		if result.IsAmbiguous() {
			return cfg, errors.New("ambiguous stack: " + strings.Join(result.Ambiguities, "; "))
		}
		cfg.Stack = config.StackType(result.Stack)
		if len(result.AppCandidates) == 1 {
			cfg.AppType = config.AppType(result.AppCandidates[0])
		}
		cfg.BuildCmd = result.BuildCmd
		cfg.TestCmd = result.TestCmd
		cfg.StartCmd = result.StartCmd
	}
	if err := config.ValidateConfig(cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
