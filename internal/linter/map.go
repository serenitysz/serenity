package linter

import (
	"reflect"

	"github.com/serenitysz/serenity/internal/rules"
	"github.com/serenitysz/serenity/internal/rules/bestpractices"
	"github.com/serenitysz/serenity/internal/rules/complexity"
	"github.com/serenitysz/serenity/internal/rules/correctness"
	"github.com/serenitysz/serenity/internal/rules/errs"
	"github.com/serenitysz/serenity/internal/rules/imports"
	"github.com/serenitysz/serenity/internal/rules/naming"
	"github.com/serenitysz/serenity/internal/rules/style"
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

	if imp := r.Imports; imp != nil && imp.Use {
		if imp.NoDotImports != nil {
			register(&imports.NoDotImportsRule{})
		}

		if imp.DisallowedPackages != nil {
			register(&imports.DisallowedPackagesRule{})
		}

		if imp.RedundantImportAlias != nil {
			register(&imports.RedundantImportAliasRule{})
		}
	}

	if err := r.Errors; err != nil && err.Use {
		if err.ErrorStringFormat != nil {
			register(&errs.ErrorStringFormatRule{})
		}
		if err.ErrorNotWrapped != nil {
			register(&errs.ErrorNotWrappedRule{})
		}
	}

	if bp := r.BestPractices; bp != nil && bp.Use {
		if bp.MaxParams != nil {
			register(&bestpractices.MaxParamsRule{})
		}
		if bp.UseContextInFirstParam != nil {
			register(&bestpractices.ContextFirstRule{})
		}
		if bp.AvoidEmptyStructs != nil {
			register(&bestpractices.AvoidEmptyStructsRule{})
		}

		if bp.NoMagicNumbers != nil {
			register(&bestpractices.NoMagicNumbersRule{})
		}

		if bp.AlwaysPreferConst != nil {
			register(&bestpractices.AlwaysPreferConstRule{})
		}

		if bp.NoDeferInLoop != nil {
			register(&bestpractices.NoDeferInLoopRule{})
		}

		if bp.UseSliceCapacity != nil {
			register(&bestpractices.UseSliceCapacityRule{})
		}
		if bp.NoBareReturns != nil {
			register(&bestpractices.NoBareReturnsRule{})
		}
		if bp.GetMustReturnValue != nil {
			register(&bestpractices.GetMustReturnValueRule{})
		}
	}

	if stl := r.Style; stl != nil && stl.Use {
		if stl.PreferIncDec != nil {
			register(&style.PreferIncDecRule{})
		}
	}

	if cp := r.Complexity; cp != nil && cp.Use {
		if cp.MaxFuncLines != nil {
			register(&complexity.CheckMaxFuncLinesRule{})
		}
	}

	if crr := r.Correctness; crr != nil && crr.Use {
		if crr.EmptyBlock != nil {
			register(&correctness.EmptyBlockRule{})
		}

		if crr.AmbiguousReturns != nil {
			register(&correctness.AmbiguousReturnRule{})
		}
	}

	if n := r.Naming; n != nil && n.Use {
		if n.ReceiverNames != nil {
			register(&naming.ReceiverNamesRule{})
		}

		if n.ImportedIdentifiers != nil {
			register(&naming.ImportedIdentifiersRule{})
		}
		if n.ExportedIdentifiers != nil {
			register(&naming.ExportedIdentifiersRule{})
		}
	}

	return activeRules
}
