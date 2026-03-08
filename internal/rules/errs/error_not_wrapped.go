package errs

import (
	"go/ast"
	"go/token"

	"github.com/serenitysz/serenity/internal/rules"
)

type ErrorNotWrappedRule struct {
	Severity rules.Severity
}

func (r *ErrorNotWrappedRule) Name() string {
	return "error-not-wrapped"
}

func (r *ErrorNotWrappedRule) Targets() []ast.Node {
	return []ast.Node{(*ast.ReturnStmt)(nil)}
}

func (r *ErrorNotWrappedRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	if runner.ReachedMax() {
		return
	}

	ret := node.(*ast.ReturnStmt)

	if len(ret.Results) != 1 {
		return
	}

	ident, ok := ret.Results[0].(*ast.Ident)

	if !ok || ident.Name == "_" || ident.Name == "nil" {
		return
	}

	issue := rules.Issue{
		ArgStr1:  rules.PackContext2(ident.Name, rules.CurrentFunctionName(runner)),
		ID:       rules.ErrorNotWrappedID,
		Severity: r.Severity,
	}

	if runner.ShouldAutofix() {
		importName := ensureImport(runner.File, "fmt")
		if importName == "" {
			importName = "fmt"
		}

		ret.Results[0] = &ast.CallExpr{
			Fun: selectorForImport(importName, "Errorf"),
			Args: []ast.Expr{
				&ast.BasicLit{
					Kind:  token.STRING,
					Value: `"%w"`,
				},
				ident,
			},
		}

		runner.Modified = true
		runner.ReportFixed(ident.Pos(), issue)
		return
	}

	runner.ReportFixable(ident.Pos(), issue)
}

func selectorForImport(importName, ident string) ast.Expr {
	if importName == "" {
		return ast.NewIdent(ident)
	}

	return &ast.SelectorExpr{
		X:   ast.NewIdent(importName),
		Sel: ast.NewIdent(ident),
	}
}
