package check

import (
	"github.com/serenitysz/serenity/internal/exception"
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
		return exception.InternalError("failed due to error")
	}

	return exception.InternalError("issues found")
}
