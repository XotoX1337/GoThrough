package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/XotoX1337/GoThrough/configstore"
	"github.com/XotoX1337/GoThrough/progress"
	"github.com/XotoX1337/GoThrough/settings"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Reset saved progress, settings, or the downloaded config cache",
	Long: `Reset persisted GoThrough state.

These subcommands run headless (no overlay) and exit:

  gothrough clear progress <game> [chapter]   reset a whole game, or one chapter
  gothrough clear settings                     restore default settings
  gothrough clear cache                        delete the downloaded config cache
  gothrough clear all                          progress + settings + cache`,
}

var clearProgressCmd = &cobra.Command{
	Use:   "progress <game> [chapter]",
	Short: "Delete saved progress for a game (or a single chapter)",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runClearProgress,
}

var clearSettingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Reset all settings to their defaults",
	Args:  cobra.NoArgs,
	RunE:  runClearSettings,
}

var clearCacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Delete every downloaded config from the cache",
	Args:  cobra.NoArgs,
	RunE:  runClearCache,
}

var clearAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Reset progress, settings and the config cache",
	Args:  cobra.NoArgs,
	RunE:  runClearAll,
}

func init() {
	clearCmd.AddCommand(clearProgressCmd, clearSettingsCmd, clearCacheCmd, clearAllCmd)
	rootCmd.AddCommand(clearCmd)
}

// openProgressStore loads the on-disk progress store at its default path.
func openProgressStore() (*progress.Store, error) {
	path, err := progress.DefaultPath()
	if err != nil {
		return nil, err
	}
	return progress.Open(path)
}

func runClearProgress(_ *cobra.Command, args []string) error {
	store, err := openProgressStore()
	if err != nil {
		return err
	}
	game := args[0]

	if len(args) == 2 {
		chapter, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("chapter must be a number: %q", args[1])
		}
		if err := store.DeleteChapter(game, chapter); err != nil {
			return err
		}
		fmt.Printf("Cleared progress for %s chapter %d.\n", game, chapter)
		return nil
	}

	if err := store.DeleteGame(game); err != nil {
		return err
	}
	fmt.Printf("Cleared all progress for %s.\n", game)
	return nil
}

func runClearSettings(_ *cobra.Command, _ []string) error {
	if err := resetSettings(); err != nil {
		return err
	}
	fmt.Println("Settings reset to defaults.")
	return nil
}

func runClearCache(_ *cobra.Command, _ []string) error {
	if err := configstore.ClearCache(); err != nil {
		return err
	}
	fmt.Println("Config cache cleared.")
	return nil
}

func runClearAll(_ *cobra.Command, _ []string) error {
	store, err := openProgressStore()
	if err != nil {
		return err
	}
	if err := store.Clear(); err != nil {
		return err
	}
	if err := resetSettings(); err != nil {
		return err
	}
	if err := configstore.ClearCache(); err != nil {
		return err
	}
	fmt.Println("Cleared all progress, reset settings, and emptied the config cache.")
	return nil
}

// resetSettings persists the default settings, replacing whatever was stored.
func resetSettings() error {
	path, err := settings.DefaultPath()
	if err != nil {
		return err
	}
	store, err := settings.Open(path)
	if err != nil {
		return err
	}
	return store.Save(settings.Defaults())
}
