// Package cmd defines the CLI surface via Cobra.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/XotoX1337/GoThrough/overlay"
)

func init() {
	// GoThrough is a GUI app — disable Cobra's Windows mousetrap which would
	// otherwise show "This is a command line tool" and exit when the binary is
	// launched by double-clicking in Windows Explorer.
	cobra.MousetrapHelpText = ""
}

var rootCmd = &cobra.Command{
	Use:   "gothrough",
	Short: "Game-agnostic walkthrough overlay",
	Long: `GoThrough — game-agnostic walkthrough overlay.

Double-click the binary (or run with no arguments) to open the config picker.
Use 'gothrough run <config.yaml>' to start a specific walkthrough directly.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return overlay.New(nil, openSettings(), "").Run()
	},
}

// Execute runs the root command and returns any error.
func Execute() error {
	return rootCmd.Execute()
}

// SetVersion injects the build version into the root command.
func SetVersion(v string) {
	rootCmd.Version = v
}
