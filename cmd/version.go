package cmd

import (
	"fmt"

	"github.com/serenitysz/serenity/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the current version of Serenity",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(version.Version)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
