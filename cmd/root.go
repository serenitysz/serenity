package cmd

import (
	"os"

	"github.com/serenitysz/serenity/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: false,
	Version:       version.Version,
	Use:           "serenity <command> [flags]",
	Short:         "Serenity is an aggressive, no-noise and ultra fast Go linter",
}

func Exec() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("color", "auto", "Color output: auto, off, force")

	rootCmd.PersistentFlags().Bool("verbose", false, "Print additional diagnostics and processed files")

	rootCmd.PersistentFlags().String("config", "", "Path to configuration file (Auto-discovered if omitted)")

	rootCmd.PersistentFlags().Int("max-issues", 20, "Limit the maximum number of reported issues (0 = unlimited)")

	rootCmd.PersistentFlags().Bool("skip-parse-errors", false, "Skip files with syntax errors instead of failing")
}
