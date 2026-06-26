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

var runCmd = &cobra.Command{
	Use:   "run <config.yaml>",
	Short: "Start a walkthrough from a YAML file",
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
	attachProgressCLI(eng, wt)

	set := openSettings()
	abs := rememberLastConfig(set, args[0])
	return overlay.New(eng, set, abs).Run()
}

// rememberLastConfig records a CLI-loaded walkthrough so a later double-click
// (picker entry) reopens it, and returns the absolute path it stored. The path
// is absolute because the next launch may have a different working directory,
// and so a `next:` hand-off resolves correctly regardless of CWD. Best-effort:
// a failure to persist must not stop the walkthrough from running.
func rememberLastConfig(set *settings.Store, path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	ns := set.Get()
	ns.LastConfig = settings.LastConfig{Path: abs, Embedded: false}
	_ = set.Save(ns)
	return abs
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

// attachProgressCLI wires the engine to the on-disk progress store for the CLI
// run command, restoring any saved position. Use `gothrough clear progress` to
// reset.
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
	if index, stepID, choices, ok := h.Load(); ok {
		eng.Restore(index, stepID, choices)
	}
	eng.UsePersister(h)
}
