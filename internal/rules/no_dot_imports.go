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

	if err := VerifyIssues(cfg, out); err != nil {
		return out
	}

	if cfg.Linter.Rules.Imports == nil {
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
