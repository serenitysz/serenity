package config

import "github.com/serenitysz/serenity/internal/rules"

func GenDefaultConfig() *rules.LinterOptions {
	var OneMBInBytes int64 = 1048576

	config := rules.LinterOptions{
		File: &rules.GoFileOptions{
			MaxFileSize: &OneMBInBytes,
			Exclude:     &[]string{"**/vendor/**", "**/*.test.go"},
		},
		Schema: "https://raw.githubusercontent.com/serenitysz/schema/main/versions/1.0.0.json",
		Linter: rules.LinterRules{
			Use: Bool(true),
			Rules: &rules.LinterRulesGroup{
				UseRecommended: Bool(true),
			},
			Issues: &rules.LinterIssuesOptions{
				Max: Int8(15),
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

func Bool(v bool) *bool { return &v }
func Int8(v int8) *int8 { return &v }
