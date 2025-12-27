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

	fn := node.(*ast.FuncDecl)

	if fn.Body == nil {
		return
	}

	complexity := runner.Cfg.Linter.Rules.Complexity
	if complexity == nil {
		return
	}

	if complexity.Use != nil && !*complexity.Use {
		return
	}

	ruleConfig := complexity.MaxFuncLines
	if ruleConfig == nil {
		return
	}

	var limit int16 = 20
	if ruleConfig.Max != nil {
		limit = int16(*ruleConfig.Max)
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
		Severity: rules.ParseSeverity(ruleConfig.Severity),
		ArgInt1:  int(limit),
		ArgInt2:  lines,
	})
}

