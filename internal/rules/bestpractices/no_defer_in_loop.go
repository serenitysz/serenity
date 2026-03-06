package bestpractices

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

type NoDeferInLoopRule struct {
	Severity rules.Severity
}

func (d *NoDeferInLoopRule) Name() string {
	return "no-defer-in-loop"
}

func (d *NoDeferInLoopRule) Targets() []ast.Node {
	return []ast.Node{(*ast.DeferStmt)(nil)}
}

func (d *NoDeferInLoopRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	if runner.ReachedMax() || runner.LoopDepth == 0 {
		return
	}

	stmt := node.(*ast.DeferStmt)
	runner.Report(stmt.Pos(), rules.Issue{
		Severity: d.Severity,
		ID:       rules.NoDeferInLoopID,
	})
}
