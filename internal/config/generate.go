package config

import (
	"github.com/serenitysz/serenity/internal/rules"
	"github.com/serenitysz/serenity/internal/utils"
	"github.com/serenitysz/serenity/internal/version"
)

func GenDefaultConfig(autofix *bool) *rules.LinterOptions {
	var OneMBInBytes int64 = 1 * 1024 * 1024

	config := rules.LinterOptions{
		File: &rules.GoFileOptions{
			MaxFileSize: &OneMBInBytes,
			Exclude:     &[]string{"**/vendor/**", "**/*.test.go"},
		},
		Schema: "https://raw.githubusercontent.com/serenitysz/schema/main/versions/" + version.Version + ".json",
		Linter: rules.LinterRules{
			Use: true,
			Rules: rules.LinterRulesGroup{
				UseRecommended: utils.Ptr(true),
			},
			Issues: &rules.LinterIssuesOptions{
				Use: true,
				Max: uint16(20),
			},
		},
		Performance: &rules.PerformanceOptions{
			Use:     true,
			Caching: utils.Ptr(true),
		},
		Assistance: &rules.AssistanceOptions{
			AutoFix: autofix,
			Use:     true,
		},
	}

	return &config
}

func GenStrictDefaultConfig(autofix *bool) *rules.LinterOptions {
	var (
		oneMB           int64  = 1 * 1024 * 1024
		maxParams       uint16 = 4
		maxFuncLines    uint16 = 40
		maxNesting      uint16 = 3
		maxCyclomatic   uint16 = 8
		receiverMaxSize        = 1
	)

	return &rules.LinterOptions{
		Schema: "https://raw.githubusercontent.com/serenitysz/schema/main/versions/" + version.Version + ".json",
		File: &rules.GoFileOptions{
			MaxFileSize: &oneMB,
			Exclude:     &[]string{"**/vendor/**", "**/*.test.go"},
		},
		Linter: rules.LinterRules{
			Use: true,
			Issues: &rules.LinterIssuesOptions{
				Use: true,
				Max: uint16(10),
			},
			Rules: rules.LinterRulesGroup{
				UseRecommended: utils.Ptr(false),
				Errors: &rules.ErrorHandlingRulesGroup{
					Use: true,
					NoErrorShadowing: &rules.LinterBaseRule{
						Severity: "error",
					},
					ErrorStringFormat: &rules.LinterBaseRule{
						Severity: "error",
					},
					ErrorNotWrapped: &rules.LinterBaseRule{
						Severity: "error",
					},
				},
				Imports: &rules.ImportRulesGroup{
					Use: true,
					NoDotImports: &rules.LinterBaseRule{
						Severity: "error",
					},
					DisallowedPackages: &rules.DisallowedPackagesRule{
						Severity: "error",
						Packages: []string{
							"log",
						},
					},
				},
				BestPractices: &rules.BestPracticesRulesGroup{
					Use: true,
					NoDeferInLoop: &rules.LinterBaseRule{
						Severity: "error",
					},
					UseContextInFirstParam: &rules.LinterBaseRule{
						Severity: "error",
					},
					NoBareReturns: &rules.LinterBaseRule{
						Severity: "error",
					},
					NoMagicNumbers: &rules.LinterBaseRule{
						Severity: "error",
					},
					UseSliceCapacity: &rules.LinterBaseRule{
						Severity: "error",
					},
					AvoidEmptyStructs: &rules.LinterBaseRule{
						Severity: "error",
					},
					AlwaysPreferConst: &rules.LinterBaseRule{
						Severity: "error",
					},
					MaxParams: &rules.AnyMaxValueBasedRule{
						Max:      &maxParams,
						Severity: "error",
					},
				},
				Correctness: &rules.CorrectnessRulesGroup{
					Use: true,

					UnusedReceiver: &rules.LinterBaseRule{
						Severity: "error",
					},
					UnusedParams: &rules.LinterBaseRule{
						Severity: "error",
					},
					EmptyBlock: &rules.LinterBaseRule{
						Severity: "error",
					},
				},
				Complexity: &rules.ComplexityRulesGroup{
					Use: true,
					MaxFuncLines: &rules.AnyMaxValueBasedRule{
						Severity: "error",
						Max:      &maxFuncLines,
					},
					MaxNestingDepth: &rules.AnyMaxValueBasedRule{
						Severity: "error",
						Max:      &maxNesting,
					},
					CyclomaticComplexity: &rules.AnyMaxValueBasedRule{
						Severity: "error",
						Max:      &maxCyclomatic,
					},
				},
				Naming: &rules.NamingRulesGroup{
					Use: true,
					ReceiverNames: &rules.ReceiverNamesRule{
						Severity: "error",
						MaxSize:  &receiverMaxSize,
					},
					ExportedIdentifiers: &rules.AnyPatternBasedRule{
						Severity: "error",
						Pattern:  utils.Ptr("^[A-Z][A-Za-z0-9]*$"),
					},
					ImportedIdentifiers: &rules.AnyPatternBasedRule{
						Severity: "error",
						Pattern:  utils.Ptr("^[a-z][a-z0-9]*$"),
					},
				},
			},
		},
		Performance: &rules.PerformanceOptions{
			Threads: nil,
			Use:     true,
			Caching: utils.Ptr(true),
		},
		Assistance: &rules.AssistanceOptions{
			AutoFix: autofix,
			Use:     true,
		},
	}
}

func ApplyRecommended(cfg *rules.LinterOptions) {
	if cfg.Linter.Rules.UseRecommended == nil || !*cfg.Linter.Rules.UseRecommended {
		return
	}

	assistance := cfg.Assistance

	if assistance == nil {
		use := true

		assistance = &rules.AssistanceOptions{
			Use:     use,
			AutoFix: &use,
		}
	}

	rulesGroup := cfg.Linter.Rules

	if rulesGroup.Imports == nil {
		rulesGroup.Imports = &rules.ImportRulesGroup{}
	}
	if rulesGroup.Imports.NoDotImports == nil {
		rulesGroup.Imports.NoDotImports = &rules.LinterBaseRule{Severity: "error"}
	}

	if rulesGroup.BestPractices == nil {
		rulesGroup.BestPractices = &rules.BestPracticesRulesGroup{}
	}

	if rulesGroup.BestPractices.UseContextInFirstParam == nil {
		rulesGroup.BestPractices.UseContextInFirstParam = &rules.LinterBaseRule{Severity: "warn"}
	}
	if rulesGroup.BestPractices.MaxParams == nil {
		var m uint16 = 5

		rulesGroup.BestPractices.MaxParams = &rules.AnyMaxValueBasedRule{Max: &m, Severity: "warn"}
	}
	if rulesGroup.BestPractices.AvoidEmptyStructs == nil {
		rulesGroup.BestPractices.AvoidEmptyStructs = &rules.LinterBaseRule{Severity: "warn"}
	}

	if rulesGroup.BestPractices.NoMagicNumbers == nil {
		rulesGroup.BestPractices.NoMagicNumbers = &rules.LinterBaseRule{Severity: "warn"}
	}

	if rulesGroup.BestPractices.AlwaysPreferConst == nil {
		rulesGroup.BestPractices.AlwaysPreferConst = &rules.LinterBaseRule{Severity: "warn"}
	}

	if rulesGroup.BestPractices.NoDeferInLoop == nil {
		rulesGroup.BestPractices.NoDeferInLoop = &rules.LinterBaseRule{Severity: "error"}
	}

	if rulesGroup.BestPractices.UseSliceCapacity == nil {
		rulesGroup.BestPractices.UseSliceCapacity = &rules.LinterBaseRule{Severity: "warn"}
	}

	if rulesGroup.BestPractices.NoBareReturns == nil {
		rulesGroup.BestPractices.NoBareReturns = &rules.LinterBaseRule{Severity: "error"}
	}

	if rulesGroup.Complexity == nil {
		rulesGroup.Complexity = &rules.ComplexityRulesGroup{}
	}

	if rulesGroup.Complexity.MaxFuncLines == nil {
		rulesGroup.Complexity.MaxFuncLines = &rules.AnyMaxValueBasedRule{Severity: "warn"}
	}

	if rulesGroup.Complexity.MaxFuncLines.Max == nil {
		var m uint16 = 20

		rulesGroup.Complexity.MaxFuncLines.Max = &m
	}
}
