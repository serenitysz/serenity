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

	issue := rules.Issue{
		ArgStr1:  rules.PackContext2(rules.ExprName(stmt.Lhs[0]), rules.CurrentFunctionName(runner)),
		ID:       rules.PreferIncDecID,
		Severity: r.Severity,
	}

	if runner.ShouldAutofix() {
		var op token.Token

		switch stmt.Tok {
		case token.ADD_ASSIGN:
			op = token.INC
		case token.SUB_ASSIGN:
			op = token.DEC
		default:
			runner.Report(stmt.Pos(), issue)
			return
		}

		replacement := &ast.IncDecStmt{
			X:      stmt.Lhs[0],
			TokPos: stmt.TokPos,
			Tok:    op,
		}

		if rules.ReplaceNode(runner.Parent, stmt, replacement) {
			runner.Modified = true
			runner.ReportFixed(stmt.Pos(), issue)
			return
		}
	}

	runner.ReportFixable(stmt.Pos(), issue)
}
