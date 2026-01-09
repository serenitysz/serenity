package bestpractices

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

type NoBareReturnsRule struct{}

func (n *NoBareReturnsRule) Name() string {
	return "no-bar-returns"
}

func (n *NoBareReturnsRule) Targets() []ast.Node {
	return []ast.Node{
		(*ast.FuncLit)(nil),
		(*ast.FuncDecl)(nil),
	}
}

func (n *NoBareReturnsRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	if max := runner.Cfg.GetMaxIssues(); max > 0 && *runner.IssuesCount >= max {
		return
	}

	bp := runner.Cfg.Linter.Rules.BestPractices

	if bp == nil || !bp.Use || bp.NoBareReturns == nil {
		return
	}

	var funcType *ast.FuncType
	var body *ast.BlockStmt
	var funcName string

	switch t := node.(type) {
	case *ast.FuncDecl:
		funcType = t.Type
		body = t.Body
		funcName = t.Name.Name

	case *ast.FuncLit:
		funcType = t.Type
		body = t.Body
		funcName = "anonymous"
	default:
		return
	}

	if funcType.Results == nil || len(funcType.Results.List) == 0 {
		return
	}

	if len(funcType.Results.List[0].Names) == 0 {
		return
	}

	severity := rules.ParseSeverity(bp.NoBareReturns.Severity)

	ast.Inspect(body, func(n ast.Node) bool {
		switch t := n.(type) {
		case *ast.FuncLit:
			return false
		case *ast.ReturnStmt:
			if len(t.Results) > 0 {
				return false
			}

			*runner.IssuesCount++

			*runner.Issues = append(*runner.Issues, rules.Issue{
				ID:       rules.NoBareReturnsID,
				Pos:      runner.Fset.Position(t.Pos()),
				Severity: severity,
				ArgStr1:  funcName,
			})

			return false
		}

		return true
	})
}
