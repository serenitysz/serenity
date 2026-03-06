package style

import (
	"go/ast"
	"go/token"

	"github.com/serenitysz/serenity/internal/rules"
)

type PreferIncDecRule struct {
	Severity rules.Severity
}

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

	if runner.ReachedMax() {
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

	runner.Report(stmt.Pos(), rules.Issue{
		ID:       rules.PreferIncDecID,
		Severity: r.Severity,
	})
}
