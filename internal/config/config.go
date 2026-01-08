package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/pelletier/go-toml/v2"
	"github.com/serenitysz/serenity/internal/rules"
)

func ReadConfig(path string) (*rules.LinterOptions, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var cfg rules.LinterOptions

	ext := strings.ToLower(filepath.Ext(path))

	if ext == "" {
		return nil, fmt.Errorf("config file has no extension: %s", path)
	}

	if err := unmarshalByExt(ext, data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config %q: %w", path, err)
	}

	return &cfg, nil
}

func unmarshalByExt(ext string, data []byte, out any) error {
	switch ext {
	case ".json":
		return json.Unmarshal(data, out)

	case ".toml":
		return toml.Unmarshal(data, out)

	case ".yml", ".yaml":
		return yaml.Unmarshal(data, out)

	default:
		return fmt.Errorf(
			"unsupported config format %q (supported: JSON, TOML, YAML, YML)",
			ext,
		)
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
