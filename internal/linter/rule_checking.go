package linter

import (
	"reflect"

	"github.com/serenitysz/serenity/internal/rules"
	"github.com/serenitysz/serenity/internal/rules/bestpractices"
	"github.com/serenitysz/serenity/internal/rules/complexity"
	"github.com/serenitysz/serenity/internal/rules/imports"
)

func GetActiveRulesMap(cfg *rules.LinterOptions) map[reflect.Type][]rules.Rule {
	activeRules := make(map[reflect.Type][]rules.Rule)
	const initialCap = 8

	register := func(r rules.Rule) {
		for _, target := range r.Targets() {
			t := rules.GetNodeType(target)
			if activeRules[t] == nil {
				activeRules[t] = make([]rules.Rule, 0, initialCap)
			}
			activeRules[t] = append(activeRules[t], r)
		}
	}

	r := cfg.Linter.Rules

	if imp := r.Imports; imp != nil && (imp.Use == nil || *imp.Use) {
		if imp.NoDotImports != nil {
			register(&imports.NoDotImportsRule{})
		}
	}

	if bp := r.BestPractices; bp != nil && (bp.Use == nil || *bp.Use) {
		if bp.MaxParams != nil {
			register(&bestpractices.MaxParamsRule{})
		}
		if bp.UseContextInFirstParam != nil {
			register(&bestpractices.ContextFirstRule{})
		}
		if bp.AvoidEmptyStructs != nil {
			register(&bestpractices.AvoidEmptyStructsRule{})
		}
	}

	if cp := r.Complexity; cp != nil && (cp.Use == nil || *cp.Use) {
		if cp.MaxFuncLines != nil {
			register(&complexity.CheckMaxFuncLinesRule{})
		}
	}

	return activeRules
}
