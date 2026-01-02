package imports

import (
	"go/ast"
	"slices"
	"strings"

	"github.com/serenitysz/serenity/internal/rules"
)

type DisallowedPackagesRule struct{}

func (r *DisallowedPackagesRule) Name() string {
	return "disallowed-packages"
}

func (r *DisallowedPackagesRule) Targets() []ast.Node {
	return []ast.Node{(*ast.ImportSpec)(nil)}
}

func (r *DisallowedPackagesRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	imports := runner.Cfg.Linter.Rules.Imports

	if imports == nil || (imports.Use != nil && !*imports.Use) || imports.DisallowedPackages == nil {
		return
	}

	maxIssues := rules.GetMaxIssues(runner.Cfg)

	if maxIssues > 0 && *runner.IssuesCount >= maxIssues {
		return
	}

	spec := node.(*ast.ImportSpec)
	path := strings.Trim(spec.Path.Value, `"`)

	if slices.Contains(imports.DisallowedPackages.Packages, path) {
		*runner.IssuesCount++

		*runner.Issues = append(*runner.Issues, rules.Issue{
			ArgStr1:  path,
			ID:       rules.DisallowedPackagesID,
			Pos:      runner.Fset.Position(spec.Path.ValuePos),
			Severity: rules.ParseSeverity(imports.DisallowedPackages.Severity),
		})
	}
}
