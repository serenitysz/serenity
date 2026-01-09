package naming

import (
	"go/ast"
	"regexp"

	"github.com/serenitysz/serenity/internal/rules"
)

type ExportedIdentifiersRule struct {
	re *regexp.Regexp
}

func (r *ExportedIdentifiersRule) Name() string {
	return "exported-identifiers"
}

func (r *ExportedIdentifiersRule) Targets() []ast.Node {
	return []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.TypeSpec)(nil),
		(*ast.ValueSpec)(nil),
		(*ast.Field)(nil),
	}
}

func (r *ExportedIdentifiersRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	cfg := runner.Cfg.Linter.Rules.Naming

	if cfg == nil || !cfg.Use || cfg.ExportedIdentifiers == nil {
		return
	}

	if max := runner.Cfg.GetMaxIssues(); max > 0 && *runner.IssuesCount >= max {
		return
	}

	if r.re == nil {
		re, err := regexp.Compile(*cfg.ExportedIdentifiers.Pattern)

		if err != nil {
			return
		}

		r.re = re
	}

	check := func(id *ast.Ident) {
		if id == nil || !ast.IsExported(id.Name) || r.re.MatchString(id.Name) {
			return
		}

		*runner.IssuesCount++
		*runner.Issues = append(*runner.Issues, rules.Issue{
			ArgStr1:  id.Name,
			ID:       rules.ExportedIdentifiersID,
			Pos:      runner.Fset.Position(id.NamePos),
			Severity: rules.ParseSeverity(cfg.ExportedIdentifiers.Severity),
		})
	}

	switch n := node.(type) {
	case *ast.FuncDecl:
		check(n.Name)

	case *ast.TypeSpec:
		check(n.Name)

	case *ast.ValueSpec:
		for _, name := range n.Names {
			check(name)
		}

	case *ast.Field:
		for _, name := range n.Names {
			check(name)
		}
	}
}
