package bestpractices

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

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

func CheckContextFirstParamNode(runner *rules.Runner) []rules.Issue {
	bestPractices := runner.Cfg.Linter.Rules.BestPractices

	if bestPractices == nil {
		return nil
	}

	if bestPractices.Use != nil && !*bestPractices.Use {
		return nil
	}

	if bestPractices.UseContextInFirstParam == nil {
		return nil
	}

	fn, ok := runner.Node.(*ast.FuncDecl)
	if !ok {
		return nil
	}

	params := fn.Type.Params
	if params == nil || len(params.List) < 2 {
		return nil
	}

	for i := 1; i < len(params.List); i++ {
		p := params.List[i]

		if isContextType(p.Type) {
			return []rules.Issue{{
				Pos:     runner.Fset.Position(p.Pos()),
				Message: "context.Context should be the first parameter",
				Fix: func() {
					FixContextFirstParam(fn)
				},
			}}
		}
	}

	return nil
}

func FixContextFirstParam(fn *ast.FuncDecl) bool {
	params := fn.Type.Params
	ctxIndex := findContextParam(params.List)

	if ctxIndex <= 0 {
		return false
	}

	ctxField := params.List[ctxIndex]

	copy(
		params.List[1:ctxIndex+1],
		params.List[0:ctxIndex],
	)

	params.List[0] = ctxField

	return true
}

func findContextParam(fields []*ast.Field) int {
	for i, f := range fields {
		if isContextType(f.Type) {
			return i
		}
	}

	return -1
}
