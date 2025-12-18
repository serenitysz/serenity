package rules

import (
	"go/ast"
	"go/token"
)

func CheckNoDotImports(
	f *ast.File,
	fset *token.FileSet,
	out []Issue,
) []Issue {
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
