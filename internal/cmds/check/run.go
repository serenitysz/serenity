package check

import (
	"os"

	"github.com/serenitysz/serenity/internal/exception"
	"github.com/serenitysz/serenity/internal/linter"
	"github.com/serenitysz/serenity/internal/rules"
	"github.com/spf13/cobra"
)

func Run(cmd *cobra.Command, args []string, opts *CheckOptions) error {
	if err := validateOptions(opts); err != nil {
		return err
	}

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

	return runOnPaths(l, args)
}

func validateOptions(opts *CheckOptions) error {
	if opts != nil && opts.Unsafe && !opts.Write {
		return exception.CommandError("--unsafe requires --write")
	}

	return nil
}

func resolveMaxIssues(cmd *cobra.Command, cfg *rules.LinterOptions) (int, error) {
	maxIssues, err := cmd.Flags().GetInt("max-issues")

	if err != nil {
		return 0, exception.InternalError("could not read --max-issues: %w", err)
	}

	if !cmd.Flags().Changed("max-issues") {
		maxIssues = int(cfg.GetMaxIssues())
	}

	return maxIssues, nil
}

func runOnPaths(l *linter.Linter, args []string) error {
	if len(args) == 0 {
		args = []string{"."}
	}

	summary := newIssueSummary(l.Write)
	remaining := l.MaxIssues

	for _, p := range args {
		if remaining == 0 && l.MaxIssues > 0 {
			break
		}

		if p == "" || p == "." {
			wd, err := os.Getwd()

			if err != nil {
				return exception.InternalError("could not determine the current working directory: %w", err)
			}

			p = wd
		}

		if l.MaxIssues > 0 {
			l.MaxIssues = remaining
		}

		issues, err := l.ProcessPath(p)

		if err != nil {
			return exception.InternalError("could not lint %q: %w", p, err)
		}

		summary.add(issues)

		if remaining > 0 {
			remaining -= len(issues)
			if remaining < 0 {
				remaining = 0
			}
		}
	}

	return summary.err()
}
