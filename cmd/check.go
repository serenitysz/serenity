package cmd

import (
	// "cmp"
	"errors"
	"fmt"
	"os"

	"github.com/serenitysz/serenity/internal/config"
	"github.com/serenitysz/serenity/internal/linter"
	"github.com/serenitysz/serenity/internal/rules"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	RunE:  Check,
	Use:   "check [path...]",
	Short: "Check code for issues",
}

var maxFileSize int64
var write, unsafe bool

func init() {
	checkCmd.Flags().BoolVarP(&unsafe, "unsafe", "u", false, "Apply unsafe fixes")
	checkCmd.Flags().BoolVarP(&write, "write", "w", false, "Write changes to files")
	checkCmd.Flags().Int64VarP(&maxFileSize, "max-file-size", "m", int64(0), "Use a custom maximum file size in the check")

	rootCmd.AddCommand(checkCmd)
}

func Check(cmd *cobra.Command, args []string) error {
	var linterCfg *rules.LinterOptions

	path, err := config.GetConfigFilePath()

	if err != nil {
		return err
	}

	exists, err := config.CheckHasConfigFile(path)

	if err != nil {
		return err
	}

	linterCfg = config.GenDefaultConfig()

	if exists {
		cfg, err := config.ReadConfig(path)
		if err != nil {
			return err
		}
		linterCfg = cfg
	}

	var issues []rules.Issue

	maxIssues, err := cmd.Flags().GetInt("max-issues")

	if err != nil {
		return err
	}

	l := linter.New(write, unsafe, linterCfg, maxIssues, maxFileSize)

	for _, v := range args {
		if v == "" || v == "." {
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("error to get working directory: %w", err)
			}

			v = wd
		}

		i, err := l.ProcessPath(v)

		if err != nil {
			return err
		}

		issues = append(issues, i...)
	}

	if len(issues) > 0 {
		for _, v := range issues {
			fmt.Printf("%s:%d:%d: %s\n", v.Pos.Filename, v.Pos.Line, v.Pos.Column, v.Message)
			// TODO: melhorar esse log dps
		}

		return errors.New("issues found")
	}

	return nil
}
