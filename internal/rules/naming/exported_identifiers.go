package naming

import (
	"go/ast"
	"regexp"
	"strings"

	"github.com/serenitysz/serenity/internal/rules"
)

type ExportedIdentifiersRule struct {
	Severity rules.Severity
	Re       *regexp.Regexp
}

func NewExportedIdentifiersRule(cfg *rules.AnyPatternBasedRule) *ExportedIdentifiersRule {
	rule := &ExportedIdentifiersRule{}

	if cfg != nil {
		rule.Severity = rules.ParseSeverity(cfg.Severity)
	}

	if cfg != nil && cfg.Pattern != nil {
		if re, err := regexp.Compile(*cfg.Pattern); err == nil {
			rule.Re = re
		}
	}

	return rule
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

	if runner.ReachedMax() {
		return
	}

	check := func(kind string, id *ast.Ident) {
		if id == nil || !ast.IsExported(id.Name) || (r.Re != nil && r.Re.MatchString(id.Name)) {
			return
		}

		runner.Report(id.NamePos, rules.Issue{
			ArgStr1:  rules.PackContext2(kind, id.Name),
			ID:       rules.ExportedIdentifiersID,
			Severity: r.Severity,
		})
	}

	switch n := node.(type) {
	case *ast.FuncDecl:
		check("function", n.Name)

	case *ast.TypeSpec:
		check("type", n.Name)

	case *ast.ValueSpec:
		kind := "value"
		if decl, ok := runner.Parent.(*ast.GenDecl); ok {
			kind = strings.ToLower(decl.Tok.String())
		}

		for _, name := range n.Names {
			check(kind, name)
		}

	case *ast.Field:
		for _, name := range n.Names {
			check("field", name)
		}
	}
}
