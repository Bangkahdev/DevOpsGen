// Package config provides validation helpers for Config fields.
//
// Validation is pure and side-effect free: it never touches the
// filesystem, network, or credentials.
package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	errEmptyStack        = errors.New("stack is required")
	errInvalidStack      = errors.New("stack must be 'go' or 'nodejs'")
	errInvalidCI         = errors.New("ci must be 'github-actions'")
	errInvalidRegistry   = errors.New("registry must be 'dockerhub' or 'ghcr'")
	errInvalidImage      = errors.New("image must be 'namespace/repository' without host or tag")
	errInvalidPort       = errors.New("port must be in range 1-65535")
	errInvalidAppType    = errors.New("app-type is not supported for the selected stack")
	errEmptyImage        = errors.New("image is required")
	errInvalidHostname    = errors.New("ingress-host must be a valid RFC 1123 hostname")
	errInvalidHealthPath = errors.New("health-path must be a valid HTTP path")
	errInvalidCommand    = errors.New("command must be a single line without control characters")
)

// RFC 1123 hostname: lowercase labels of 1-63 chars, alnum and '-'.
// Length is enforced separately in validateHostname.
var hostnameRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)*$`)

// imageRe enforces "namespace/repository" with no host and no tag.
var imageRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9._-]{0,61}[a-z0-9])?/[a-z0-9]([a-z0-9._-]{0,61}[a-z0-9])?$`)

// healthPathRe enforces a rooted HTTP path.
var healthPathRe = regexp.MustCompile(`^/([a-zA-Z0-9._~/-]+)?$`)

// validAppTypes maps each stack to its allowed app types.
var validAppTypes = map[StackType][]AppType{
	StackGo:   {AppTypeBinary},
	StackNode: {AppTypeExpress, AppTypeNextJS, AppTypeVite, AppTypeNestJS, AppTypeGeneric},
}

// ValidateConfig checks the full configuration contract.
func ValidateConfig(cfg Config) error {
	if cfg.Stack == "" {
		return errEmptyStack
	}
	if cfg.Stack != StackGo && cfg.Stack != StackNode {
		return errInvalidStack
	}
	if cfg.CI != CIGitHubActions {
		return errInvalidCI
	}
	if cfg.Registry != RegistryDockerHub && cfg.Registry != RegistryGHCR {
		return errInvalidRegistry
	}
	if strings.TrimSpace(cfg.Image) == "" {
		return errEmptyImage
	}
	if !imageRe.MatchString(cfg.Image) {
		return errInvalidImage
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return errInvalidPort
	}
	if !validAppTypeForStack(cfg.Stack, cfg.AppType) {
		return errInvalidAppType
	}
	if err := validateCommand(cfg.BuildCmd); err != nil {
		return fmt.Errorf("build-cmd: %w", err)
	}
	if err := validateCommand(cfg.TestCmd); err != nil {
		return fmt.Errorf("test-cmd: %w", err)
	}
	if err := validateCommand(cfg.StartCmd); err != nil {
		return fmt.Errorf("start-cmd: %w", err)
	}
	if cfg.K8s.Ingress && cfg.K8s.IngressHost == "" {
		return errInvalidHostname
	}
	if cfg.K8s.IngressHost != "" {
		if err := validateHostname(cfg.K8s.IngressHost); err != nil {
			return err
		}
	}
	if cfg.K8s.HealthPath != "" && !healthPathRe.MatchString(cfg.K8s.HealthPath) {
		return errInvalidHealthPath
	}
	return nil
}

// validateHostname enforces RFC 1123: 1-253 chars, lowercase labels of
// 1-63 alnum/hyphen chars separated by dots.
func validateHostname(host string) error {
	if len(host) > 253 {
		return errInvalidHostname
	}
	if !hostnameRe.MatchString(host) {
		return errInvalidHostname
	}
	for _, label := range strings.Split(host, ".") {
		if len(label) == 0 || len(label) > 63 {
			return errInvalidHostname
		}
	}
	return nil
}

// ValidateOutputDir checks that the output directory is a safe path.
// Absolute paths are accepted (the target root may legitimately be
// absolute), but traversal outside the target is rejected.
func ValidateOutputDir(dir string) error {
	if strings.TrimSpace(dir) == "" {
		return errors.New("output-dir is required")
	}
	clean := filepath.Clean(dir)
	// Reject traversal that escapes the target root.
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || strings.HasPrefix(clean, "../") {
		return errors.New("output-dir must not escape the target directory")
	}
	// Reject embedded traversal segments.
	if strings.Contains(clean, string(filepath.Separator)+".." + string(filepath.Separator)) {
		return errors.New("output-dir must not escape the target directory")
	}
	return nil
}

// validAppTypeForStack reports whether appType is allowed for stack.
func validAppTypeForStack(stack StackType, appType AppType) bool {
	if appType == "" {
		return true // empty means default; resolved later by detector/template
	}
	for _, allowed := range validAppTypes[stack] {
		if allowed == appType {
			return true
		}
	}
	return false
}

// validateCommand enforces single-line, control-character-free commands.
func validateCommand(cmd string) error {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return nil // optional
	}
	if strings.ContainsAny(cmd, "\x00\r\n") {
		return errInvalidCommand
	}
	// Reject control characters other than tab.
	for _, r := range cmd {
		if r < 0x20 && r != '\t' {
			return errInvalidCommand
		}
	}
	return nil
}

// ImageRef returns the fully-qualified image reference for the registry.
func (cfg Config) ImageRef() string {
	switch cfg.Registry {
	case RegistryGHCR:
		return "ghcr.io/" + cfg.Image
	default:
		return "docker.io/" + cfg.Image
	}
}

// TagList returns the tags to publish, always including the commit SHA alias.
func (cfg Config) TagList() []string {
	alias := cfg.Tag
	if alias == "" {
		alias = DefaultTag
	}
	return []string{alias, "${{ github.sha }}"}
}