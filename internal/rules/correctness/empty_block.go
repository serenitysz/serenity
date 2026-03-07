package correctness

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

type EmptyBlockRule struct {
	Severity rules.Severity
}

func (r *EmptyBlockRule) Name() string {
	return "empty-block"
}

func (r *EmptyBlockRule) Targets() []ast.Node {
	return []ast.Node{(*ast.BlockStmt)(nil)}
}

func (r *EmptyBlockRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	if runner.ReachedMax() {
		return
	}

	spec := node.(*ast.BlockStmt)

	if len(spec.List) == 0 {
		runner.Report(spec.Pos(), rules.Issue{
			ArgStr1:  rules.CurrentFunctionName(runner),
			ID:       rules.EmptyBlockID,
			Severity: r.Severity,
		})
	}
}
