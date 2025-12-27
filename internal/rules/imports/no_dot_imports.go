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

	importSpec := node.(*ast.ImportSpec)

	imports := runner.Cfg.Linter.Rules.Imports
	if imports == nil || (imports.Use != nil && !*imports.Use) {
		return
	}

	if imports.NoDotImports == nil {
		return
	}

	if importSpec.Name != nil && importSpec.Name.Name == "." {
		maxIssues := rules.GetMaxIssues(runner.Cfg)
		if maxIssues > 0 && int16(len(*runner.Issues)) >= maxIssues {
			return
		}

		*runner.Issues = append(*runner.Issues, rules.Issue{
			ID:       rules.NoDotImportsID,
			Pos:      runner.Fset.Position(importSpec.Name.NamePos),
			Severity: rules.ParseSeverity(imports.NoDotImports.Severity),
		})
	}
}

