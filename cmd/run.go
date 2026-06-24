package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/XotoX1337/GoThrough/config"
	"github.com/XotoX1337/GoThrough/engine"
	"github.com/XotoX1337/GoThrough/overlay"
	"github.com/XotoX1337/GoThrough/progress"
	"github.com/XotoX1337/GoThrough/settings"
)

var freshStart bool

var runCmd = &cobra.Command{
	Use:   "run <config.yaml>",
	Short: "Start a walkthrough from a YAML file",
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
	attachProgressCLI(eng, wt)

	return overlay.New(eng, openSettings()).Run()
}

// openSettings loads the user settings store, falling back to an in-memory
// defaults store on failure so the app always starts.
func openSettings() *settings.Store {
	path, err := settings.DefaultPath()
	if err == nil {
		if store, oerr := settings.Open(path); oerr == nil {
			return store
		} else {
			err = oerr
		}
	}
	fmt.Fprintf(os.Stderr, "warning: settings unavailable, using defaults: %v\n", err)
	store, _ := settings.Open(filepath.Join(os.TempDir(), "gothrough-settings.json"))
	return store
}

// attachProgressCLI wires the engine to the on-disk progress store for the
// CLI run command, honouring the --fresh flag.
func attachProgressCLI(eng *engine.Engine, wt *config.Walkthrough) {
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
