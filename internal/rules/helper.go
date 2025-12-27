package rules

func CanAutoFix(cfg *LinterOptions) bool {
	if cfg.Assistance != nil &&
		cfg.Assistance.Use != nil && *cfg.Assistance.Use &&
		cfg.Assistance.AutoFix != nil {
		return *cfg.Assistance.AutoFix
	}
	return false
}

func GetMaxIssues(cfg *LinterOptions) int16 {
	if cfg.Linter.Issues != nil && cfg.Linter.Issues.Max != nil {
		return *cfg.Linter.Issues.Max
	}
	return 0
}
