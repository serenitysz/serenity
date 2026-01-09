package check

import (
	"fmt"
	"os"

	"github.com/serenitysz/serenity/internal/linter"
	"github.com/serenitysz/serenity/internal/rules"
	"github.com/spf13/cobra"
)

func Run(cmd *cobra.Command, args []string, opts *CheckOptions) error {
	cfg, err := loadConfig(opts.ConfigPath)

	if err != nil {
		return err
	}

	maxIssues, err := resolveMaxIssues(cmd, cfg)

	if err != nil {
		return err
	}

	l := linter.New(
		opts.Write,
		opts.Unsafe,
		cfg,
		maxIssues,
		opts.MaxFileSize,
	)

	issues, err := runOnPaths(l, args)

	if err != nil {
		return err
	}

	return handleIssues(issues)
}

func resolveMaxIssues(cmd *cobra.Command, cfg *rules.LinterOptions) (int, error) {
	maxIssues, err := cmd.Flags().GetInt("max-issues")
	if err != nil {
		return 0, err
	}

	if !cmd.Flags().Changed("max-issues") {
		maxIssues = int(cfg.GetMaxIssues())
	}

	return maxIssues, nil
}

func runOnPaths(l *linter.Linter, args []string) ([]rules.Issue, error) {
	var all []rules.Issue

	for _, p := range args {
		if p == "" || p == "." {
			wd, err := os.Getwd()

			if err != nil {
				return nil, fmt.Errorf("get wd: %w", err)
			}

			p = wd
		}

		issues, err := l.ProcessPath(p)

		if err != nil {
			return nil, err
		}

		all = append(all, issues...)
	}

	return all, nil
}
