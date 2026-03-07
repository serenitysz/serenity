package linter

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/serenitysz/serenity/internal/rules"
	"github.com/serenitysz/serenity/internal/utils"
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

func TestProcessPath_CacheInvalidatesOnContentChange(t *testing.T) {
	dir := t.TempDir()
	cacheDir := t.TempDir()
	path := filepath.Join(dir, "sample.go")

	t.Setenv("SERENITY_CACHE_DIR", cacheDir)

	src1 := `package sample

var report = 7

func keep() { _ = report }
`

	src2 := `package sample

var report = 7

func keep() { report++   }
`

	if len(src1) != len(src2) {
		t.Fatalf("fixture lengths must match: %d != %d", len(src1), len(src2))
	}

	if err := os.WriteFile(path, []byte(src1), 0o644); err != nil {
		t.Fatalf("write first fixture: %v", err)
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
		Performance: &rules.PerformanceOptions{
			Use:     true,
			Caching: utils.Ptr(true),
		},
	}

	l := New(false, false, cfg, 0, 0)

	firstIssues, err := l.ProcessPath(path)
	if err != nil {
		t.Fatalf("first ProcessPath failed: %v", err)
	}

	if got := countIssuesByID(firstIssues, rules.AlwaysPreferConstID); got != 1 {
		t.Fatalf("expected 1 always-prefer-const issue before mutation, got %d", got)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat fixture: %v", err)
	}

	stamp := info.ModTime().Round(0)
	if err := os.WriteFile(path, []byte(src2), 0o644); err != nil {
		t.Fatalf("write second fixture: %v", err)
	}
	if err := os.Chtimes(path, stamp, stamp); err != nil {
		t.Fatalf("restore times: %v", err)
	}

	secondIssues, err := l.ProcessPath(path)
	if err != nil {
		t.Fatalf("second ProcessPath failed: %v", err)
	}

	if got := countIssuesByID(secondIssues, rules.AlwaysPreferConstID); got != 0 {
		t.Fatalf("expected cache invalidation after content change, got %d issues", got)
	}
}

func TestProcessPath_CachePreservesSyntheticIssuePath(t *testing.T) {
	dir := t.TempDir()
	cacheDir := t.TempDir()
	path := filepath.Join(dir, "sample.go")

	t.Setenv("SERENITY_CACHE_DIR", cacheDir)

	src := `// @serenity-ignore-all no-dot-imports: test
package sample

func keep() {}
`

	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	cfg := &rules.LinterOptions{
		Linter: rules.LinterRules{
			Use:    true,
			Rules:  rules.LinterRulesGroup{},
			Issues: &rules.LinterIssuesOptions{},
		},
		Performance: &rules.PerformanceOptions{
			Use:     true,
			Caching: utils.Ptr(true),
		},
	}

	l := New(false, false, cfg, 0, 0)

	if _, err := l.ProcessPath(path); err != nil {
		t.Fatalf("warm-up ProcessPath failed: %v", err)
	}

	issues, err := l.ProcessPath(path)
	if err != nil {
		t.Fatalf("cached ProcessPath failed: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("expected 1 cached synthetic issue, got %d", len(issues))
	}

	if issues[0].ID != rules.UnusedSuppressionID {
		t.Fatalf("expected unused suppression warning, got issue id %d", issues[0].ID)
	}

	if issues[0].Filename() != path {
		t.Fatalf("expected cached issue path %q, got %q", path, issues[0].Filename())
	}
}

func TestProcessPath_CacheDoesNotRewriteOnMetadataOnlyChange(t *testing.T) {
	dir := t.TempDir()
	cacheDir := t.TempDir()
	path := filepath.Join(dir, "sample.go")

	t.Setenv("SERENITY_CACHE_DIR", cacheDir)

	src := `package sample

var report = 7
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
					AlwaysPreferConst: &rules.LinterBaseRule{
						Severity: "warn",
					},
				},
			},
			Issues: &rules.LinterIssuesOptions{},
		},
		Performance: &rules.PerformanceOptions{
			Use:     true,
			Caching: utils.Ptr(true),
		},
	}

	l := New(false, false, cfg, 0, 0)

	firstIssues, err := l.ProcessPath(path)
	if err != nil {
		t.Fatalf("first ProcessPath failed: %v", err)
	}

	if got := countIssuesByID(firstIssues, rules.AlwaysPreferConstID); got != 1 {
		t.Fatalf("expected 1 always-prefer-const issue, got %d", got)
	}

	probes, err := probePackageInputs([]string{path})
	if err != nil {
		t.Fatalf("probe cache path: %v", err)
	}

	entryPath, err := l.Cache.entryPath(probes)
	if err != nil {
		t.Fatalf("resolve cache entry path: %v", err)
	}

	before, err := os.Stat(entryPath)
	if err != nil {
		t.Fatalf("stat cache entry: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	nextTime := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(path, nextTime, nextTime); err != nil {
		t.Fatalf("touch source file: %v", err)
	}

	secondIssues, err := l.ProcessPath(path)
	if err != nil {
		t.Fatalf("second ProcessPath failed: %v", err)
	}

	if got := countIssuesByID(secondIssues, rules.AlwaysPreferConstID); got != 1 {
		t.Fatalf("expected cache hit after metadata-only change, got %d issues", got)
	}

	after, err := os.Stat(entryPath)
	if err != nil {
		t.Fatalf("stat cache entry after touch: %v", err)
	}

	if !before.ModTime().Equal(after.ModTime()) {
		t.Fatalf("expected cache entry to be reused without rewrite, before=%v after=%v", before.ModTime(), after.ModTime())
	}
}

func countIssuesByID(issues []rules.Issue, id uint16) int {
	count := 0

	for _, issue := range issues {
		if issue.ID == id {
			count++
		}
	}

	return count
}
