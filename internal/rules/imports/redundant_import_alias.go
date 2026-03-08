package imports

import (
	"go/ast"
	"strings"

	"github.com/serenitysz/serenity/internal/rules"
)

type RedundantImportAliasRule struct {
	Severity rules.Severity
}

func (r *RedundantImportAliasRule) Name() string {
	return "redundant-import-alias"
}

func (r *RedundantImportAliasRule) Targets() []ast.Node {
	return []ast.Node{(*ast.ImportSpec)(nil)}
}

func (r *RedundantImportAliasRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	if runner.ReachedMax() {
		return
	}

	spec := node.(*ast.ImportSpec)

	if spec.Name == nil {
		return
	}

	path := strings.Trim(spec.Path.Value, `"`)

	defaultName := defaultImportName(path)

	if spec.Name.Name == defaultName {
		pos := spec.Name.Pos()
		issue := rules.Issue{
			ArgStr1:  rules.PackContext2(spec.Name.Name, path),
			ID:       rules.RedundantImportAliasID,
			Severity: r.Severity,
		}

		if runner.ShouldAutofix() {
			spec.Name = nil
			runner.Modified = true
			runner.ReportFixed(pos, issue)
			return
		}

		runner.ReportFixable(pos, issue)
	}
}

func defaultImportName(path string) string {
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}

	return path
}
