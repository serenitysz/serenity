package bestpractices

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

type NoBareReturnsRule struct {
	Severity rules.Severity
}

func (n *NoBareReturnsRule) Name() string {
	return "no-bare-returns"
}

func (n *NoBareReturnsRule) Targets() []ast.Node {
	return []ast.Node{(*ast.ReturnStmt)(nil)}
}

func (n *NoBareReturnsRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	if runner.ReachedMax() || runner.CurrentFunc == nil || !runner.CurrentFunc.HasNamedResults {
		return
	}

	ret := node.(*ast.ReturnStmt)
	if len(ret.Results) > 0 {
		return
	}

	runner.Report(ret.Pos(), rules.Issue{
		ID:       rules.NoBareReturnsID,
		Severity: n.Severity,
		ArgStr1:  runner.CurrentFunc.Name,
	})
}
