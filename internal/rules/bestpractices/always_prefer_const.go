package bestpractices

import (
	"go/ast"
	"go/types"

	"github.com/serenitysz/serenity/internal/rules"
)

type AlwaysPreferConstRule struct{}

func (a *AlwaysPreferConstRule) Name() string {
	return "always-prefer-const"
}

func (a *AlwaysPreferConstRule) Targets() []ast.Node {
	return []ast.Node{(*ast.ValueSpec)(nil)}
}

func (a *AlwaysPreferConstRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	v := node.(*ast.ValueSpec)
	if len(v.Values) == 0 {
		return
	}

	bp := runner.Cfg.Linter.Rules.BestPractices
	if bp == nil || (bp.Use != nil && !*bp.Use) || bp.AlwaysPreferConst == nil || (bp.AlwaysPreferConst.Use != nil && !*bp.AlwaysPreferConst.Use) {
		return
	}

	maxIssues := rules.GetMaxIssues(runner.Cfg)
	severity := rules.ParseSeverity(bp.AlwaysPreferConst.Severity)

	for i, name := range v.Names {
		obj := runner.TypesInfo.Defs[name]

		if obj == nil {
			continue
		}

		if _, ok := obj.(*types.Var); !ok {
			continue
		}

		if i >= len(v.Values) {
			break
		}

		if _, ok := v.Values[i].(*ast.BasicLit); !ok {
			continue
		}

		if !runner.MutatedObjects[obj] {
			if maxIssues > 0 && int16(len(*runner.Issues)) >= maxIssues {
				return
			}

			*runner.Issues = append(*runner.Issues, rules.Issue{
				ID:       rules.AlwaysPreferConstID,
				Pos:      runner.Fset.Position(name.Pos()),
				Severity: severity,
				ArgStr1:  name.Name,
			})
		}

	}
}
