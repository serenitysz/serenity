package complexity

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

func CheckMaxFuncLinesNode(runner *rules.Runner) []rules.Issue {
	complexity := runner.Cfg.Linter.Rules.Complexity

	if complexity != nil && complexity.Use != nil && !*complexity.Use {
		return nil
	}

	fn, ok := runner.Node.(*ast.FuncDecl)
	if !ok {
		return nil
	}

	// salve := fn.Body.List
	return nil
}
