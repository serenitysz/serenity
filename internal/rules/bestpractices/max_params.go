package bestpractices

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

type MaxParamsRule struct{}

func (r *MaxParamsRule) Name() string {
	return "max-params"
}

func (r *MaxParamsRule) Targets() []ast.Node {
	return []ast.Node{(*ast.FuncDecl)(nil)}
}

func (r *MaxParamsRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	if max := runner.Cfg.GetMaxIssues(); max > 0 && *runner.IssuesCount >= max {
		return
	}

	bp := runner.Cfg.Linter.Rules.BestPractices

	if bp == nil || !bp.Use {
		return
	}

	fn := node.(*ast.FuncDecl)

	if fn.Type.Params == nil {
		return
	}

	var limit uint16 = 5

	if bp.MaxParams != nil {
		limit = *bp.MaxParams.Max
	}

	var count uint16

	for _, field := range fn.Type.Params.List {
		if len(field.Names) == 0 {
			count++

			continue
		}

		count += uint16(len(field.Names))
	}

	if limit > 0 && count <= (limit) {
		return
	}

	*runner.IssuesCount++

	*runner.Issues = append(*runner.Issues, rules.Issue{
		ArgInt1: int(limit),
		ArgInt2: int(count),
		ID:      rules.MaxParamsID,
		Pos:     runner.Fset.Position(fn.Pos()),
	})
}
