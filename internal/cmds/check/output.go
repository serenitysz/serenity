package check

import (
	"errors"

	"github.com/serenitysz/serenity/internal/rules"
	"github.com/serenitysz/serenity/internal/utils"
)

func handleIssues(issues []rules.Issue) error {
	if len(issues) == 0 {
		return nil
	}

	hasError := false

	for _, issue := range issues {
		msg := rules.FormatMessage(issue)

		if issue.Severity == rules.SeverityError {
			hasError = true
		}

		utils.FormatLog(issue, msg)
	}

	if hasError {
		return errors.New("failed due to error")
	}

	return errors.New("issues found")
}
