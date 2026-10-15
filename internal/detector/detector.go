// Package detector performs read-only stack detection from project indicators.
//
// Detection must never modify the project or execute project scripts. It
// reports candidates, indicators, and unresolved ambiguities.
package detector

import "devopsgen.dev/devopsgen/internal/config"

// DetectionResult summarizes what the detector found in a project root.
type DetectionResult struct {
	Stack       config.StackType
	Indicators  []string
	AppCandidates []config.AppType
	BuildCmd    string
	TestCmd     string
	StartCmd    string
	Lockfile    string
	Ambiguities []string
}

// Detector is the interface for stack detection implementations.
type Detector interface {
	Detect(root string) (DetectionResult, error)
}