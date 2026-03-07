package cmd

import (
	"os"

	"github.com/serenitysz/serenity/internal/exception"
	"github.com/serenitysz/serenity/internal/render"
	"github.com/serenitysz/serenity/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       version.Version,
	Use:           "serenity <command> [flags]",
	Short:         "Serenity is an aggressive, no-noise and ultra fast Go linter",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		noColor, err := cmd.Flags().GetBool("no-color")
		if err != nil {
			return exception.InternalError("could not read --no-color: %w", err)
		}

		render.SetNoColor(noColor)
		return nil
	},
}

func Exec() {
	err := rootCmd.Execute()
	noColor, _ := rootCmd.PersistentFlags().GetBool("no-color")

	exception.Write(os.Stderr, err, noColor)

	os.Exit(exception.ExitCode(err))
}

func init() {
	rootCmd.PersistentFlags().Bool("no-color", false, "Remove color from the output")
	rootCmd.PersistentFlags().Bool("verbose", false, "Print additional diagnostics and processed files")
	rootCmd.PersistentFlags().String("config", "", "Path to configuration file (Auto-discovered if omitted)")
}
