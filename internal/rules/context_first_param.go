package rules

import (
	"go/ast"
	"go/token"
)

func isContextType(typeExpr ast.Expr) bool {
	var selector *ast.SelectorExpr

	switch t := typeExpr.(type) {
	case *ast.SelectorExpr:
		selector = t
	case *ast.StarExpr:
		if sel, ok := t.X.(*ast.SelectorExpr); ok {
			selector = sel
		}
	}

	if selector == nil {
		return false
	}

	if x, ok := selector.X.(*ast.Ident); ok {
		return x.Name == "context" && selector.Sel.Name == "Context"
	}

	return false
}

func CheckContextFirstParam(f *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	ast.Inspect(f, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if fn.Type == nil || fn.Type.Params == nil || len(fn.Type.Params.List) < 1 {
			return true
		}

		for i, param := range fn.Type.Params.List {
			if i == 0 {
				continue
			}

			if isContextType(param.Type) {
				issues = append(issues, Issue{
					Pos:     fset.Position(param.Pos()),
					Message: "context.Context should be the first parameter",
				})
			}

		}

		return true
	})

	return issues
}
