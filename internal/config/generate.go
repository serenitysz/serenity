package config

import "github.com/almeidazs/gowther/internal/rules"

func GenDefaultConfig() *rules.Config {
	config := rules.Config{
		Schema: "https://json-schema.org/draft/2020-12/schema",
		Naming: &rules.NamingRuleGroup{
			Enabled: true,
			Rules: &rules.NamingRules{
				ExportedIdentifiers: &rules.PatternRule{
					Severity:    "warn",
					Description: "exported identifiers must use CamelCase",
					Pattern:     "^[A-Z][a-zA-Z0-9]*$",
				},
				UnexportedIdentifiers: &rules.PatternRule{
					Severity:    "warn",
					Description: "unexported identifiers must use camelCase",
					Pattern:     "^[a-z][a-zA-Z0-9]*$",
				},
				ReceiverNames: &rules.MaxLengthRule{
					Severity:    "warn",
					Description: "method receivers should be short and consistent",
					MaxLength:   2,
				},
			},
		},
		Complexity: &rules.ComplexityRuleGroup{
			Enabled: true,
			Rules: &rules.ComplexityRules{
				CyclomaticComplexity: &rules.ThresholdRule{
					Severity:  "error",
					Threshold: 15,
				},
				MaxFunctionLines: &rules.ThresholdRule{
					Severity:  "warn",
					Threshold: 80,
				},
				MaxNestingDepth: &rules.ThresholdRule{
					Severity:  "warn",
					Threshold: 4,
				},
			},
		},
		BestPractices: &rules.BestPracticesRuleGroup{
			Enabled: true,
			Rules: &rules.BestPracticesRules{
				NoBareReturns: &rules.DescriptionRule{
					Severity:    "info",
					Description: "avoid bare returns for better readability",
				},
				ContextFirstParam: &rules.DescriptionRule{
					Severity:    "error",
					Description: "context.Context must be the first parameter",
				},
				NoDeferInLoop: &rules.DescriptionRule{
					Severity:    "warn",
					Description: "avoid defer statements inside loops",
				},
			},
		},
		ErrorHandling: &rules.ErrorHandlingRuleGroup{
			Enabled: true,
			Rules: &rules.ErrorHandlingRules{
				ErrorWrapping: &rules.ErrorWrappingRule{
					Severity:    "error",
					Description: "errors must be wrapped to preserve context",
					RequireFmtW: true,
				},
				ErrorStringFormat: &rules.ErrorStringFormatRule{
					Severity:      "warn",
					Case:          "lower",
					NoPunctuation: true,
				},
				NoErrorShadowing: &rules.DescriptionRule{
					Severity:    "error",
					Description: "do not shadow error variables",
				},
			},
		},
		Imports: &rules.ImportsRuleGroup{
			Enabled: true,
			Rules: &rules.ImportsRules{
				NoDotImports: &rules.SeverityRule{
					Severity: "warn",
				},
				DisallowedPackages: &rules.DisallowedPackagesRule{
					Severity: "error",
					Packages: []string{"io/ioutil"},
				},
			},
		},
		Exclude: []string{
			"**/vendor/**",
			"**/*.test.go",
		},
	}

	return &config
}
