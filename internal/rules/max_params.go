package rules

import (
	"go/ast"
	"go/token"
)

func CheckMaxParams(
	f *ast.File,
	fset *token.FileSet,
	out []Issue,
	cfg *LinterOptions,
) []Issue {
	if cfg.Linter.Use != nil && !*cfg.Linter.Use {
		return out
	}

	var limit int8 = 5
	var maxIssues int8
	if cfg.Linter.Issues != nil && cfg.Linter.Issues.Max != nil {
		maxIssues = *cfg.Linter.Issues.Max
	}

	if int8(len(out)) >= maxIssues {
		return out
	}

	if cfg.Linter.Rules.BestPractices.MaxParams != nil && cfg.Linter.Rules.BestPractices.MaxParams.Quantity != nil {
		limit = *cfg.Linter.Rules.BestPractices.MaxParams.Quantity
	}

	ast.Inspect(f, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		params := fn.Type.Params
		if params == nil {
			return true
		}

		count := 0

		for _, field := range params.List {
			count += len(field.Names)

			if len(field.Names) == 0 {
				count++
			}
		}

		if int8(count) <= limit {
			return true
		}

		out = append(out, Issue{
			Pos:     fset.Position(fn.Pos()),
			Message: "functions exceed the maximum parameter limit",
			Fix: func() {
				// Unsafe
				params.List = params.List[:limit]
			},
		})

		return true
	})

	return out
}

