// Package cli exposes the version command.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"devopsgen.dev/devopsgen/internal/templates"
)

// versionVars holds values injected at build time via ldflags.
// Defaults keep dev builds usable without a release pipeline.
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// NewVersionCommand returns the version command.
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long: `Print the application version, commit, build date, and bundled
template version.

On the MVP the template version tracks the binary version. Development
builds may report "dev" when a release version is unavailable.`,
		Args: classifyArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "devopsgen %s (commit %s, built %s)\n", version, commit, date)
			fmt.Fprintf(cmd.OutOrStdout(), "template version: %s\n", templates.Version)
			return nil
		},
	}
}

// ExitVersion is a test helper that prints version info to stdout.
func ExitVersion() {
	fmt.Fprintf(os.Stdout, "devopsgen %s (commit %s, built %s)\n", version, commit, date)
	fmt.Fprintf(os.Stdout, "template version: %s\n", templates.Version)
}
