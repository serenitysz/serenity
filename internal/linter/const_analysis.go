package linter

import (
	"go/ast"
	"go/token"
)

func (l *Linter) buildConstCandidates(params AnalysisParams) map[*ast.Ident]struct{} {
	if params.rules == nil || !params.rules.NeedsConstAnalysis || len(params.pkgFiles) == 0 {
		return nil
	}

	if !hasPotentialConstCandidates(params.pkgFiles) {
		return nil
	}

	files := make(map[string]*ast.File, len(params.pkgFiles))
	for i, file := range params.pkgFiles {
		files[params.pkgPaths[i]] = file
	}

	_, _ = ast.NewPackage(params.fset, files, nil, nil)

	candidates := make(map[*ast.Object]*ast.Ident)

	for _, file := range params.pkgFiles {
		ast.Inspect(file, func(n ast.Node) bool {
			spec, ok := n.(*ast.ValueSpec)
			if !ok {
				return true
			}

			for i, name := range spec.Names {
				if i >= len(spec.Values) {
					break
				}

				if name == nil || name.Obj == nil || name.Obj.Kind != ast.Var {
					continue
				}

				if _, ok := spec.Values[i].(*ast.BasicLit); !ok {
					continue
				}

				candidates[name.Obj] = name
			}

			return true
		})
	}

	if len(candidates) == 0 {
		return nil
	}

	mutated := make(map[*ast.Object]struct{}, len(candidates))

	for _, file := range params.pkgFiles {
		ast.Inspect(file, func(n ast.Node) bool {
			switch stmt := n.(type) {
			case *ast.AssignStmt:
				for _, lhs := range stmt.Lhs {
					markMutatedObject(lhs, mutated)
				}
			case *ast.IncDecStmt:
				markMutatedObject(stmt.X, mutated)
			case *ast.UnaryExpr:
				if stmt.Op == token.AND {
					markMutatedObject(stmt.X, mutated)
				}
			}

			return true
		})
	}

	constCandidates := make(map[*ast.Ident]struct{}, len(candidates))

	for obj, ident := range candidates {
		if _, ok := mutated[obj]; ok {
			continue
		}

		constCandidates[ident] = struct{}{}
	}

	return constCandidates
}

func hasPotentialConstCandidates(files []*ast.File) bool {
	for _, file := range files {
		found := false

		ast.Inspect(file, func(n ast.Node) bool {
			spec, ok := n.(*ast.ValueSpec)
			if !ok {
				return true
			}

			limit := min(len(spec.Names), len(spec.Values))
			for i := range limit {
				if _, ok := spec.Values[i].(*ast.BasicLit); ok {
					found = true
					return false
				}
			}

			return true
		})

		if found {
			return true
		}
	}

	return false
}

func markMutatedObject(expr ast.Expr, mutated map[*ast.Object]struct{}) {
	ident, ok := expr.(*ast.Ident)
	if !ok || ident.Obj == nil {
		return
	}

	mutated[ident.Obj] = struct{}{}
}
