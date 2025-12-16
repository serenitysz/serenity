package cmd

import (
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

var checkWrite, checkUnsafe bool

func init() {
	checkCmd.Flags().BoolVarP(&checkUnsafe, "unsafe", "u", false, "Apply unsafe fixes")
	checkCmd.Flags().BoolVarP(&checkWrite, "write", "w", false, "Write changes to files")

	rootCmd.AddCommand(checkCmd)
}

func Check(cmd *cobra.Command, args []string) error {
	var linterCfg *rules.Config

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
		cfg, err := config.ReadConfigFile(path)
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

	l := linter.New(checkWrite, checkUnsafe, linterCfg, maxIssues, linterCfg.MaxFileSize)

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
