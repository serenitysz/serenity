package bestpractices

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

type NoDeferInLoopRule struct{}

func (d *NoDeferInLoopRule) Name() string {
	return "no-defer-in-loop"
}

func (d *NoDeferInLoopRule) Targets() []ast.Node {
	return []ast.Node{
		(*ast.ForStmt)(nil),   // tradicional loops
		(*ast.RangeStmt)(nil), // loops into slices, maps, channels
	}
}

func (d *NoDeferInLoopRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	bp := runner.Cfg.Linter.Rules.BestPractices

	if bp == nil || !bp.Use || bp.NoDeferInLoop == nil {
		return
	}

	var body *ast.BlockStmt

	switch n := node.(type) {
	case *ast.RangeStmt:
		body = n.Body

	case *ast.ForStmt:
		body = n.Body
	}

	if body == nil {
		return
	}

	maxIssues := runner.Cfg.GetMaxIssues()
	severity := rules.ParseSeverity(bp.NoDeferInLoop.Severity)

	ast.Inspect(body, func(n ast.Node) bool {
		if maxIssues > 0 && *runner.IssuesCount >= maxIssues {
			return false
		}

		switch t := n.(type) {
		case *ast.FuncLit, *ast.RangeStmt, *ast.ForStmt:
			return false
		case *ast.DeferStmt:
			*runner.IssuesCount++
			*runner.Issues = append(*runner.Issues, rules.Issue{
				Severity: severity,
				ID:       rules.NoDeferInLoopID,
				Pos:      runner.Fset.Position(t.Pos()),
			})
		}

		return true
	})
}
