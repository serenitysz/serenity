package linter

import (
	"github.com/serenitysz/serenity/internal/rules"
	"github.com/serenitysz/serenity/internal/rules/bestpractices"
	"github.com/serenitysz/serenity/internal/rules/complexity"
)

func GetActiveNodeRules(cfg *rules.LinterOptions) []func(*rules.Runner) {
	activeRules := make([]func(*rules.Runner), 0, 10)
	r := cfg.Linter.Rules

	if bp := r.BestPractices; bp != nil && (bp.Use == nil || *bp.Use) {
		if bp.MaxParams != nil {
			activeRules = append(activeRules, bestpractices.CheckMaxParamsNode)
		}
		if bp.UseContextInFirstParam != nil {
			activeRules = append(activeRules, bestpractices.CheckContextFirstParamNode)
		}
	}

	if cp := r.Complexity; cp != nil && (cp.Use == nil || *cp.Use) {
		if cp.MaxFuncLines != nil {
			activeRules = append(activeRules, complexity.CheckMaxFuncLinesNode)
		}
	}

	return activeRules
}

