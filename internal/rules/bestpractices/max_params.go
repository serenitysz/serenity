package bestpractices

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

func CheckMaxParamsNode(runner *rules.Runner) []rules.Issue {
	bestPractices := runner.Cfg.Linter.Rules.BestPractices

	if bestPractices == nil {
		return nil
	}

	if bestPractices.Use != nil && !*bestPractices.Use {
		return nil
	}

	var limit int8 = 5

	if bestPractices.MaxParams != nil &&
		bestPractices.MaxParams.Quantity != nil {
		limit = *bestPractices.MaxParams.Quantity
	}

	fn, ok := runner.Node.(*ast.FuncDecl)
	if !ok || fn.Type.Params == nil {
		return nil
	}

	var count int16 = 0
	for _, field := range fn.Type.Params.List {
		if len(field.Names) == 0 {
			count++
			continue
		}
		count += int16(len(field.Names))
	}

	if limit > 0 && int16(limit) >= count {
		return nil
	}

	return []rules.Issue{{
		ID:  rules.MaxParamsID,
		Pos: runner.Fset.Position(fn.Pos()),
	}}
}
