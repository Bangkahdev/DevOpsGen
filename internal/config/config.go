// Package config defines the configuration model, enums, value resolution
// priority, and input validation for the core engine.
//
// This package must not import CLI, TUI, or filesystem implementations.
package config

// StackType identifies a supported programming-language stack.
type StackType string

const (
	StackGo   StackType = "go"
	StackNode StackType = "nodejs"
)

// AppType identifies the application shape within a stack.
type AppType string

const (
	AppTypeBinary  AppType = "binary"  // Go
	AppTypeExpress AppType = "express" // Node.js
	AppTypeNextJS  AppType = "nextjs"  // Node.js
	AppTypeVite    AppType = "vite"    // Node.js
	AppTypeNestJS  AppType = "nestjs"  // Node.js
	AppTypeGeneric AppType = "generic" // Node.js
)

// CIProvider identifies a supported CI/CD provider.
type CIProvider string

const (
	CIGitHubActions CIProvider = "github-actions"
)

// Registry identifies a supported container image registry.
type Registry string

const (
	RegistryDockerHub Registry = "dockerhub"
	RegistryGHCR      Registry = "ghcr"
)

// K8sOptions holds Kubernetes-related configuration.
type K8sOptions struct {
	Enabled     bool
	Ingress     bool
	IngressHost string
	HealthPath  string
}

// GenerationOptions holds filesystem-operation policies.
type GenerationOptions struct {
	Force  bool
	DryRun bool
}

// Config is the complete configuration contract consumed by the generator.
//
// Field validation happens before rendering; required fields are enforced
// per the SRS 4.1 parameter matrix.
type Config struct {
	Stack      StackType
	AppType    AppType
	CI         CIProvider
	Registry   Registry
	Image      string // namespace/repository, no registry host, no tag
	Port       int
	BuildCmd   string
	TestCmd    string
	StartCmd   string
	Tag        string // additional alias; default "latest"
	K8s        K8sOptions
	OutputDir  string
	Generation GenerationOptions
}

// DefaultTag is the default mutable alias published alongside the commit SHA.
const DefaultTag = "latest"

// MaxTagLen is the maximum allowed length for a tag value.
const MaxTagLen = 128
