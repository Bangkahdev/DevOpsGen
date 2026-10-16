// Package wizard provides the interactive TUI built on Bubble Tea.
//
// The wizard collects and validates configuration, surfaces detection
// candidates for confirmation, and lets the user review the summary
// before generation. It depends only on the config model and the core
// engine; it does not import Cobra.
package wizard

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"devopsgen.dev/devopsgen/internal/config"
	"devopsgen.dev/devopsgen/internal/detector"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
	StepRegistry
	StepImage
	StepK8s
	StepSummary
)

// errWizardAborted is returned when the user quits the wizard.
var errWizardAborted = errors.New("wizard aborted by user")

// Model is the Bubble Tea model for the wizard.
type Model struct {
	step      Step
	detect    detector.DetectionResult
	cfg       config.Config
	err       error
	choices   []string
	choiceIdx int
	input     textinput.Model
	width     int
	height    int
	quitting  bool
	detected  bool
}

// New creates a wizard Model, running stack detection first.
func New() Model {
	m := Model{
		step:    StepDetect,
		input:   textinput.New(),
		choices: []string{"go", "nodejs"},
		cfg: config.Config{
			CI:       config.CIGitHubActions,
			Registry: config.RegistryDockerHub,
			Tag:      config.DefaultTag,
			K8s:      config.K8sOptions{Enabled: true},
		},
	}
	m.input.Placeholder = "port"
	return m
}

// detectionDone is a message carrying the detection result.
type detectionDone struct {
	result detector.DetectionResult
	err    error
}

// runDetection performs stack detection and returns a detectionDone message.
func runDetection() tea.Msg {
	d := detector.New()
	result, err := d.Detect(".")
	return detectionDone{result: result, err: err}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return runDetection
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			return m.handleEnter()
		case "up", "k":
			if m.choiceIdx > 0 {
				m.choiceIdx--
			}
		case "down", "j":
			if m.choiceIdx < len(m.choices)-1 {
				m.choiceIdx++
			}
		}
	case keyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			return m.handleEnter()
		case "up", "k":
			if m.choiceIdx > 0 {
				m.choiceIdx--
			}
		case "down", "j":
			if m.choiceIdx < len(m.choices)-1 {
				m.choiceIdx++
			}
		}
	case detectionDone:
		m.detected = true
		m.detect = msg.result
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.err = nil
			res := msg.result
			if res.Stack != "" {
				m.cfg.Stack = config.StackType(res.Stack)
			}
			if len(res.AppCandidates) == 1 {
				m.cfg.AppType = config.AppType(res.AppCandidates[0])
			}
			m.cfg.BuildCmd = res.BuildCmd
			m.cfg.TestCmd = res.TestCmd
			m.cfg.StartCmd = res.StartCmd
			// Set choice index based on detected stack.
			for i, c := range m.choices {
				if c == string(m.cfg.Stack) {
					m.choiceIdx = i
					break
				}
			}
		}
		m.step = StepStack
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case StepStack:
		m.cfg.Stack = config.StackType(m.choices[m.choiceIdx])
		// Set app-type candidates based on stack.
		if m.cfg.Stack == config.StackGo {
			m.cfg.AppType = config.AppTypeBinary
			m.choices = []string{"binary"}
			m.choiceIdx = 0
			m.step = StepCommands
		} else {
			m.choices = []string{"express", "nextjs", "vite", "nestjs", "generic"}
			m.choiceIdx = 0
			m.step = StepAppType
		}
	case StepAppType:
		m.cfg.AppType = config.AppType(m.choices[m.choiceIdx])
		m.step = StepCommands
	case StepCommands:
		// Commands are pre-filled from detection; user can edit in summary.
		m.step = StepPort
		m.input.Placeholder = "port (1-65535)"
		m.input.CharLimit = 5
		m.input.SetValue("")
	case StepPort:
		if err := m.validatePort(m.input.Value()); err != nil {
			m.err = err
			return m, nil
		}
		if m.input.Value() != "" {
			var port int
			fmt.Sscanf(m.input.Value(), "%d", &port)
			m.cfg.Port = port
		}
		m.step = StepRegistry
		m.input.Placeholder = "dockerhub or ghcr"
		m.input.SetValue(string(m.cfg.Registry))
	case StepRegistry:
		reg := config.Registry(m.input.Value())
		if reg != config.RegistryDockerHub && reg != config.RegistryGHCR {
			m.err = fmt.Errorf("registry must be 'dockerhub' or 'ghcr'")
			return m, nil
		}
		m.cfg.Registry = reg
		m.step = StepImage
		m.input.Placeholder = "namespace/repository"
		m.input.SetValue("")
	case StepImage:
		if err := config.ValidateTag(""); err != nil {
			// Should not happen
		}
		if m.input.Value() != "" {
			if !config.ImageRe.MatchString(m.input.Value()) {
				m.err = fmt.Errorf("image must be 'namespace/repository'")
				return m, nil
			}
			m.cfg.Image = m.input.Value()
		}
		m.step = StepK8s
		m.input.Placeholder = "e.g. /health (optional)"
		m.input.SetValue(m.cfg.K8s.HealthPath)
	case StepK8s:
		if m.input.Value() != "" {
			m.cfg.K8s.HealthPath = m.input.Value()
		}
		m.step = StepSummary
	case StepSummary:
		// Validate final config before returning.
		if err := config.ValidateConfig(m.cfg); err != nil {
			m.err = fmt.Errorf("configuration invalid: %w", err)
			return m, nil
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) validatePort(s string) error {
	if s == "" {
		return nil // port will be prompted again or required later
	}
	var port int
	if _, err := fmt.Sscanf(s, "%d", &port); err != nil {
		return fmt.Errorf("port must be a number between 1 and 65535")
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}

// View implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		return ""
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render("DevOpsGen Wizard"))
	b.WriteString("\n\n")

	switch m.step {
	case StepDetect:
		b.WriteString("Detecting project stack...\n")
		if m.err != nil {
			b.WriteString(errorStyle.Render("warning: "+m.err.Error()) + "\n")
		}
		if m.detected {
			b.WriteString(fmt.Sprintf("stack: %s\n", m.detect.Stack))
			b.WriteString(fmt.Sprintf("indicators: %s\n", strings.Join(m.detect.Indicators, ", ")))
			if len(m.detect.Warnings) > 0 {
				b.WriteString(fmt.Sprintf("warnings: %s\n", strings.Join(m.detect.Warnings, "; ")))
			}
			if len(m.detect.AppCandidates) > 0 {
				b.WriteString(fmt.Sprintf("app-type candidates: %s\n", strings.Join(m.detect.AppCandidates, ", ")))
			}
		}
		b.WriteString("\nPress enter to continue.\n")
	case StepStack:
		b.WriteString("Select project stack:\n")
		for i, c := range m.choices {
			prefix := "  "
			if i == m.choiceIdx {
				prefix = "> "
			}
			b.WriteString(prefix + c + "\n")
		}
		b.WriteString("\n(use ↑/↓ to navigate, Enter to select)\n")
	case StepAppType:
		b.WriteString(fmt.Sprintf("Select app-type for %s:\n", m.cfg.Stack))
		for i, c := range m.choices {
			prefix := "  "
			if i == m.choiceIdx {
				prefix = "> "
			}
			b.WriteString(prefix + c + "\n")
		}
		b.WriteString("\n(use ↑/↓ to navigate, Enter to select)\n")
	case StepCommands:
		b.WriteString("Detected commands:\n")
		if m.detect.BuildCmd != "" {
			b.WriteString(fmt.Sprintf("  build-cmd: %s\n", m.detect.BuildCmd))
			m.cfg.BuildCmd = m.detect.BuildCmd
		}
		if m.detect.TestCmd != "" {
			b.WriteString(fmt.Sprintf("  test-cmd: %s\n", m.detect.TestCmd))
			m.cfg.TestCmd = m.detect.TestCmd
		}
		if m.detect.StartCmd != "" {
			b.WriteString(fmt.Sprintf("  start-cmd: %s\n", m.detect.StartCmd))
			m.cfg.StartCmd = m.detect.StartCmd
		}
		if m.detect.BuildCmd == "" && m.detect.TestCmd == "" && m.detect.StartCmd == "" {
			b.WriteString("  (none detected; you can set them later)\n")
		}
		b.WriteString("\nPress enter to continue.\n")
	case StepPort:
		b.WriteString("Enter application port (1-65535):\n")
		b.WriteString(m.input.View() + "\n")
		if m.err != nil {
			b.WriteString("\n" + errorStyle.Render(m.err.Error()) + "\n")
		}
	case StepRegistry:
		b.WriteString("Select container registry:\n")
		b.WriteString(m.input.View() + "\n")
		if m.err != nil {
			b.WriteString("\n" + errorStyle.Render(m.err.Error()) + "\n")
		}
	case StepImage:
		b.WriteString("Enter image as namespace/repository:\n")
		b.WriteString(m.input.View() + "\n")
		if m.err != nil {
			b.WriteString("\n" + errorStyle.Render(m.err.Error()) + "\n")
		}
	case StepK8s:
		b.WriteString("Enter health check path (optional, e.g. /health):\n")
		b.WriteString(m.input.View() + "\n")
		if m.err != nil {
			b.WriteString("\n" + errorStyle.Render(m.err.Error()) + "\n")
		}
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
		if m.cfg.K8s.Ingress {
			b.WriteString(fmt.Sprintf("  ingress:      %v\n", m.cfg.K8s.Ingress))
			b.WriteString(fmt.Sprintf("  ingress-host: %s\n", m.cfg.K8s.IngressHost))
		}
		if m.cfg.K8s.HealthPath != "" {
			b.WriteString(fmt.Sprintf("  health-path:  %s\n", m.cfg.K8s.HealthPath))
		}
		b.WriteString("\nPress enter to generate, or q to abort.\n")
	}
	if m.err != nil && m.step != StepDetect {
		b.WriteString("\n" + errorStyle.Render(m.err.Error()))
	}
	return b.String()
}

// Run runs the full Bubble Tea wizard loop and returns a validated config.
// If the terminal is not interactive, Run returns an error so the caller
// can fall back to flag-based (headless) mode.
func Run() (config.Config, error) {
	if !interactiveFn() {
		return config.Config{}, fmt.Errorf("wizard: terminal is not interactive; use flags instead")
	}
	m := New()
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return config.Config{}, err
	}
	final := finalModel.(Model)
	if final.quitting {
		return config.Config{}, errWizardAborted
	}
	return final.cfg, nil
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

// keyMsg is a test-only message type simulating a key press.
// It satisfies tea.Msg.
type keyMsg string

func (k keyMsg) String() string { return string(k) }

// Compile-time check that keyMsg satisfies tea.Msg.
var _ tea.Msg = keyMsg("")

// titleStyle styles the wizard header.
var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")).Padding(0, 1)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)
