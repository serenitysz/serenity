package check

import (
	"github.com/serenitysz/serenity/internal/exception"
	"github.com/serenitysz/serenity/internal/rules"
	"github.com/serenitysz/serenity/internal/utils"
)

type issueSummary struct {
	hasIssues bool
	hasError  bool
}

func (s *issueSummary) add(issues []rules.Issue) {
	if len(issues) == 0 {
		return
	}

	s.hasIssues = true

	for _, issue := range issues {
		msg := rules.FormatMessage(issue)

		if issue.Severity == rules.SeverityError {
			s.hasError = true
		}

		utils.FormatLog(issue, msg)
	}
}

func (s issueSummary) err() error {
	if !s.hasIssues {
		return nil
	}

	if s.hasError {
		return exception.CommandError("failed due to error")
	}

	return exception.CommandError("issues found")
}
