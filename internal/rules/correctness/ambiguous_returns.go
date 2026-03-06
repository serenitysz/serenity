package correctness

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

type AmbiguousReturnRule struct {
	Severity   rules.Severity
	MaxAllowed int
}

func (r *AmbiguousReturnRule) Name() string {
	return "ambiguous-return"
}

func (r *AmbiguousReturnRule) Targets() []ast.Node {
	return []ast.Node{(*ast.FuncDecl)(nil)}
}

func (r *AmbiguousReturnRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	if runner.ReachedMax() {
		return
	}

	fn := node.(*ast.FuncDecl)

	if fn.Type.Results == nil {
		return
	}

	results := fn.Type.Results.List

	if len(results) < 2 {
		return
	}

	seen := make(map[string]int, len(results))
	hasUnnamed := false

	for _, field := range results {
		if len(field.Names) > 0 {
			return
		}

		hasUnnamed = true
		seen[typeKey(field.Type)]++
	}

	if !hasUnnamed {
		return
	}

	for typ, count := range seen {
		if count > r.MaxAllowed && !isAllowedDuplicateReturn(typ) {
			runner.Report(fn.Type.Results.Pos(), rules.Issue{
				ArgInt1:  count,
				ArgInt2:  r.MaxAllowed,
				ID:       rules.AmbiguousReturnID,
				Severity: r.Severity,
			})

			return
		}
	}
}

func typeKey(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return t.X.(*ast.Ident).Name + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + typeKey(t.X)
	default:
		return ""
	}
}

func isAllowedDuplicateReturn(typ string) bool {
	return typ == "error"
}
