package bestpractices

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

type AvoidEmptyStructsRule struct{}

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

	bp := runner.Cfg.Linter.Rules.BestPractices
	if bp == nil || (bp.Use != nil && !*bp.Use) || bp.AvoidEmptyStructs == nil {
		return
	}

	t := node.(*ast.TypeSpec)
	st, ok := t.Type.(*ast.StructType)
	if !ok {
		return
	}

	if st.Fields == nil || len(st.Fields.List) == 0 {
		maxIssues := rules.GetMaxIssues(runner.Cfg)

		if maxIssues > 0 && int16(len(*runner.Issues)) >= maxIssues {
			return
		}

		*runner.Issues = append(*runner.Issues, rules.Issue{
			ID:       rules.AvoidEmptyStructsID,
			Pos:      runner.Fset.Position(st.Pos()),
			Severity: rules.ParseSeverity(bp.AvoidEmptyStructs.Severity),
			ArgStr1:  t.Name.Name,
		})
	}
}
