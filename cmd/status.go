package cmd

import (
	"github.com/serenitysz/serenity/internal/cmds/status"
	"github.com/serenitysz/serenity/internal/exception"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display the current status of Serenity",
	RunE: func(cmd *cobra.Command, args []string) error {
		noColor, err := cmd.Flags().GetBool("no-color")

		if err != nil {
			return exception.InternalError("%v", err)
		}

		return status.Get(noColor)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
