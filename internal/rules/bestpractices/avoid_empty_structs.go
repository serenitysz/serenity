package bestpractices

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

type AvoidEmptyStructsRule struct {
	Severity rules.Severity
}

func (a *AvoidEmptyStructsRule) Name() string {
	return "avoid-empty-structs"
}

func (a *AvoidEmptyStructsRule) Targets() []ast.Node {
	return []ast.Node{(*ast.TypeSpec)(nil)}
}

func (a *AvoidEmptyStructsRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	if runner.ReachedMax() {
		return
	}

	t := node.(*ast.TypeSpec)
	st, ok := t.Type.(*ast.StructType)

	if !ok {
		return
	}

	if st.Fields == nil || len(st.Fields.List) == 0 {
		runner.Report(st.Pos(), rules.Issue{
			ID:       rules.AvoidEmptyStructsID,
			Severity: a.Severity,
			ArgStr1:  t.Name.Name,
		})
	}
}
