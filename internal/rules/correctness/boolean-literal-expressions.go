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

	issue := rules.Issue{
		ArgStr1:  rules.CurrentFunctionName(runner),
		ID:       rules.BoolLiteralExpressionsID,
		Severity: r.Severity,
	}

	if runner.ShouldAutofix() {
		replacement, ok := simplifyBoolLiteralExpression(expr)
		if ok && rules.ReplaceNode(runner.Parent, expr, replacement) {
			runner.Modified = true
			runner.ReportFixed(expr.Pos(), issue)
			return
		}
	}

	runner.ReportFixable(expr.Pos(), issue)
}

func hasBoolLiteral(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)

	return ok && (ident.Name == "true" || ident.Name == "false")
}

func simplifyBoolLiteralExpression(expr *ast.BinaryExpr) (ast.Expr, bool) {
	leftValue, leftBool := boolLiteralValue(expr.X)
	rightValue, rightBool := boolLiteralValue(expr.Y)

	switch {
	case leftBool && rightBool:
		value := leftValue == rightValue
		if expr.Op == token.NEQ {
			value = !value
		}

		if value {
			return ast.NewIdent("true"), true
		}

		return ast.NewIdent("false"), true
	case leftBool:
		return simplifyBoolLiteralOperand(expr.Op, expr.Y, leftValue), true
	case rightBool:
		return simplifyBoolLiteralOperand(expr.Op, expr.X, rightValue), true
	default:
		return nil, false
	}
}

func simplifyBoolLiteralOperand(op token.Token, expr ast.Expr, literal bool) ast.Expr {
	if (op == token.EQL && literal) || (op == token.NEQ && !literal) {
		return expr
	}

	return &ast.UnaryExpr{
		OpPos: expr.Pos(),
		Op:    token.NOT,
		X:     expr,
	}
}

func boolLiteralValue(expr ast.Expr) (bool, bool) {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false, false
	}

	switch ident.Name {
	case "true":
		return true, true
	case "false":
		return false, true
	default:
		return false, false
	}
}
