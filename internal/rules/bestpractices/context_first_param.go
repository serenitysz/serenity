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

	contextFirst := bestPractices.UseContextInFirstParam
	if contextFirst == nil {
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

	maxIssues := rules.GetMaxIssues(runner.Cfg)
	issues := make([]rules.Issue, 0, len(params.List))

	for i := 1; i < len(params.List); i++ {
		p := params.List[i]

		if isContextType(p.Type) {
			if maxIssues > 0 && int16(len(issues)) >= maxIssues {
				break
			}

			issues = append(issues, rules.Issue{
				ID:       rules.UseContextInFirstParamID,
				Pos:      runner.Fset.Position(p.Pos()),
				Message:  "context.Context should be the first parameter",
				Severity: rules.ParseSeverity(contextFirst.Severity),
			})
		}
	}

	return issues
}
