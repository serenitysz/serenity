package naming

import (
	"go/ast"
	"regexp"

	"github.com/serenitysz/serenity/internal/rules"
)

type ImportedIdentifiersRule struct {
	re *regexp.Regexp
}

func (r *ImportedIdentifiersRule) Name() string {
	return "imported-identifiers"
}

func (r *ImportedIdentifiersRule) Targets() []ast.Node {
	return []ast.Node{(*ast.ImportSpec)(nil)}
}

func (r *ImportedIdentifiersRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	naming := runner.Cfg.Linter.Rules.Naming

	if naming == nil || !naming.Use || naming.ImportedIdentifiers == nil {
		return
	}

	if max := runner.Cfg.GetMaxIssues(); max > 0 && *runner.IssuesCount >= max {
		return
	}

	if r.re == nil {
		re, err := regexp.Compile(*naming.ExportedIdentifiers.Pattern)

		if err != nil {
			return
		}

		r.re = re
	}

	spec := node.(*ast.ImportSpec)
	name := spec.Name

	if name != nil && !r.re.MatchString(name.Name) {
		*runner.IssuesCount++

		*runner.Issues = append(*runner.Issues, rules.Issue{
			ArgStr1:  name.Name,
			ID:       rules.ImportedIdentifiersID,
			Pos:      runner.Fset.Position(spec.Pos()),
			Severity: rules.ParseSeverity(naming.ImportedIdentifiers.Severity),
		})
	}
}
