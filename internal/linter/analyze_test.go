package linter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/serenitysz/serenity/internal/rules"
)

func TestProcessPath_ContextAwareRules(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")

	src := `package sample

func nested() (err error) {
	for i := 0; i < 1; i++ {
		func() {
			defer println(i)
		}()
	}

	return
}

func direct() {
	for i := 0; i < 1; i++ {
		defer println(i)
	}
}
`

	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	cfg := &rules.LinterOptions{
		Linter: rules.LinterRules{
			Use: true,
			Rules: rules.LinterRulesGroup{
				BestPractices: &rules.BestPracticesRulesGroup{
					Use: true,
					NoBareReturns: &rules.LinterBaseRule{
						Severity: "warn",
					},
					NoDeferInLoop: &rules.LinterBaseRule{
						Severity: "warn",
					},
				},
			},
			Issues: &rules.LinterIssuesOptions{},
		},
	}

	l := New(false, false, cfg, 0, 0)
	issues, err := l.ProcessPath(path)
	if err != nil {
		t.Fatalf("ProcessPath failed: %v", err)
	}

	counts := make(map[uint16]int, len(issues))

	for _, issue := range issues {
		counts[issue.ID]++
	}

	if got := counts[rules.NoBareReturnsID]; got != 1 {
		t.Fatalf("expected 1 no-bare-returns issue, got %d", got)
	}

	if got := counts[rules.NoDeferInLoopID]; got != 1 {
		t.Fatalf("expected 1 no-defer-in-loop issue, got %d", got)
	}
}

func TestProcessPath_AlwaysPreferConstAcrossFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	files := map[string]string{
		"a.go": `package sample

var keep = 7
var report = 9
`,
		"b.go": `package sample

func touch() {
	keep = 8
}
`,
	}

	for name, src := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	cfg := &rules.LinterOptions{
		Linter: rules.LinterRules{
			Use: true,
			Rules: rules.LinterRulesGroup{
				BestPractices: &rules.BestPracticesRulesGroup{
					Use: true,
					AlwaysPreferConst: &rules.LinterBaseRule{
						Severity: "warn",
					},
				},
			},
			Issues: &rules.LinterIssuesOptions{},
		},
	}

	l := New(false, false, cfg, 0, 0)
	issues, err := l.ProcessPath(dir)
	if err != nil {
		t.Fatalf("ProcessPath failed: %v", err)
	}

	var reported []string

	for _, issue := range issues {
		if issue.ID == rules.AlwaysPreferConstID {
			reported = append(reported, issue.ArgStr1)
		}
	}

	if len(reported) != 1 || reported[0] != "report" {
		t.Fatalf("expected only report to be flagged, got %v", reported)
	}
}
