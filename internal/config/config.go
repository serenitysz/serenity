package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/serenitysz/serenity/internal/rules"
)

func ReadConfig(path string) (*rules.LinterOptions, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config rules.LinterOptions

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &config, nil
}

func ApplyRecommended(cfg *rules.LinterOptions) {
	if cfg.Linter.Rules == nil || cfg.Linter.Rules.UseRecommended == nil || !*cfg.Linter.Rules.UseRecommended {
		return
	}

	assistance := cfg.Assistance
	if assistance == nil {
		use := true

		assistance = &rules.AssistanceOptions{
			Use:     &use,
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
		var q int16 = 5
		u := true
		rulesGroup.BestPractices.MaxParams = &rules.LinterIssuesOptions{Max: &q, Use: &u}
	}

	if rulesGroup.Complexity == nil {
		rulesGroup.Complexity = &rules.ComplexityRulesGroup{}
	}

	if rulesGroup.Complexity.MaxFuncLines == nil {
		rulesGroup.Complexity.MaxFuncLines = &rules.AnyMaxValueBasedRule{Severity: "warn"}
	}

	if rulesGroup.Complexity.MaxFuncLines.Max == nil {
		m := 20
		rulesGroup.Complexity.MaxFuncLines.Max = &m
	}
}
