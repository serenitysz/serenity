package complexity

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

func CheckMaxFuncLinesNode(runner *rules.Runner) {
	if runner.Cfg.Linter.Rules == nil {
		return
	}

	complexity := runner.Cfg.Linter.Rules.Complexity
	if complexity == nil {
		return
	}

	if complexity.Use != nil && !*complexity.Use {
		return
	}

	var limit int16 = 20
	if complexity.MaxFuncLines != nil && complexity.MaxFuncLines.Max != nil {
		limit = int16(*complexity.MaxFuncLines.Max)
	}

	fn, ok := runner.Node.(*ast.FuncDecl)
	if !ok || fn.Body == nil {
		return
	}

	start := runner.Fset.Position(fn.Pos()).Line
	end := runner.Fset.Position(fn.End()).Line
	lines := end - start + 1

	if int16(lines) <= limit {
		return
	}

	maxIssues := rules.GetMaxIssues(runner.Cfg)
	if maxIssues > 0 && int16(len(*runner.Issues)) >= maxIssues {
		return
	}

	*runner.Issues = append(*runner.Issues, rules.Issue{
		ID:       rules.MaxFuncLinesID,
		Pos:      runner.Fset.Position(fn.Pos()),
		Severity: rules.ParseSeverity(complexity.MaxFuncLines.Severity),
		ArgInt1:  int(limit),
		ArgInt2:  lines,
	})
}
