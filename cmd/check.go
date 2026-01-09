package cmd

import (
	"github.com/serenitysz/serenity/internal/cmds/check"
	"github.com/spf13/cobra"
)

func NewCheckCmd() *cobra.Command {
	opts := &check.CheckOptions{}

	cmd := &cobra.Command{
		Use:   "check [path...]",
		Short: "Check code for issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			return check.Run(cmd, args, opts)
		},
	}

	cmd.Flags().Int("max-issues", 0, "Maximum number of issues")
	cmd.Flags().BoolVarP(&opts.Unsafe, "unsafe", "u", false, "Apply unsafe fixes")
	cmd.Flags().BoolVarP(&opts.Write, "write", "w", false, "Write changes to files")
	cmd.Flags().StringVarP(&opts.ConfigPath, "config", "c", "", "Use a custom config")
	cmd.Flags().Int64VarP(&opts.MaxFileSize, "max-file-size", "m", 0, "Maximum file size")

	return cmd
}

func init() {
	rootCmd.AddCommand(NewCheckCmd())
}
