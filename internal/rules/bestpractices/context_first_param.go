package bestpractices

import (
	"go/ast"
	"go/token"

	"github.com/serenitysz/serenity/internal/rules"
)

type ContextFirstRule struct {
	Severity rules.Severity
}

func (c *ContextFirstRule) Name() string {
	return "context-first-param"
}

func (c *ContextFirstRule) Targets() []ast.Node {
	return []ast.Node{(*ast.FuncDecl)(nil)}
}

func (c *ContextFirstRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	if runner.ReachedMax() {
		return
	}

	fn := node.(*ast.FuncDecl)

	if fn.Type.Params == nil || len(fn.Type.Params.List) < 2 {
		return
	}

	params := fn.Type.Params.List
	offenders := make([]rules.Issue, 0, len(params)-1)
	positions := make([]token.Pos, 0, len(params)-1)

	for i := 1; i < len(params); i++ {
		p := params[i]

		if isContextType(p.Type) {
			if runner.ReachedMax() {
				break
			}

			paramName := ""
			if len(p.Names) > 0 {
				paramName = p.Names[0].Name
			}

			offenders = append(offenders, rules.Issue{
				ArgStr1:  rules.PackContext2(paramName, fn.Name.Name),
				ArgInt1:  uint32(i + 1),
				ID:       rules.UseContextInFirstParamID,
				Severity: c.Severity,
			})
			positions = append(positions, p.Pos())
		}
	}

	if len(offenders) == 0 {
		return
	}

	if runner.ShouldAutofixUnsafe() {
		reordered := reorderContextParams(params)
		if !sameFieldOrder(params, reordered) {
			fn.Type.Params.List = reordered
			runner.Modified = true
		}

		for i, issue := range offenders {
			runner.ReportFixedUnsafe(positions[i], issue)
		}
		return
	}

	for i, issue := range offenders {
		runner.ReportUnsafeFixable(positions[i], issue)
	}
}

func reorderContextParams(params []*ast.Field) []*ast.Field {
	reordered := make([]*ast.Field, 0, len(params))

	for _, param := range params {
		if isContextType(param.Type) {
			reordered = append(reordered, param)
		}
	}

	for _, param := range params {
		if !isContextType(param.Type) {
			reordered = append(reordered, param)
		}
	}

	return reordered
}

func sameFieldOrder(before, after []*ast.Field) bool {
	if len(before) != len(after) {
		return false
	}

	for i := range before {
		if before[i] != after[i] {
			return false
		}
	}

	return true
}

func isContextType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		x, ok := t.X.(*ast.Ident)

		return ok && x.Name == "context" && t.Sel.Name == "Context"
	case *ast.StarExpr:
		if sel, ok := t.X.(*ast.SelectorExpr); ok {
			x, ok := sel.X.(*ast.Ident)

			return ok && x.Name == "context" && sel.Sel.Name == "Context"
		}
	}

	return false
}
