package rules

import "go/token"

type Config struct {
	Schema        string                  `json:"$schema,omitempty"`
	Naming        *NamingRuleGroup        `json:"naming,omitempty"`
	Complexity    *ComplexityRuleGroup    `json:"complexity,omitempty"`
	BestPractices *BestPracticesRuleGroup `json:"bestPractices,omitempty"`
	ErrorHandling *ErrorHandlingRuleGroup `json:"errorHandling,omitempty"`
	Imports       *ImportsRuleGroup       `json:"imports,omitempty"`
	Exclude       []string                `json:"exclude,omitempty"`
}

type NamingRuleGroup struct {
	Enabled bool         `json:"enabled"`
	Rules   *NamingRules `json:"rules,omitempty"`
}

type ComplexityRuleGroup struct {
	Enabled bool             `json:"enabled"`
	Rules   *ComplexityRules `json:"rules,omitempty"`
}

type BestPracticesRuleGroup struct {
	Enabled bool                `json:"enabled"`
	Rules   *BestPracticesRules `json:"rules,omitempty"`
}

type ErrorHandlingRuleGroup struct {
	Enabled bool                `json:"enabled"`
	Rules   *ErrorHandlingRules `json:"rules,omitempty"`
}

type ImportsRuleGroup struct {
	Enabled bool          `json:"enabled"`
	Rules   *ImportsRules `json:"rules,omitempty"`
}

type NamingRules struct {
	ExportedIdentifiers   *PatternRule   `json:"exportedIdentifiers,omitempty"`
	UnexportedIdentifiers *PatternRule   `json:"unexportedIdentifiers,omitempty"`
	ReceiverNames         *MaxLengthRule `json:"receiverNames,omitempty"`
}

type ComplexityRules struct {
	CyclomaticComplexity *ThresholdRule `json:"cyclomaticComplexity,omitempty"`
	MaxFunctionLines     *ThresholdRule `json:"maxFunctionLines,omitempty"`
	MaxNestingDepth      *ThresholdRule `json:"maxNestingDepth,omitempty"`
}

type BestPracticesRules struct {
	NoBareReturns     *DescriptionRule `json:"noBareReturns,omitempty"`
	ContextFirstParam *DescriptionRule `json:"contextFirstParam,omitempty"`
	NoDeferInLoop     *DescriptionRule `json:"noDeferInLoop,omitempty"`
}

type ErrorHandlingRules struct {
	ErrorWrapping     *ErrorWrappingRule     `json:"errorWrapping,omitempty"`
	ErrorStringFormat *ErrorStringFormatRule `json:"errorStringFormat,omitempty"`
	NoErrorShadowing  *DescriptionRule       `json:"noErrorShadowing,omitempty"`
}

type ImportsRules struct {
	NoDotImports       *SeverityRule           `json:"noDotImports,omitempty"`
	DisallowedPackages *DisallowedPackagesRule `json:"disallowedPackages,omitempty"`
}

type PatternRule struct {
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Pattern     string `json:"pattern"`
}

type MaxLengthRule struct {
	Severity    string `json:"severity"`
	Description string `json:"description"`
	MaxLength   int    `json:"maxLength"`
}

type ThresholdRule struct {
	Severity  string `json:"severity"`
	Threshold int    `json:"threshold"`
}

type DescriptionRule struct {
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

type ErrorWrappingRule struct {
	Severity    string `json:"severity"`
	Description string `json:"description"`
	RequireFmtW bool   `json:"requireFmtW"`
}

type ErrorStringFormatRule struct {
	Severity      string `json:"severity"`
	Case          string `json:"case"`
	NoPunctuation bool   `json:"noPunctuation"`
}

type SeverityRule struct {
	Severity string `json:"severity"`
}

type DisallowedPackagesRule struct {
	Severity string   `json:"severity"`
	Packages []string `json:"packages"`
}

type Issue struct {
	Pos     token.Position
	Message string
}

