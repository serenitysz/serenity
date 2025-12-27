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

	fn := node.(*ast.FuncDecl)

	if fn.Type.Params == nil {
		return
	}

	bp := runner.Cfg.Linter.Rules.BestPractices
	if bp == nil || (bp.Use != nil && !*bp.Use) {
		return
	}

	var limit int8 = 5
	if bp.MaxParams != nil && bp.MaxParams.Max != nil {
		limit = int8(*bp.MaxParams.Max)
	}

	var count int16 = 0
	maxIssues := rules.GetMaxIssues(runner.Cfg)

	for _, field := range fn.Type.Params.List {
		if len(field.Names) == 0 {
			count++
			continue
		}
		count += int16(len(field.Names))
	}

	if limit > 0 && count <= int16(limit) {
		return
	}

	if maxIssues > 0 && int16(len(*runner.Issues)) >= maxIssues {
		return
	}

	*runner.Issues = append(*runner.Issues, rules.Issue{
		ID:      rules.MaxParamsID,
		Pos:     runner.Fset.Position(fn.Pos()),
		ArgInt1: int(limit),
		ArgInt2: int(count),
	})
}
