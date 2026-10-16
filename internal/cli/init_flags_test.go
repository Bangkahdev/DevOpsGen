package cli

import (
	"testing"

	"github.com/spf13/cobra"

	"devopsgen.dev/devopsgen/internal/config"
)

func TestAddInitFlags_ToConfig(t *testing.T) {
	cmd := &cobra.Command{Use: "init"}
	flags := AddInitFlags(cmd)
	flags.Stack = "go"
	flags.CI = "github-actions"
	flags.Registry = "ghcr"
	flags.Image = "bangkah/my-app"
	flags.Port = 8080

	cfg, err := flags.ToConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Registry != config.RegistryGHCR {
		t.Fatalf("registry: %s", cfg.Registry)
	}
	if cfg.ImageRef() != "ghcr.io/bangkah/my-app" {
		t.Fatalf("image ref: %s", cfg.ImageRef())
	}
}

func TestToConfig_EmptyStackAllowed(t *testing.T) {
	cmd := &cobra.Command{Use: "init"}
	flags := AddInitFlags(cmd)
	flags.Image = "a/b"
	flags.Port = 8080
	// Stack is allowed to be empty here; resolved later by detection.
	cfg, err := flags.ToConfig()
	if err != nil {
		t.Fatalf("unexpected error for empty stack: %v", err)
	}
	if cfg.Image != "a/b" {
		t.Fatalf("image: %s", cfg.Image)
	}
}

func TestToConfig_InvalidPort(t *testing.T) {
	cmd := &cobra.Command{Use: "init"}
	flags := AddInitFlags(cmd)
	flags.Stack = "go"
	flags.Image = "a/b"
	flags.Port = 99999
	if _, err := flags.ToConfig(); err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestMapCommandError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, 0},
		{"conflict", errConflict, 3},
		{"validation", errValidation, 4},
		{"ambiguous", errAmbiguousStack, 5},
		{"other", errValidation, 4},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := mapCommandError(c.err); got != c.want {
				t.Fatalf("mapCommandError(%v)=%d want %d", c.err, got, c.want)
			}
		})
	}
}

func TestParsePort(t *testing.T) {
	if p, err := ParsePort("8080"); err != nil || p != 8080 {
		t.Fatalf("ParsePort 8080: %d %v", p, err)
	}
	if p, err := ParsePort(""); err != nil || p != 0 {
		t.Fatalf("ParsePort empty: %d %v", p, err)
	}
	if _, err := ParsePort("abc"); err == nil {
		t.Fatal("expected error for non-numeric port")
	}
}
