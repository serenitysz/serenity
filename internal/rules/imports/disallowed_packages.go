package imports

import (
	"go/ast"
	"strings"

	"github.com/serenitysz/serenity/internal/rules"
)

type DisallowedPackagesRule struct {
	Severity rules.Severity
	Packages map[string]struct{}
}

func NewDisallowedPackagesRule(cfg *rules.DisallowedPackagesRule) *DisallowedPackagesRule {
	rule := &DisallowedPackagesRule{
		Packages: make(map[string]struct{}),
	}

	if cfg != nil {
		rule.Severity = rules.ParseSeverity(cfg.Severity)

		for _, pkg := range cfg.Packages {
			rule.Packages[pkg] = struct{}{}
		}
	}

	return rule
}

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

	if runner.ReachedMax() {
		return
	}

	spec := node.(*ast.ImportSpec)
	path := strings.Trim(spec.Path.Value, `"`)

	if _, blocked := r.Packages[path]; blocked {
		runner.Report(spec.Path.ValuePos, rules.Issue{
			ArgStr1:  path,
			ID:       rules.DisallowedPackagesID,
			Severity: r.Severity,
		})
	}
}
