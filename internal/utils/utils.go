package utils

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/serenitysz/serenity/internal/rules"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
)

func FormatLog(issue rules.Issue, msg string) {
	var label, color string

	switch issue.Severity {
	case rules.SeverityError:
		label = "ERROR"
		color = colorRed
	case rules.SeverityWarn:
		label = "WARN"
		color = colorYellow
	case rules.SeverityInfo:
		label = "INFO"
		color = colorBlue
	default:
		label = "ISSUE"
		color = colorReset
	}

	fmt.Printf("%s%s:%d:%d: [%s] %s%s\n",
		color,
		issue.Pos.Filename,
		issue.Pos.Line,
		issue.Pos.Column,
		msg,
		label,
		colorReset,
	)
}

func GetActualCommit() (string, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty-format:%H")

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error to get hash from last commit: %w", err)
	}

	hash := strings.TrimSpace(string(out))
	if len(hash) < 8 {
		return "", fmt.Errorf("invalid commit hash")
	}

	log := hash[:8]

	return log, nil
}
