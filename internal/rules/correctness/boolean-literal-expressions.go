package correctness

import (
	"go/ast"
	"go/token"

	"github.com/serenitysz/serenity/internal/rules"
)

type BooleanLiteralExpressionsRule struct {
	Severity rules.Severity
}

func (r *BooleanLiteralExpressionsRule) Name() string {
	return "boolean-literal-expressions"
}

func (r *BooleanLiteralExpressionsRule) Targets() []ast.Node {
	return []ast.Node{(*ast.BinaryExpr)(nil)}
}

func (r *BooleanLiteralExpressionsRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	if runner.ReachedMax() {
		return
	}

	expr := node.(*ast.BinaryExpr)

	if expr.Op != token.EQL && expr.Op != token.NEQ {
		return
	}

	if !hasBoolLiteral(expr.X) && !hasBoolLiteral(expr.Y) {
		return
	}

	runner.Report(expr.Pos(), rules.Issue{
		ID:       rules.BoolLiteralExpressionsID,
		Severity: r.Severity,
	})
}

func hasBoolLiteral(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)

	return ok && (ident.Name == "true" || ident.Name == "false")
}
