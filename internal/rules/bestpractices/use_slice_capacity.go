package bestpractices

import (
	"go/ast"
	"go/token"

	"github.com/serenitysz/serenity/internal/rules"
)

type UseSliceCapacityRule struct {
	Severity rules.Severity
}

func (u *UseSliceCapacityRule) Name() string {
	return "use-slice-capacity"
}

func (u *UseSliceCapacityRule) Targets() []ast.Node {
	return []ast.Node{(*ast.CallExpr)(nil)}
}

func (u *UseSliceCapacityRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	if runner.ReachedMax() {
		return
	}

	call := node.(*ast.CallExpr)
	ident, ok := call.Fun.(*ast.Ident)

	if !ok || ident.Name != "make" || len(call.Args) == 0 {
		return
	}

	if _, ok := call.Args[0].(*ast.ArrayType); !ok {
		return
	}

	if len(call.Args) == 2 {
		issue := rules.Issue{
			ArgStr1:  rules.PackContext2(rules.AssignmentTargetName(runner.Parent, call), rules.CurrentFunctionName(runner)),
			ID:       rules.UseSliceCapacityID,
			Severity: u.Severity,
		}

		if runner.ShouldAutofix() && isSafeCapacityExpr(call.Args[1]) {
			call.Args = append(call.Args, cloneCapacityExpr(call.Args[1]))
			runner.Modified = true
			runner.ReportFixed(call.Pos(), issue)
			return
		}

		if isSafeCapacityExpr(call.Args[1]) {
			runner.ReportFixable(call.Pos(), issue)
			return
		}

		runner.Report(call.Pos(), issue)
	}
}

func isSafeCapacityExpr(expr ast.Expr) bool {
	switch n := expr.(type) {
	case *ast.Ident, *ast.BasicLit:
		return true
	case *ast.SelectorExpr:
		return isSafeCapacityExpr(n.X)
	case *ast.CallExpr:
		ident, ok := n.Fun.(*ast.Ident)
		if !ok || ident.Name != "len" || len(n.Args) != 1 || n.Ellipsis != token.NoPos {
			return false
		}

		return isSafeCapacityExpr(n.Args[0])
	default:
		return false
	}
}

func cloneCapacityExpr(expr ast.Expr) ast.Expr {
	switch n := expr.(type) {
	case *ast.Ident:
		return ast.NewIdent(n.Name)
	case *ast.BasicLit:
		return &ast.BasicLit{
			ValuePos: n.ValuePos,
			Kind:     n.Kind,
			Value:    n.Value,
		}
	case *ast.SelectorExpr:
		return &ast.SelectorExpr{
			X: cloneCapacityExpr(n.X),
			Sel: &ast.Ident{
				NamePos: n.Sel.NamePos,
				Name:    n.Sel.Name,
			},
		}
	case *ast.CallExpr:
		args := make([]ast.Expr, len(n.Args))
		for i, arg := range n.Args {
			args[i] = cloneCapacityExpr(arg)
		}

		return &ast.CallExpr{
			Fun:      cloneCapacityExpr(n.Fun),
			Lparen:   n.Lparen,
			Args:     args,
			Ellipsis: token.NoPos,
			Rparen:   n.Rparen,
		}
	default:
		return expr
	}
}
