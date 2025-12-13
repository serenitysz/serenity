package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "golint",
	Short: "A modern linter for Go, inspired by Biome",
	Long:  "Fast, opinionated linter and formatter for Go with auto-fix capabilities",
}
