package bestpractices

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

func CheckMaxParamsNode(runner *rules.Runner) {
	bestPractices := runner.Cfg.Linter.Rules.BestPractices

	if bestPractices == nil {
		return
	}

	if bestPractices.Use != nil && !*bestPractices.Use {
		return
	}

	var limit int8 = 5

	if bestPractices.MaxParams != nil &&
		bestPractices.MaxParams.Use != nil {
		limit = int8(*bestPractices.MaxParams.Max)
	}

	fn, ok := runner.Node.(*ast.FuncDecl)
	if !ok || fn.Type.Params == nil {
		return
	}

	maxIssues := rules.GetMaxIssues(runner.Cfg)
	var count int16 = 0

	for _, field := range fn.Type.Params.List {
		if maxIssues > 0 && int16(len(*runner.Issues)) >= maxIssues {
			return
		}

		if len(field.Names) == 0 {
			count++
			continue
		}
		count += int16(len(field.Names))
	}

	if limit > 0 && int16(limit) >= count {
		return
	}

	*runner.Issues = append(*runner.Issues, rules.Issue{
		ID:      rules.MaxParamsID,
		Pos:     runner.Fset.Position(fn.Pos()),
		ArgInt1: int(limit),
		ArgInt2: int(count),
	})
}
