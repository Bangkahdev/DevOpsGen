// Package validator performs internal artifact validation.
//
// Validation is always internal and runs even when external validators
// (e.g. kubeconform) are unavailable. External validation is additive and
// is reported separately from internal results.
package validator

import (
	"devopsgen.dev/devopsgen/internal/output"
)

// Result categorizes a validation outcome.
type Result int

const (
	ResultValid Result = iota
	ResultInvalid
	ResultExternalSkipped
)

// Validator is the artifact validation contract.
type Validator interface {
	Validate(artifacts []output.Artifact) (Result, error)
}