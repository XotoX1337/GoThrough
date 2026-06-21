package cmd

import (
	"fmt"
	"os"

	"github.com/XotoX1337/GoThrough/config"
	"github.com/XotoX1337/GoThrough/engine"
	"github.com/XotoX1337/GoThrough/overlay"
	"github.com/XotoX1337/GoThrough/progress"
	"github.com/spf13/cobra"
)

var freshStart bool

var runCmd = &cobra.Command{
	Use:   "run <config.yaml>",
	Short: "Start a walkthrough",
	Args:  cobra.ExactArgs(1),
	RunE:  runWalkthrough,
}

func init() {
	runCmd.Flags().BoolVar(&freshStart, "fresh", false, "ignore saved progress and start at step 1")
	rootCmd.AddCommand(runCmd)
}

func runWalkthrough(_ *cobra.Command, args []string) error {
	wt, err := config.Load(args[0])
	if err != nil {
		return err
	}

	eng := engine.New(wt)
	attachProgress(eng, wt)

	return overlay.New(eng).Run()
}

// attachProgress wires the engine to the on-disk progress store: it restores
// the last saved step (unless --fresh) and enables autosave on future changes.
// Persistence is best-effort — if the store can't be opened we warn and run
// without it rather than refusing to start the walkthrough.
func attachProgress(eng *engine.Engine, wt *config.Walkthrough) {
	path, err := progress.DefaultPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: progress disabled: %v\n", err)
		return
	}
	store, err := progress.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: progress disabled: %v\n", err)
		return
	}

	h := store.For(wt)
	if !freshStart {
		if index, stepID, ok := h.Load(); ok {
			eng.Restore(index, stepID)
		}
	}
	eng.UsePersister(h)
}
