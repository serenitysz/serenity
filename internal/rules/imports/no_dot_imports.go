package imports

import (
	"go/ast"
	"strings"

	"github.com/serenitysz/serenity/internal/rules"
)

type NoDotImportsRule struct {
	Severity rules.Severity
}

func (r *NoDotImportsRule) Name() string {
	return "no-dot-imports"
}

func (r *NoDotImportsRule) Targets() []ast.Node {
	return []ast.Node{(*ast.ImportSpec)(nil)}
}

func (r *NoDotImportsRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	if runner.ReachedMax() {
		return
	}

	importSpec := node.(*ast.ImportSpec)

	if importSpec.Name != nil && importSpec.Name.Name == "." {
		runner.Report(importSpec.Name.NamePos, rules.Issue{
			ArgStr1:  strings.Trim(importSpec.Path.Value, `"`),
			ID:       rules.NoDotImportsID,
			Severity: r.Severity,
		})
	}
}
