package imports

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

type NoDotImportsRule struct{}

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

	if max := runner.Cfg.GetMaxIssues(); max > 0 && *runner.IssuesCount >= max {
		return
	}

	imports := runner.Cfg.Linter.Rules.Imports

	if imports == nil || !imports.Use || imports.NoDotImports == nil {
		return
	}

	importSpec := node.(*ast.ImportSpec)

	if importSpec.Name != nil && importSpec.Name.Name == "." {
		*runner.IssuesCount++

		*runner.Issues = append(*runner.Issues, rules.Issue{
			ID:       rules.NoDotImportsID,
			Pos:      runner.Fset.Position(importSpec.Name.NamePos),
			Severity: rules.ParseSeverity(imports.NoDotImports.Severity),
		})
	}
}
