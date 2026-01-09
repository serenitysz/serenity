package utils

import (
	"fmt"

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
		label = "ERROR"
		color = render.Red
	case rules.SeverityWarn:
		label = "WARN"
		color = render.Yellow
	case rules.SeverityInfo:
		label = "INFO"
		color = render.Blue
	default:
		label = "ISSUE"
		color = render.Reset
	}

	fmt.Printf("%s%s:%d:%d: [%s] %s%s\n",
		color,
		issue.Pos.Filename,
		issue.Pos.Line,
		issue.Pos.Column,
		msg,
		label,
		render.Reset,
	)
}
