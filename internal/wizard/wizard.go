// Package wizard provides the interactive TUI built on Bubble Tea.
//
// The wizard collects and validates configuration, surfaces detection
// candidates for confirmation, and lets the user review the summary
// before generation. It depends only on the config model and the core
// engine; it does not import Cobra.
package wizard

import (
	"fmt"
	"os"
	"strings"

	"devopsgen.dev/devopsgen/internal/config"
	"devopsgen.dev/devopsgen/internal/detector"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// Step is a stage in the wizard flow.
type Step int

const (
	StepDetect Step = iota
	StepStack
	StepAppType
	StepCommands
	StepPort
	StepImage
	StepK8s
	StepSummary
)

// Model is the Bubble Tea model for the wizard.
type Model struct {
	step    Step
	detect  detector.DetectionResult
	cfg     config.Config
	err     error
	choices []string
	input   textinput.Model
	width   int
	height  int
}

// New creates a wizard Model, running stack detection first.
func New() Model {
	m := Model{
		step:    StepDetect,
		input:   textinput.New(),
		choices: []string{"go", "nodejs"},
	}
	m.input.Placeholder = "port"
	m.input.CharLimit = 5
	return m
}

// detectionDone is a message carrying the detection result.
type detectionDone struct {
	result detector.DetectionResult
	err    error
}

// Run runs stack detection and returns a pre-filled configuration.
// If the terminal is not interactive, Run returns an error so the
// caller can fall back to flag-based (headless) mode.
func Run() (config.Config, error) {
	if !interactiveFn() {
		return config.Config{}, fmt.Errorf("wizard: terminal is not interactive; use flags instead")
	}
	d := detector.New()
	result, err := d.Detect(".")
	if err != nil {
		// Detection is advisory; the wizard still opens so the user can
		// supply the stack explicitly.
		fmt.Fprintf(os.Stderr, "wizard: stack detection failed: %v\n", err)
	}
	cfg := config.Config{
		CI:       config.CIGitHubActions,
		Registry: config.RegistryDockerHub,
		Tag:      config.DefaultTag,
		K8s:      config.K8sOptions{Enabled: true},
	}
	if err == nil {
		if result.Stack != "" {
			cfg.Stack = config.StackType(result.Stack)
		}
		if len(result.AppCandidates) == 1 {
			cfg.AppType = config.AppType(result.AppCandidates[0])
		}
		cfg.BuildCmd = result.BuildCmd
		cfg.TestCmd = result.TestCmd
		cfg.StartCmd = result.StartCmd
	}
	return cfg, nil
}

// isInteractive reports whether stdin is a terminal.
func isInteractive() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// interactiveFn is the function used to check interactivity. It is a
// package-level variable so tests can override it.
var interactiveFn = isInteractive

// view renders the current wizard step.
func (m Model) view() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("DevOpsGen Wizard"))
	b.WriteString("\n\n")
	switch m.step {
	case StepDetect:
		b.WriteString("Detecting project stack...\n")
		if m.err != nil {
			b.WriteString(errorStyle.Render("warning: " + m.err.Error()))
		} else {
			b.WriteString(fmt.Sprintf("stack: %s\n", m.detect.Stack))
			b.WriteString(fmt.Sprintf("indicators: %s\n", strings.Join(m.detect.Indicators, ", ")))
			if len(m.detect.AppCandidates) > 0 {
				b.WriteString(fmt.Sprintf("app-type: %s\n", strings.Join(m.detect.AppCandidates, ", ")))
			}
		}
	case StepStack:
		b.WriteString("Select project stack:\n")
		for _, c := range m.choices {
			prefix := "  "
			if string(m.cfg.Stack) == c {
				prefix = "> "
			}
			b.WriteString(prefix + c + "\n")
		}
	case StepPort:
		b.WriteString("Enter application port (1-65535):\n")
		b.WriteString(m.input.View())
	case StepImage:
		b.WriteString("Enter image as namespace/repository:\n")
		b.WriteString(m.input.View())
	case StepSummary:
		b.WriteString("Review configuration:\n")
		b.WriteString(fmt.Sprintf("  stack:        %s\n", m.cfg.Stack))
		b.WriteString(fmt.Sprintf("  app-type:     %s\n", m.cfg.AppType))
		b.WriteString(fmt.Sprintf("  ci:           %s\n", m.cfg.CI))
		b.WriteString(fmt.Sprintf("  registry:     %s\n", m.cfg.Registry))
		b.WriteString(fmt.Sprintf("  image:        %s\n", m.cfg.Image))
		b.WriteString(fmt.Sprintf("  port:         %d\n", m.cfg.Port))
		b.WriteString(fmt.Sprintf("  build-cmd:    %s\n", m.cfg.BuildCmd))
		b.WriteString(fmt.Sprintf("  test-cmd:     %s\n", m.cfg.TestCmd))
		b.WriteString(fmt.Sprintf("  start-cmd:    %s\n", m.cfg.StartCmd))
		b.WriteString(fmt.Sprintf("  tag:          %s\n", m.cfg.Tag))
		b.WriteString(fmt.Sprintf("  k8s:          %v\n", m.cfg.K8s.Enabled))
		b.WriteString("\nPress enter to generate.\n")
	}
	if m.err != nil && m.step != StepDetect {
		b.WriteString("\n" + errorStyle.Render(m.err.Error()))
	}
	return b.String()
}

// titleStyle styles the wizard header.
var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")).Padding(0, 1)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

