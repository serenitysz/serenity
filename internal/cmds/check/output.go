package check

import (
	"fmt"

	"github.com/serenitysz/serenity/internal/exception"
	"github.com/serenitysz/serenity/internal/rules"
	"github.com/serenitysz/serenity/internal/utils"
)

type issueSummary struct {
	hasIssues bool
	errors    int
	warnings  int
	infos     int
}

func (s *issueSummary) add(issues []rules.Issue) {
	if len(issues) == 0 {
		return
	}

	s.hasIssues = true

	for _, issue := range issues {
		msg := rules.FormatMessage(issue)
		s.addSeverity(issue.Severity)

		utils.FormatLog(issue, msg)
	}
}

func (s issueSummary) err() error {
	if !s.hasIssues {
		return nil
	}

	return exception.CommandError("found %s", s.describe())
}

func (s *issueSummary) addSeverity(severity rules.Severity) {
	switch severity {
	case rules.SeverityError:
		s.errors++
	case rules.SeverityInfo:
		s.infos++
	default:
		s.warnings++
	}
}

func (s issueSummary) describe() string {
	parts := make([]string, 0, 3)

	if s.errors > 0 {
		parts = append(parts, pluralize(s.errors, "error"))
	}

	if s.warnings > 0 {
		parts = append(parts, pluralize(s.warnings, "warning"))
	}

	if s.infos > 0 {
		parts = append(parts, pluralize(s.infos, "info diagnostic"))
	}

	switch len(parts) {
	case 0:
		return "no issues"
	case 1:
		return parts[0]
	case 2:
		return parts[0] + " and " + parts[1]
	default:
		return parts[0] + ", " + parts[1] + ", and " + parts[2]
	}
}

func pluralize(count int, singular string) string {
	if count == 1 {
		return "1 " + singular
	}

	return fmt.Sprintf("%d %ss", count, singular)
}
