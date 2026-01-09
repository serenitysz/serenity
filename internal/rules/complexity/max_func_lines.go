package complexity

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

type CheckMaxFuncLinesRule struct{}

func (c *CheckMaxFuncLinesRule) Name() string {
	return "max-func-lines"
}

func (c *CheckMaxFuncLinesRule) Targets() []ast.Node {
	return []ast.Node{(*ast.FuncDecl)(nil)}
}

func (c *CheckMaxFuncLinesRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	if max := runner.Cfg.GetMaxIssues(); max > 0 && *runner.IssuesCount >= max {
		return
	}

	complexity := runner.Cfg.Linter.Rules.Complexity

	if complexity == nil || !complexity.Use {
		return
	}

	ruleConfig := complexity.MaxFuncLines

	if ruleConfig == nil {
		return
	}

	fn := node.(*ast.FuncDecl)

	if fn.Body == nil {
		return
	}

	var limit int16 = 20

	if ruleConfig.Max != nil {
		limit = int16(*ruleConfig.Max)
	}

	end := runner.Fset.Position(fn.End()).Line
	start := runner.Fset.Position(fn.Pos()).Line

	linesCount := end - start + 1

	if int16(linesCount) <= limit {
		return
	}

	*runner.IssuesCount++

	*runner.Issues = append(*runner.Issues, rules.Issue{
		ArgInt1:  int(limit),
		ArgInt2:  linesCount,
		ID:       rules.MaxFuncLinesID,
		Pos:      runner.Fset.Position(fn.Pos()),
		Severity: rules.ParseSeverity(ruleConfig.Severity),
	})
}
