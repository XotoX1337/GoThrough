// Package cmd defines the CLI surface via Cobra.
package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gothrough",
	Short: "Game-agnostic walkthrough overlay",
}

// Execute runs the root command and returns any error.
func Execute() error {
	return rootCmd.Execute()
}

// SetVersion injects the build version into the root command.
func SetVersion(v string) {
	rootCmd.Version = v
}
