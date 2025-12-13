package cmd

import "github.com/spf13/cobra"

var formatCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Format files",
	RunE:  format,
}

func init() {
	rootCmd.AddCommand(formatCmd)
}

func format(cmd *cobra.Command, args []string) error {
	return nil
}
