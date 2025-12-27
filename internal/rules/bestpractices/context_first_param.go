package bestpractices

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

type ContextFirstRule struct{}

func (c *ContextFirstRule) Name() string {
	return "context-first-param"
}

func (c *ContextFirstRule) Targets() []ast.Node {
	return []ast.Node{(*ast.FuncDecl)(nil)}
}

func (c *ContextFirstRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	fn := node.(*ast.FuncDecl)

	if fn.Type.Params == nil || len(fn.Type.Params.List) < 2 {
		return
	}

	bp := runner.Cfg.Linter.Rules.BestPractices
	if bp == nil || (bp.Use != nil && !*bp.Use) || bp.UseContextInFirstParam == nil {
		return
	}

	cf := bp.UseContextInFirstParam
	params := fn.Type.Params.List
	maxIssues := rules.GetMaxIssues(runner.Cfg)

	for i := 1; i < len(params); i++ {
		p := params[i]

		if isContextType(p.Type) {
			if maxIssues > 0 && int16(len(*runner.Issues)) >= maxIssues {
				break
			}

			*runner.Issues = append(*runner.Issues, rules.Issue{
				ID:       rules.UseContextInFirstParamID,
				Pos:      runner.Fset.Position(p.Pos()),
				Severity: rules.ParseSeverity(cf.Severity),
			})
		}
	}
}

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
