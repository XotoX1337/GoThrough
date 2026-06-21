package cmd

import (
	"github.com/XotoX1337/GoThrough/config"
	"github.com/XotoX1337/GoThrough/engine"
	"github.com/XotoX1337/GoThrough/overlay"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <config.yaml>",
	Short: "Start a walkthrough",
	Args:  cobra.ExactArgs(1),
	RunE:  runWalkthrough,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runWalkthrough(_ *cobra.Command, args []string) error {
	wt, err := config.Load(args[0])
	if err != nil {
		return err
	}
	eng := engine.New(wt)
	return overlay.New(eng).Run()
}
