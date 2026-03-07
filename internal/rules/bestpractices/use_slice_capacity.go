package bestpractices

import (
	"go/ast"

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
		runner.Report(call.Pos(), rules.Issue{
			ArgStr1:  rules.PackContext2(rules.AssignmentTargetName(runner.Parent, call), rules.CurrentFunctionName(runner)),
			ID:       rules.UseSliceCapacityID,
			Severity: u.Severity,
		})
	}
}
