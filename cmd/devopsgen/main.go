// Command-line entry point for DevOpsGen.
package main

import (
	"os"

	"devopsgen.dev/devopsgen/internal/cli"
)

func main() {
	root := cli.NewRootCommand()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}