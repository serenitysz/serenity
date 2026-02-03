package cmd

import (
	"github.com/serenitysz/serenity/internal/exception"
	"github.com/serenitysz/serenity/internal/version"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Serenity to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		noColor, err := cmd.Flags().GetBool("no-color")

		if err != nil {
			return exception.InternalError("%v", err)
		}

		return version.Update(noColor)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
