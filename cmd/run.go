package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/XotoX1337/GoThrough/config"
	"github.com/XotoX1337/GoThrough/engine"
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
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("\n=== %s ===\n\n", wt.Title)
	printStep(eng)

	for {
		fmt.Print("\n[n]ext  [p]rev  [q]uit > ")
		if !scanner.Scan() {
			break
		}

		switch strings.TrimSpace(strings.ToLower(scanner.Text())) {
		case "n", "next", "":
			if err := eng.Next(); err != nil {
				if errors.Is(err, engine.ErrAlreadyLast) {
					fmt.Println("Walkthrough complete!")
					return nil
				}
				return err
			}
			printStep(eng)
		case "p", "prev":
			if err := eng.Prev(); err != nil {
				if errors.Is(err, engine.ErrAlreadyFirst) {
					fmt.Println("Already at the first step.")
					continue
				}
				return err
			}
			printStep(eng)
		case "q", "quit":
			fmt.Println("Bye.")
			return nil
		}
	}

	return scanner.Err()
}

func printStep(eng *engine.Engine) {
	current, total := eng.Progress()
	step := eng.Current()
	fmt.Printf("Step %d/%d — %s\n", current, total, step.Title)
	if step.Description != "" {
		fmt.Printf("  %s\n", step.Description)
	}
}
