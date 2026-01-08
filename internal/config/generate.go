package config

import (
	"github.com/serenitysz/serenity/internal/rules"
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
				UseRecommended: Ptr(true),
			},
			Issues: &rules.LinterIssuesOptions{
				Use: true,
				Max: uint16(20),
			},
		},
		Performance: &rules.PerformanceOptions{
			Use:     true,
			Caching: Ptr(true),
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
				UseRecommended: Ptr(false),
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
						Pattern:  Ptr("^[A-Z][A-Za-z0-9]*$"),
					},
					ImportedIdentifiers: &rules.AnyPatternBasedRule{
						Severity: "error",
						Pattern:  Ptr("^[a-z][a-z0-9]*$"),
					},
				},
			},
		},
		Performance: &rules.PerformanceOptions{
			Threads: nil,
			Use:     true,
			Caching: Ptr(true),
		},
		Assistance: &rules.AssistanceOptions{
			AutoFix: autofix,
			Use:     true,
		},
	}
}

func Ptr[T any](value T) *T {
	return &value
}
