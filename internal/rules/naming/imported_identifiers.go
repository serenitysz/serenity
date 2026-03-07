package naming

import (
	"go/ast"
	"regexp"
	"strings"

	"github.com/serenitysz/serenity/internal/rules"
)

type ImportedIdentifiersRule struct {
	Severity rules.Severity
	Re       *regexp.Regexp
}

func NewImportedIdentifiersRule(cfg *rules.AnyPatternBasedRule) *ImportedIdentifiersRule {
	rule := &ImportedIdentifiersRule{}

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

	if runner.ReachedMax() {
		return
	}

	spec := node.(*ast.ImportSpec)
	name := spec.Name

	if name != nil && (r.Re == nil || !r.Re.MatchString(name.Name)) {
		runner.Report(spec.Pos(), rules.Issue{
			ArgStr1:  rules.PackContext2(name.Name, strings.Trim(spec.Path.Value, `"`)),
			ID:       rules.ImportedIdentifiersID,
			Severity: r.Severity,
		})
	}
}
