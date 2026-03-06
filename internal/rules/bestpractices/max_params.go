package bestpractices

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

type MaxParamsRule struct {
	Limit    uint16
	Severity rules.Severity
}

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

	if runner.ReachedMax() {
		return
	}

	fn := node.(*ast.FuncDecl)

	if fn.Type.Params == nil {
		return
	}

	var count uint16

	for _, field := range fn.Type.Params.List {
		if len(field.Names) == 0 {
			count++

			continue
		}

		count += uint16(len(field.Names))
	}

	if r.Limit > 0 && count <= r.Limit {
		return
	}

	runner.Report(fn.Pos(), rules.Issue{
		ArgInt1:  int(r.Limit),
		ArgInt2:  int(count),
		ID:       rules.MaxParamsID,
		Severity: r.Severity,
	})
}
