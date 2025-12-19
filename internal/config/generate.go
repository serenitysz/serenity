package config

import "github.com/serenitysz/serenity/internal/rules"

func GenDefaultConfig() *rules.LinterOptions {
	var OneMBInBytes int64 = 1048576
	var maxParams int8 = 5

	config := rules.LinterOptions{
		File: &rules.GoFileOptions{
			MaxFileSize: &OneMBInBytes, Exclude: &[]string{"**/vendor/**", "**/*.test.go"},
		}, Schema: "https://raw.githubusercontent.com/serenitysz/schema/main/versions/1.0.0.json", Linter: rules.LinterRules{
			Use: Bool(true), Rules: &rules.LinterRulesGroup{
				UseRecommended: Bool(true),
				Imports: &rules.ImportRulesGroup{
					NoDotImports: &rules.LinterBaseRule{Severity: "error"},
				},
				BestPractices: &rules.BestPracticesRulesGroup{
					UseContextInFirstParam: &rules.LinterBaseRule{Severity: "warn"},
					MaxParams: &rules.MaxParams{
						Quantity: &maxParams,
					},
				},
			},
			Issues: &rules.LinterIssuesOptions{
				Max: Int16(15),
				Use: Bool(true),
			},
		},
		Performance: &rules.PerformanceOptions{
			Use:     Bool(true),
			Caching: Bool(true),
		},
		Assistance: &rules.AssistanceOptions{
			Use:     Bool(true),
			AutoFix: Bool(false),
		},
	}

	return &config
}

func Bool(v bool) *bool    { return &v }
func Int16(v int16) *int16 { return &v }
