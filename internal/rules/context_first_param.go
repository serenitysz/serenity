package rules

import (
	"go/ast"
	"go/token"
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

func CheckContextFirstParam(
	f *ast.File,
	fset *token.FileSet,
	out []Issue,
	cfg *LinterOptions,
) []Issue {
	if cfg.Linter.Use != nil && !*cfg.Linter.Use {
		return out
	}

	var maxIssues int8
	if cfg.Linter.Issues != nil && cfg.Linter.Issues.Max != nil {
		maxIssues = *cfg.Linter.Issues.Max
	}

	if int8(len(out)) >= maxIssues {
		return out
	}

	isBestPracticesOn := cfg.Linter.Rules.BestPractices == nil || (cfg.Linter.Rules.BestPractices.Use != nil && !*cfg.Linter.Rules.BestPractices.Use)

	isRuleDisabled := cfg.Linter.Rules.BestPractices.UseContextInFirstParam == nil

	if isBestPracticesOn || isRuleDisabled {
		return out
	}

	ast.Inspect(f, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		params := fn.Type.Params
		if params == nil || len(params.List) < 2 {
			return true
		}

		for i := 1; i < len(params.List); i++ {
			p := params.List[i]

			if isContextType(p.Type) {
				if int8(len(out)) >= maxIssues {
					return false
				}
				out = append(out, Issue{
					Pos:     fset.Position(p.Pos()),
					Message: "context.Context should be the first parameter",
					Fix: func() {
						FixContextFirstParam(fn)
					},
				})

				break
			}
		}

		return true
	})

	return out
}

func FixContextFirstParam(fn *ast.FuncDecl) bool {
	params := fn.Type.Params

	if params == nil || len(params.List) < 2 {
		return false
	}

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
