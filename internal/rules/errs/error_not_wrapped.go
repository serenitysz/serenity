package errs

import (
	"go/ast"
	"go/token"

	"github.com/serenitysz/serenity/internal/rules"
)

type ErrorNotWrappedRule struct{}

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

	if max := runner.Cfg.GetMaxIssues(); max > 0 && *runner.IssuesCount >= max {
		return
	}

	cfg := runner.Cfg.Linter.Rules.Errors

	if cfg == nil || !cfg.Use || cfg.ErrorNotWrapped == nil {
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

	if runner.Cfg.ShouldAutofix() {
		ret.Results[0] = &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent("fmt"),
				Sel: ast.NewIdent("Errorf"),
			},
			Args: []ast.Expr{
				&ast.BasicLit{
					Kind:  token.STRING,
					Value: `"%w"`,
				},
				ident,
			},
		}

		runner.Modified = true
	}

	*runner.IssuesCount++

	*runner.Issues = append(*runner.Issues, rules.Issue{
		ArgStr1:  ident.Name,
		ID:       rules.ErrorNotWrappedID,
		Pos:      runner.Fset.Position(ident.Pos()),
		Severity: rules.ParseSeverity(cfg.ErrorNotWrapped.Severity),
	})
}
