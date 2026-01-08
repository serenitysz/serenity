package rules

func CanAutoFix(cfg *LinterOptions) bool {
	return cfg.Assistance != nil &&
		cfg.Assistance.Use &&
		cfg.Assistance.AutoFix != nil && *cfg.Assistance.AutoFix
}

// TODO: Change to uint16 (unsigned)
func GetMaxIssues(cfg *LinterOptions) uint16 {
	if cfg.Linter.Issues != nil {
		return cfg.Linter.Issues.Max
	}

	return 0
}
