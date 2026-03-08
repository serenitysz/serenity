package check

import (
	"fmt"
	"io"
	"os"

	"github.com/serenitysz/serenity/internal/exception"
	"github.com/serenitysz/serenity/internal/render"
	"github.com/serenitysz/serenity/internal/rules"
)

type issueSummary struct {
	hasIssues bool
	errors    int
	warnings  int
	infos     int
	fixables  int
	writeMode bool
	renderer  issueRenderer
}

func newIssueSummary(writeMode bool) issueSummary {
	return issueSummary{
		writeMode: writeMode,
		renderer:  newIssueRenderer(os.Stderr),
	}
}

func (s *issueSummary) add(issues []rules.Issue) {
	if len(issues) == 0 {
		return
	}

	s.hasIssues = true

	for _, issue := range issues {
		msg := rules.FormatMessage(issue)
		s.addSeverity(issue.Severity)
		if rules.IsFixable(issue.ID) {
			s.fixables++
		}

		s.renderer.write(issue, msg)
	}
}

func (s issueSummary) err() error {
	if !s.hasIssues {
		return nil
	}

	return summaryError{message: s.footer()}
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

func (s issueSummary) footer() string {
	total := s.errors + s.warnings + s.infos
	base := fmt.Sprintf("%s found (%s)", pluralize(total, "issue"), s.describe())

	if s.fixables == 0 {
		return base
	}

	if s.writeMode {
		return fmt.Sprintf("%s, %s automatically fixed", base, pluralize(s.fixables, "issue"))
	}

	return fmt.Sprintf("%s, %s. Use --write to apply automatic fixes.", base, fixableText(s.fixables))
}

func fixableText(count int) string {
	if count == 1 {
		return "1 issue is fixable"
	}

	return fmt.Sprintf("%d issues are fixable", count)
}

type summaryError struct {
	message string
}

func (e summaryError) Error() string {
	return e.message
}

func (e summaryError) Unwrap() error {
	return exception.ErrCommand
}

func (e summaryError) WriteCLI(w io.Writer, noColor bool) {
	_, _ = fmt.Fprintf(w, "%s\n", render.Paint(e.message, render.Yellow, noColor))
}

func pluralize(count int, singular string) string {
	if count == 1 {
		return "1 " + singular
	}

	return fmt.Sprintf("%d %ss", count, singular)
}
