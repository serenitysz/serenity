package style

import (
	"go/ast"
	"go/token"

	"github.com/serenitysz/serenity/internal/rules"
)

type PreferIncDecRule struct{}

func (r *PreferIncDecRule) Name() string {
	return "prefer-inc-dec"
}

func (r *PreferIncDecRule) Targets() []ast.Node {
	return []ast.Node{(*ast.AssignStmt)(nil)}
}

func (r *PreferIncDecRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	if max := runner.Cfg.GetMaxIssues(); max > 0 && *runner.IssuesCount >= max {
		return
	}

	style := runner.Cfg.Linter.Rules.Style
	
	if style == nil || !style.Use || style.PreferIncDec == nil {
		return
	}

	stmt := node.(*ast.AssignStmt)

	if len(stmt.Lhs) != 1 || len(stmt.Rhs) != 1 {
		return
	}

	lit, ok := stmt.Rhs[0].(*ast.BasicLit)

	if !ok || lit.Kind != token.INT || lit.Value != "1" {
		return
	}

	if _, ok := stmt.Lhs[0].(*ast.IndexExpr); ok {
		return
	}

	*runner.IssuesCount++

	*runner.Issues = append(*runner.Issues, rules.Issue{
		ID:       rules.PreferIncDecID,
		Pos:      runner.Fset.Position(stmt.Pos()),
		Severity: rules.ParseSeverity(style.PreferIncDec.Severity),
	})
}
