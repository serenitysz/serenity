package rules

import (
	"go/ast"
	"go/token"
)

func CheckNoDotImports(
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
	if !isBestPracticesOn {
		return out
	}

	for _, i := range f.Imports {
		if i.Name != nil && i.Name.Name == "." {
			out = append(out, Issue{
				Pos:     fset.Position(i.Name.NamePos),
				Message: "Imports should not be named with '.' ",
				Fix: func() {
					i.Name = nil
				},
			})
		}
	}

	return out
}
