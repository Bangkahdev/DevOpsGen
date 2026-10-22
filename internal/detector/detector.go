// Package detector performs read-only stack detection from project indicators.
//
// Detection must never modify the project or execute project scripts. It
// reports candidates, indicators, warnings, and unresolved ambiguities.
package detector

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrNoIndicators is returned when no supported stack indicator exists in
// the inspected root. Callers may use errors.Is to convert it to the
// "unsupported stack" exit code instead of a general I/O failure code.
var ErrNoIndicators = errors.New("detector: no supported stack indicators found")

// DetectionResult summarizes what the detector found in a project root.
type DetectionResult struct {
	Stack          string
	Indicators     []string
	AppCandidates  []string
	BuildCmd       string
	TestCmd        string
	StartCmd       string
	Lockfile       string
	PackageManager string
	Warnings       []string
	Ambiguities    []string
}

// Detector is the interface for stack detection implementations.
type Detector interface {
	Detect(root string) (DetectionResult, error)
}

// detectGo checks for a Go project via go.mod.
func detectGo(root string) (DetectionResult, bool, error) {
	r := DetectionResult{Stack: "go"}
	p := filepath.Join(root, "go.mod")
	b, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return r, false, nil
	}
	if err != nil {
		return r, false, fmt.Errorf("read go.mod: %w", err)
	}
	r.Indicators = append(r.Indicators, "go.mod")
	r.AppCandidates = append(r.AppCandidates, "binary")
	if strings.Contains(string(b), "module ") {
		r.Indicators = append(r.Indicators, "go.mod module directive")
	}
	return r, true, nil
}

// detectNode checks for a Node.js project via package.json.
func detectNode(root string) (DetectionResult, bool, error) {
	r := DetectionResult{Stack: "nodejs"}
	p := filepath.Join(root, "package.json")
	b, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return r, false, nil
	}
	if err != nil {
		return r, false, fmt.Errorf("read package.json: %w", err)
	}
	r.Indicators = append(r.Indicators, "package.json")

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		Scripts         map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(b, &pkg); err != nil {
		r.Warnings = append(r.Warnings, "package.json is not valid JSON")
		return r, true, nil
	}

	// Scripts.
	if pkg.Scripts["build"] != "" {
		r.BuildCmd = pkg.Scripts["build"]
		r.Indicators = append(r.Indicators, "scripts.build")
	}
	if pkg.Scripts["test"] != "" {
		r.TestCmd = pkg.Scripts["test"]
		r.Indicators = append(r.Indicators, "scripts.test")
	}
	if pkg.Scripts["start"] != "" {
		r.StartCmd = pkg.Scripts["start"]
		r.Indicators = append(r.Indicators, "scripts.start")
	}

	// Framework detection from dependencies.
	allDeps := map[string]string{}
	for k, v := range pkg.Dependencies {
		allDeps[k] = v
	}
	for k, v := range pkg.DevDependencies {
		allDeps[k] = v
	}
	var frameworks []string
	for name := range allDeps {
		switch name {
		case "express":
			frameworks = append(frameworks, "express")
		case "next":
			frameworks = append(frameworks, "nextjs")
		case "vite":
			frameworks = append(frameworks, "vite")
		case "@nestjs/core":
			frameworks = append(frameworks, "nestjs")
		}
	}
	if len(frameworks) > 1 {
		r.Ambiguities = append(r.Ambiguities, "multiple frameworks detected: "+strings.Join(frameworks, ", "))
	} else if len(frameworks) == 1 {
		r.AppCandidates = append(r.AppCandidates, frameworks[0])
		r.Indicators = append(r.Indicators, "dependency "+frameworks[0])
	} else {
		r.AppCandidates = append(r.AppCandidates, "generic")
	}

	// Lockfile / package manager. Absence is a warning, not ambiguity.
	for _, f := range []string{"package-lock.json", "pnpm-lock.yaml", "yarn.lock", "npm-shrinkwrap.json"} {
		if _, err := os.Stat(filepath.Join(root, f)); err == nil {
			r.Lockfile = f
			switch f {
			case "package-lock.json", "npm-shrinkwrap.json":
				r.PackageManager = "npm"
			case "pnpm-lock.yaml":
				r.PackageManager = "pnpm"
			case "yarn.lock":
				r.PackageManager = "yarn"
			}
			r.Indicators = append(r.Indicators, f)
			break
		}
	}
	if r.Lockfile == "" {
		r.Warnings = append(r.Warnings, "no lockfile detected; dependency installation may not be reproducible")
	}
	return r, true, nil
}

// engine implements Detector.
type engine struct{}

// New returns a Detector backed by project indicators.
func New() Detector {
	return &engine{}
}

// Detect runs all stack detectors against root and reports the result.
// Only Go and Node.js are in MVP scope.
func (e *engine) Detect(root string) (DetectionResult, error) {
	var results []DetectionResult
	goRes, goOk, err := detectGo(root)
	if err != nil {
		return DetectionResult{}, err
	}
	if goOk {
		results = append(results, goRes)
	}
	nodeRes, nodeOk, err := detectNode(root)
	if err != nil {
		return DetectionResult{}, err
	}
	if nodeOk {
		results = append(results, nodeRes)
	}
	if len(results) == 0 {
		return DetectionResult{}, ErrNoIndicators
	}
	if len(results) > 1 {
		// Ambiguous: multiple stacks detected.
		var amb []string
		for _, r := range results {
			amb = append(amb, r.Stack+" ("+strings.Join(r.Indicators, ", ")+")")
		}
		merged := results[0]
		merged.Ambiguities = append(merged.Ambiguities, "multiple stacks detected: "+strings.Join(amb, "; "))
		return merged, nil
	}
	return results[0], nil
}

// IsAmbiguous reports whether the result has unresolved ambiguities.
func (r DetectionResult) IsAmbiguous() bool {
	return len(r.Ambiguities) > 0
}

// HasStart reports whether a startup command was detected.
func (r DetectionResult) HasStart() bool {
	return (r.Stack == "nodejs") && r.StartCmd != ""
}

// HasBuild reports whether a build command was detected.
func (r DetectionResult) HasBuild() bool {
	return r.BuildCmd != ""
}

// HasTest reports whether a test command was detected.
func (r DetectionResult) HasTest() bool {
	return r.TestCmd != ""
}
