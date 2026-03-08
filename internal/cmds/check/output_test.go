package check

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/serenitysz/serenity/internal/exception"
	"github.com/serenitysz/serenity/internal/render"
	"github.com/serenitysz/serenity/internal/rules"
)

func TestIssueSummaryErrIncludesAllSeverityCounts(t *testing.T) {
	t.Parallel()

	summary := issueSummary{
		hasIssues: true,
		errors:    2,
		warnings:  1,
		infos:     3,
		fixables:  2,
	}

	err := summary.err()
	if !errors.Is(err, exception.ErrCommand) {
		t.Fatalf("expected command error, got %v", err)
	}

	if got := exception.Message(err); got != "6 issues found (2 errors, 1 warning, and 3 info diagnostics), 2 issues are fixable. Use --write to apply automatic fixes." {
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

	if got := exception.Message(err); got != "1 issue found (1 warning)" {
		t.Fatalf("unexpected warnings-only summary: %q", got)
	}
}

func TestIssueRendererWritesBiomeStyleBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")

	if err := os.WriteFile(path, []byte("package sample\nfunc handle(ctx int) {}\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var buf bytes.Buffer

	render.SetNoColor(true)
	defer render.SetNoColor(false)

	r := issueRenderer{
		cwd: dir,
		out: &buf,
	}

	r.write(rules.Issue{
		ID:       rules.UseContextInFirstParamID,
		Path:     path,
		Line:     2,
		Column:   13,
		Severity: rules.SeverityWarn,
	}, `parameter "ctx" in function "handle" has type context.Context and must be the first parameter`)

	want := "Warn • context-first-param (7)\n" +
		"sample.go:2:13\n\n" +
		"  1 │ package sample\n" +
		"> 2 │ func handle(ctx int) {}\n" +
		"    │             ^\n\n" +
		"Hint:\n" +
		"  parameter \"ctx\" in function \"handle\" has type context.Context and must be the first parameter\n\n" +
		"──────────────────\n"

	if got := buf.String(); got != want {
		t.Fatalf("unexpected rendered issue block:\n%s", got)
	}
}

func TestIssueRendererMarksFixableIssues(t *testing.T) {
	var buf bytes.Buffer

	render.SetNoColor(true)
	defer render.SetNoColor(false)

	r := issueRenderer{
		cwd: "",
		out: &buf,
	}

	r.write(rules.Issue{
		ID:       rules.RedundantImportAliasID,
		Path:     "sample.go",
		Line:     4,
		Column:   2,
		Flags:    rules.IssueFixableFlag,
		Severity: rules.SeverityWarn,
	}, `import alias "fmt" is redundant for package "fmt"`)

	if got := buf.String(); !bytes.Contains([]byte(got), []byte("[fixable]")) {
		t.Fatalf("expected fixable marker in output, got %q", got)
	}
}

func TestIssueRendererMarksUnsafeFixableIssues(t *testing.T) {
	var buf bytes.Buffer

	render.SetNoColor(true)
	defer render.SetNoColor(false)

	r := issueRenderer{
		cwd: "",
		out: &buf,
	}

	r.write(rules.Issue{
		ID:       rules.UseContextInFirstParamID,
		Path:     "sample.go",
		Line:     4,
		Column:   2,
		Flags:    rules.IssueUnsafeFixableFlag,
		Severity: rules.SeverityWarn,
	}, `parameter "ctx" in function "handle" has type context.Context and must be the first parameter`)

	if got := buf.String(); !bytes.Contains([]byte(got), []byte("[unsafe fix]")) {
		t.Fatalf("expected unsafe fix marker in output, got %q", got)
	}
}

func TestIssueRendererHighlightsLongLineOverflow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")
	line := "var sentence = \"" + strings.Repeat("a", 120) + "\"\n"

	if err := os.WriteFile(path, []byte("package sample\n"+line), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var buf bytes.Buffer

	render.SetNoColor(true)
	defer render.SetNoColor(false)

	r := issueRenderer{
		cwd:         dir,
		out:         &buf,
		sourceCache: make(map[string][]string, 1),
	}

	r.write(rules.Issue{
		ID:       rules.MaxLineLengthID,
		Path:     path,
		Line:     2,
		Column:   1,
		ArgInt1:  100,
		ArgInt2:  138,
		Severity: rules.SeverityWarn,
	}, `line has 138 characters; limit is 100`)

	got := buf.String()

	if !bytes.Contains([]byte(got), []byte("^^^^^^^^^^^^…")) {
		t.Fatalf("expected overflow highlight marker, got %q", got)
	}

	if !bytes.Contains([]byte(got), []byte("…")) {
		t.Fatalf("expected cropped long line to include ellipsis, got %q", got)
	}
}

func TestSummaryErrorWritesFooterWithoutCommandPrefix(t *testing.T) {
	summary := issueSummary{
		hasIssues: true,
		errors:    1,
		fixables:  1,
	}

	err := summary.err()

	var buf bytes.Buffer

	render.SetNoColor(true)
	defer render.SetNoColor(false)

	exception.Write(&buf, err, true)

	if got := buf.String(); got != "1 issue found (1 error), 1 issue is fixable. Use --write to apply automatic fixes.\n" {
		t.Fatalf("unexpected summary rendering: %q", got)
	}
}

func TestSummaryErrorMentionsUnsafeFixes(t *testing.T) {
	summary := issueSummary{
		hasIssues:      true,
		warnings:       1,
		unsafeFixables: 1,
	}

	err := summary.err()
	if !errors.Is(err, exception.ErrCommand) {
		t.Fatalf("expected command error, got %v", err)
	}

	if got := exception.Message(err); got != "1 issue found (1 warning), 1 issue requires --write --unsafe to apply automatic fixes" {
		t.Fatalf("unexpected unsafe summary message: %q", got)
	}
}
