package linter

import (
	"github.com/serenitysz/serenity/internal/rules"
	"github.com/serenitysz/serenity/internal/rules/bestpractices"
	"github.com/serenitysz/serenity/internal/rules/complexity"
	"github.com/serenitysz/serenity/internal/rules/correctness"
	"github.com/serenitysz/serenity/internal/rules/errs"
	"github.com/serenitysz/serenity/internal/rules/imports"
	"github.com/serenitysz/serenity/internal/rules/naming"
	"github.com/serenitysz/serenity/internal/rules/style"
)

func BuildActiveRules(cfg *rules.LinterOptions) *ActiveRules {
	active := &ActiveRules{}

	r := cfg.Linter.Rules

	if imp := r.Imports; imp != nil && imp.Use {
		if imp.NoDotImports != nil {
			active.ImportSpec = append(active.ImportSpec, &imports.NoDotImportsRule{
				Severity: rules.ParseSeverity(imp.NoDotImports.Severity),
			})
		}

		if imp.DisallowedPackages != nil {
			packages := make(map[string]struct{}, len(imp.DisallowedPackages.Packages))
			for _, pkg := range imp.DisallowedPackages.Packages {
				packages[pkg] = struct{}{}
			}

			active.ImportSpec = append(active.ImportSpec, &imports.DisallowedPackagesRule{
				Severity: rules.ParseSeverity(imp.DisallowedPackages.Severity),
				Packages: packages,
			})
		}

		if imp.RedundantImportAlias != nil {
			active.ImportSpec = append(active.ImportSpec, &imports.RedundantImportAliasRule{
				Severity: rules.ParseSeverity(imp.RedundantImportAlias.Severity),
			})
			active.HasAutofixRules = true
		}
	}

	if errCfg := r.Errors; errCfg != nil && errCfg.Use {
		if errCfg.ErrorStringFormat != nil {
			active.ReturnStmt = append(active.ReturnStmt, &errs.ErrorStringFormatRule{
				Severity: rules.ParseSeverity(errCfg.ErrorStringFormat.Severity),
			})
			active.HasAutofixRules = true
		}

		if errCfg.ErrorNotWrapped != nil {
			active.ReturnStmt = append(active.ReturnStmt, &errs.ErrorNotWrappedRule{
				Severity: rules.ParseSeverity(errCfg.ErrorNotWrapped.Severity),
			})
			active.HasAutofixRules = true
		}
	}

	if bp := r.BestPractices; bp != nil && bp.Use {
		if bp.MaxParams != nil {
			limit := uint16(5)
			if bp.MaxParams.Max != nil {
				limit = *bp.MaxParams.Max
			}

			active.FuncDecl = append(active.FuncDecl, &bestpractices.MaxParamsRule{
				Limit:    limit,
				Severity: rules.ParseSeverity(bp.MaxParams.Severity),
			})
		}

		if bp.UseContextInFirstParam != nil {
			active.FuncDecl = append(active.FuncDecl, &bestpractices.ContextFirstRule{
				Severity: rules.ParseSeverity(bp.UseContextInFirstParam.Severity),
			})
		}

		if bp.AvoidEmptyStructs != nil {
			active.TypeSpec = append(active.TypeSpec, &bestpractices.AvoidEmptyStructsRule{
				Severity: rules.ParseSeverity(bp.AvoidEmptyStructs.Severity),
			})
		}

		if bp.NoMagicNumbers != nil {
			active.BasicLit = append(active.BasicLit, &bestpractices.NoMagicNumbersRule{
				Severity: rules.ParseSeverity(bp.NoMagicNumbers.Severity),
			})
		}

		if bp.AlwaysPreferConst != nil {
			active.ValueSpec = append(active.ValueSpec, &bestpractices.AlwaysPreferConstRule{
				Severity: rules.ParseSeverity(bp.AlwaysPreferConst.Severity),
			})
			active.NeedsConstAnalysis = true
		}

		if bp.NoDeferInLoop != nil {
			active.DeferStmt = append(active.DeferStmt, &bestpractices.NoDeferInLoopRule{
				Severity: rules.ParseSeverity(bp.NoDeferInLoop.Severity),
			})
		}

		if bp.UseSliceCapacity != nil {
			active.CallExpr = append(active.CallExpr, &bestpractices.UseSliceCapacityRule{
				Severity: rules.ParseSeverity(bp.UseSliceCapacity.Severity),
			})
		}

		if bp.NoBareReturns != nil {
			active.ReturnStmt = append(active.ReturnStmt, &bestpractices.NoBareReturnsRule{
				Severity: rules.ParseSeverity(bp.NoBareReturns.Severity),
			})
		}

		if bp.GetMustReturnValue != nil {
			active.FuncDecl = append(active.FuncDecl, &bestpractices.GetMustReturnValueRule{
				Severity: rules.ParseSeverity(bp.GetMustReturnValue.Severity),
			})
		}
	}

	if stl := r.Style; stl != nil && stl.Use {
		if stl.PreferIncDec != nil {
			active.AssignStmt = append(active.AssignStmt, &style.PreferIncDecRule{
				Severity: rules.ParseSeverity(stl.PreferIncDec.Severity),
			})
		}
	}

	if cp := r.Complexity; cp != nil && cp.Use {
		if cp.MaxFuncLines != nil {
			limit := int16(20)
			if cp.MaxFuncLines.Max != nil {
				limit = int16(*cp.MaxFuncLines.Max)
			}

			active.FuncDecl = append(active.FuncDecl, &complexity.CheckMaxFuncLinesRule{
				Limit:    limit,
				Severity: rules.ParseSeverity(cp.MaxFuncLines.Severity),
			})
		}

		if cp.MaxLineLength != nil {
			limit := 80
			if cp.MaxLineLength.Max != nil {
				limit = int(*cp.MaxLineLength.Max)
			}

			active.File = append(active.File, &complexity.CheckMaxLineLengthRule{
				Limit:    limit,
				Severity: rules.ParseSeverity(cp.MaxLineLength.Severity),
			})
		}
	}

	if crr := r.Correctness; crr != nil && crr.Use {
		if crr.EmptyBlock != nil {
			active.BlockStmt = append(active.BlockStmt, &correctness.EmptyBlockRule{
				Severity: rules.ParseSeverity(crr.EmptyBlock.Severity),
			})
		}

		if crr.AmbiguousReturns != nil {
			maxAllowed := 1
			if crr.AmbiguousReturns.MaxUnnamedSameType != nil {
				maxAllowed = *crr.AmbiguousReturns.MaxUnnamedSameType
			}

			active.FuncDecl = append(active.FuncDecl, &correctness.AmbiguousReturnRule{
				Severity:   rules.ParseSeverity(crr.AmbiguousReturns.Severity),
				MaxAllowed: maxAllowed,
			})
		}

		if crr.BoolLiteralExpressions != nil {
			active.BinaryExpr = append(active.BinaryExpr, &correctness.BooleanLiteralExpressionsRule{
				Severity: rules.ParseSeverity(crr.BoolLiteralExpressions.Severity),
			})
		}
	}

	if n := r.Naming; n != nil && n.Use {
		if n.ReceiverNames != nil {
			maxSize := 1
			if n.ReceiverNames.MaxSize != nil {
				maxSize = *n.ReceiverNames.MaxSize
			}

			active.FuncDecl = append(active.FuncDecl, &naming.ReceiverNamesRule{
				Severity: rules.ParseSeverity(n.ReceiverNames.Severity),
				MaxSize:  maxSize,
			})
		}

		if n.ImportedIdentifiers != nil {
			active.ImportSpec = append(active.ImportSpec, naming.NewImportedIdentifiersRule(n.ImportedIdentifiers))
		}

		if n.ExportedIdentifiers != nil {
			rule := naming.NewExportedIdentifiersRule(n.ExportedIdentifiers)
			active.FuncDecl = append(active.FuncDecl, rule)
			active.TypeSpec = append(active.TypeSpec, rule)
			active.ValueSpec = append(active.ValueSpec, rule)
			active.Field = append(active.Field, rule)
		}
	}

	return active
}
