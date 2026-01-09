package correctness

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

type EmptyBlockRule struct{}

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

	if max := runner.Cfg.GetMaxIssues(); max > 0 && *runner.IssuesCount >= max {
		return
	}

	correctness := runner.Cfg.Linter.Rules.Correctness

	if correctness == nil || !correctness.Use || correctness.EmptyBlock == nil {
		return
	}

	spec := node.(*ast.BlockStmt)

	if len(spec.List) == 0 {
		*runner.IssuesCount++

		*runner.Issues = append(*runner.Issues, rules.Issue{
			ID:       rules.EmptyBlockID,
			Pos:      runner.Fset.Position(spec.Pos()),
			Severity: rules.ParseSeverity(correctness.EmptyBlock.Severity),
		})
	}
}
