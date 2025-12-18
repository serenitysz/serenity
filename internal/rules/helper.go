package rules

import (
	"errors"
)

func VerifyIssues(cfg *LinterOptions, issues []Issue) error {
	var maxIssues int8

	if cfg.Linter.Issues != nil && cfg.Linter.Issues.Max != nil {
		maxIssues = *cfg.Linter.Issues.Max
	}

	if int8(len(issues)) >= maxIssues {
		return errors.New("issues limit reached")
	}

	return nil
}
