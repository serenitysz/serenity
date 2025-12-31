package bestpractices

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

type UseSliceCapacityRule struct{}

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

	bp := runner.Cfg.Linter.Rules.BestPractices
	if bp == nil || (bp.Use != nil && !*bp.Use) || bp.UseSliceCapacity == nil || (bp.UseSliceCapacity.Use != nil && !*bp.UseSliceCapacity.Use) {
		return
	}

	maxIssues := rules.GetMaxIssues(runner.Cfg)
	if maxIssues > 0 && int16(len(*runner.Issues)) >= maxIssues {
		return
	}

	call := node.(*ast.CallExpr)
	ident, ok := call.Fun.(*ast.Ident)
	if !ok || ident.Name != "make" {
		return
	}

	if len(call.Args) == 0 {
		return
	}

	if _, ok := call.Args[0].(*ast.ArrayType); !ok {
		return
	}

	if len(call.Args) == 2 {
		*runner.Issues = append(*runner.Issues, rules.Issue{
			ID:       rules.UseSliceCapacityID,
			Pos:      runner.Fset.Position(call.Pos()),
			Severity: rules.ParseSeverity(bp.UseSliceCapacity.Severity),
		})
	}
}
