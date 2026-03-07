package utils

import (
	"fmt"
	"os"

	"github.com/serenitysz/serenity/internal/render"
	"github.com/serenitysz/serenity/internal/rules"
)

func Ptr[T any](value T) *T {
	return &value
}

func FormatLog(issue rules.Issue, msg string) {
	var label, color string

	switch issue.Severity {
	case rules.SeverityError:
		label = "error"
		color = render.Red
	case rules.SeverityWarn:
		label = "warn"
		color = render.Yellow
	case rules.SeverityInfo:
		label = "info"
		color = render.Blue
	default:
		label = "issue"
		color = render.Reset
	}

	formattedLabel := render.Tag(label, color, false)
	if color == render.Reset {
		formattedLabel = fmt.Sprintf("%-5s", label)
	}

	fmt.Fprintf(os.Stderr, "%s %s:%d:%d  %s\n",
		formattedLabel,
		issue.Filename(),
		issue.LineNumber(),
		issue.ColumnNumber(),
		msg,
	)
}
