package bestpractices

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

type AlwaysPreferConstRule struct {
	Severity rules.Severity
}

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

	if runner.ReachedMax() {
		return
	}

	v := node.(*ast.ValueSpec)

	if len(v.Values) == 0 || len(runner.ConstCandidates) == 0 {
		return
	}

	for i, name := range v.Names {
		if i >= len(v.Values) {
			break
		}

		if _, ok := v.Values[i].(*ast.BasicLit); !ok {
			continue
		}

		if _, ok := runner.ConstCandidates[name]; ok {
			runner.Report(name.Pos(), rules.Issue{
				Severity: a.Severity,
				ArgStr1:  name.Name,
				ID:       rules.AlwaysPreferConstID,
			})
		}
	}
}
