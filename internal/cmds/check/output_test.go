package check

import (
	"errors"
	"testing"

	"github.com/serenitysz/serenity/internal/exception"
)

func TestIssueSummaryErrIncludesAllSeverityCounts(t *testing.T) {
	t.Parallel()

	summary := issueSummary{
		hasIssues: true,
		errors:    2,
		warnings:  1,
		infos:     3,
	}

	err := summary.err()
	if !errors.Is(err, exception.ErrCommand) {
		t.Fatalf("expected command error, got %v", err)
	}

	if got := exception.Message(err); got != "found 2 errors, 1 warning, and 3 info diagnostics" {
		t.Fatalf("unexpected summary message: %q", got)
	}
}

func TestIssueSummaryErrHandlesWarningsOnly(t *testing.T) {
	t.Parallel()

	summary := issueSummary{
		hasIssues: true,
		warnings:  1,
	}

	err := summary.err()
	if !errors.Is(err, exception.ErrCommand) {
		t.Fatalf("expected command error, got %v", err)
	}

	if got := exception.Message(err); got != "found 1 warning" {
		t.Fatalf("unexpected warnings-only summary: %q", got)
	}
}
