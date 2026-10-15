// Package generator renders templates into in-memory artifacts.
//
// The generator is pure: it performs no filesystem I/O, no network access,
// and no credential handling.
package generator

import (
	"devopsgen.dev/devopsgen/internal/config"
	"devopsgen.dev/devopsgen/internal/output"
)

// Generator is the core rendering contract.
type Generator interface {
	Generate(cfg config.Config) ([]output.Artifact, error)
}