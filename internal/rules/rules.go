package rules

import (
	"go/ast"
	"go/token"
	"go/types"
)

type Runner struct {
	File           *ast.File
	Fset           *token.FileSet
	Cfg            *LinterOptions
	Issues         *[]Issue
	Autofix        bool
	Unsafe         bool
	Modified       bool
	ShouldStop     func() bool
	MutatedObjects map[types.Object]bool
	TypesInfo      *types.Info
	IssuesCount    *uint16
}

type LinterOptions struct {
	Linter      LinterRules         `json:"linter" yaml:"linter" toml:"linter"`
	File        *GoFileOptions      `json:"go,omitempty" yaml:"go,omitempty" toml:"go,omitempty"`
	Git         *GitOptions         `json:"git,omitempty" yaml:"git,omitempty" toml:"git,omitempty"`
	Schema      string              `json:"$schema" yaml:"$schema,omitempty" toml:"$schema,omitempty"`
	Extends     *[]string           `json:"extends,omitempty" yaml:"extends,omitempty" toml:"extends,omitempty"`
	Assistance  *AssistanceOptions  `json:"assistance,omitempty" yaml:"assistance,omitempty" toml:"assistance,omitempty"`
	Performance *PerformanceOptions `json:"performance,omitempty" yaml:"performance,omitempty" toml:"performance,omitempty"`
}

func (l *LinterOptions) GetMaxIssues() uint16 {
	if l.Linter.Issues != nil {
		return l.Linter.Issues.Max
	}

	return 0
}

func (l *LinterOptions) ShouldAutofix() bool {
	return l.Assistance != nil &&
		l.Assistance.Use &&
		l.Assistance.AutoFix != nil && *l.Assistance.AutoFix
}

type Issue struct {
	Pos      token.Position
	ID       uint16
	Flags    uint8
	Severity Severity
	ArgInt1  int
	ArgInt2  int
	ArgStr1  string
}

// Issue flags

const (
	IssueExperimentalFlag uint8 = 1 << iota
	IssueFixedFlags
)

type GitOptions struct {
	Use         bool    `json:"use" yaml:"use" toml:"use"`
	Ignore      *bool   `json:"ignore,omitempty" yaml:"ignore,omitempty" toml:"ignore,omitempty"`
	ChangedOnly *bool   `json:"changedOnly,omitempty" yaml:"changedOnly,omitempty" toml:"changedOnly,omitempty"`
	StagedOnly  *bool   `json:"stagedOnly,omitempty" yaml:"stagedOnly,omitempty" toml:"stagedOnly,omitempty"`
	Branch      *string `json:"branch,omitempty" yaml:"branch,omitempty" toml:"branch,omitempty"`
	Root        *string `json:"root,omitempty" yaml:"root,omitempty" toml:"root,omitempty"`
}

type GoFileOptions struct {
	Exclude     *[]string `json:"exclude,omitempty" yaml:"exclude,omitempty" toml:"exclude,omitempty"`
	MaxFileSize *int64    `json:"maxFileSize,omitempty" yaml:"maxFileSize,omitempty" toml:"maxFileSize,omitempty"`
}

type PerformanceOptions struct {
	Use     bool  `json:"use,omitempty" yaml:"use,omitempty" toml:"use,omitempty"`
	Threads *int  `json:"threads,omitempty" yaml:"threads,omitempty" toml:"threads,omitempty"`
	Caching *bool `json:"caching,omitempty" yaml:"caching,omitempty" toml:"caching,omitempty"`
}

type AssistanceOptions struct {
	Use     bool  `json:"use,omitempty" yaml:"use,omitempty" toml:"use,omitempty"`
	AutoFix *bool `json:"autofix,omitempty" yaml:"autofix,omitempty" toml:"autofix,omitempty"`
}

type LinterRules struct {
	Use    bool                 `json:"use" yaml:"use" toml:"use"`
	Rules  LinterRulesGroup     `json:"rules"  yaml:"rules" toml:"rules"`
	Issues *LinterIssuesOptions `json:"issues,omitempty" yaml:"issues,omitempty" toml:"issues,omitempty"`
}

type LinterIssuesOptions struct {
	Use bool   `json:"use" yaml:"use" toml:"use"`
	Max uint16 `json:"max" yaml:"max" toml:"max"`
}

type LinterBaseRule struct {
	Severity string `json:"severity" yaml:"severity" toml:"severity"`
}

type AnyMaxValueBasedRule struct {
	Severity string  `json:"severity" yaml:"severity" toml:"severity"`
	Max      *uint16 `json:"max,omitempty" yaml:"max,omitempty" toml:"max,omitempty"`
}

type AnyPatternBasedRule struct {
	Severity string  `json:"severity" yaml:"severity" toml:"severity"`
	Pattern  *string `json:"pattern,omitempty" yaml:"pattern,omitempty" toml:"pattern,omitempty"`
}

type LinterRulesGroup struct {
	UseRecommended *bool                    `json:"recommended,omitempty" yaml:"recommended,omitempty" toml:"recommended,omitempty"`
	Errors         *ErrorHandlingRulesGroup `json:"errors,omitempty" yaml:"errors,omitempty" toml:"errors,omitempty"`
	Imports        *ImportRulesGroup        `json:"imports,omitempty" yaml:"imports,omitempty" toml:"imports,omitempty"`
	BestPractices  *BestPracticesRulesGroup `json:"bestPractices,omitempty" yaml:"bestPractices,omitempty" toml:"bestPractices,omitempty"`
	Correctness    *CorrectnessRulesGroup   `json:"correctness,omitempty" yaml:"correctness,omitempty" toml:"correctness,omitempty"`
	Complexity     *ComplexityRulesGroup    `json:"complexity,omitempty" yaml:"complexity,omitempty" toml:"complexity,omitempty"`
	Naming         *NamingRulesGroup        `json:"naming,omitempty" yaml:"naming,omitempty" toml:"naming,omitempty"`
	Style          *StyleRulesGroup         `json:"style,omitempty" yaml:"style,omitempty" toml:"style,omitempty"`
}

type ErrorHandlingRulesGroup struct {
	Use               bool            `json:"use" yaml:"use" toml:"use"`
	NoErrorShadowing  *LinterBaseRule `json:"noErrorShadowing,omitempty" yaml:"noErrorShadowing,omitempty" toml:"noErrorShadowing,omitempty"`
	ErrorStringFormat *LinterBaseRule `json:"errorStringFormat,omitempty" yaml:"errorStringFormat,omitempty" toml:"errorStringFormat,omitempty"`
	ErrorNotWrapped   *LinterBaseRule `json:"errorNotWrapped,omitempty" yaml:"errorNotWrapped,omitempty" toml:"errorNotWrapped,omitempty"`
}

type ImportRulesGroup struct {
	Use                  bool                    `json:"use" yaml:"use" toml:"use"`
	NoDotImports         *LinterBaseRule         `json:"noDotImports,omitempty" yaml:"noDotImports,omitempty" toml:"noDotImports,omitempty"`
	DisallowedPackages   *DisallowedPackagesRule `json:"disallowedPackages,omitempty" yaml:"disallowedPackages,omitempty" toml:"disallowedPackages,omitempty"`
	RedundantImportAlias *LinterBaseRule         `json:"redundantImportAlias,omitempty" yaml:"redundantImportAlias,omitempty" toml:"redundantImportAlias,omitempty"`
}

type BestPracticesRulesGroup struct {
	Use                    bool                  `json:"use" yaml:"use" toml:"use"`
	SimplifyBooleanReturn  *LinterBaseRule       `json:"simplifyBooleanReturn,omitempty" yaml:"simplifyBooleanReturn,omitempty" toml:"simplifyBooleanReturn,omitempty"`
	GetMustReturnValue     *LinterBaseRule       `json:"getMustReturnValue,omitempty" yaml:"getMustReturnValue,omitempty" toml:"getMustReturnValue,omitempty"`
	PreferEarlyReturn      *LinterBaseRule       `json:"preferEarlyReturn,omitempty" yaml:"preferEarlyReturn,omitempty" toml:"preferEarlyReturn,omitempty"`
	RedundantErrorCheck    *LinterBaseRule       `json:"redundantErrorCheck,omitempty" yaml:"redundantErrorCheck,omitempty" toml:"redundantErrorCheck,omitempty"`
	NoDeferInLoop          *LinterBaseRule       `json:"noDeferInLoop,omitempty" yaml:"noDeferInLoop,omitempty" toml:"noDeferInLoop,omitempty"`
	UseContextInFirstParam *LinterBaseRule       `json:"useContextInFirstParam,omitempty" yaml:"useContextInFirstParam,omitempty" toml:"useContextInFirstParam,omitempty"`
	NoBareReturns          *LinterBaseRule       `json:"noBareReturns,omitempty" yaml:"noBareReturns,omitempty" toml:"noBareReturns,omitempty"`
	NoMagicNumbers         *LinterBaseRule       `json:"noMagicNumbers,omitempty" yaml:"noMagicNumbers,omitempty" toml:"noMagicNumbers,omitempty"`
	UseSliceCapacity       *LinterBaseRule       `json:"useSliceCapacity,omitempty" yaml:"useSliceCapacity,omitempty" toml:"useSliceCapacity,omitempty"`
	MaxParams              *AnyMaxValueBasedRule `json:"maxParams,omitempty" yaml:"maxParams,omitempty" toml:"maxParams,omitempty"`
	AvoidEmptyStructs      *LinterBaseRule       `json:"avoidEmptyStructs,omitempty" yaml:"avoidEmptyStructs,omitempty" toml:"avoidEmptyStructs,omitempty"`
	AlwaysPreferConst      *LinterBaseRule       `json:"alwaysPreferConst,omitempty" yaml:"alwaysPreferConst,omitempty" toml:"alwaysPreferConst,omitempty"`
}

type CorrectnessRulesGroup struct {
	Use                    bool                  `json:"use" yaml:"use" toml:"use"`
	UnusedReceiver         *LinterBaseRule       `json:"unusedReceiver,omitempty" yaml:"unusedReceiver,omitempty" toml:"unusedReceiver,omitempty"`
	UnusedParams           *LinterBaseRule       `json:"ununsedParams,omitempty" yaml:"unusedParams,omitempty" toml:"unusedParams,omitempty"`
	EmptyBlock             *LinterBaseRule       `json:"emptyBlock,omitempty" yaml:"emptyBlock,omitempty" toml:"emptyBlock,omitempty"`
	BoolLiteralExpressions *LinterBaseRule       `json:"boolLiteralExpressions,omitempty" yaml:"boolLiteralExpressions,omitempty" toml:"boolLiteralExpressions,omitempty"`
	AmbiguousReturns       *AmbiguousReturnsRule `json:"ambiguousReturns,omitempty" yaml:"ambiguousReturns,omitempty" toml:"ambiguousReturns,omitempty"`
}

type ComplexityRulesGroup struct {
	Use                  bool                  `json:"use" yaml:"use" toml:"use"`
	MaxFuncLines         *AnyMaxValueBasedRule `json:"maxFuncLines,omitempty" yaml:"maxFuncLines,omitempty" toml:"maxFuncLines,omitempty"`
	MaxNestingDepth      *AnyMaxValueBasedRule `json:"maxNestingDepth,omitempty" yaml:"maxNestingDepth,omitempty" toml:"maxNestingDepth,omitempty"`
	CyclomaticComplexity *AnyMaxValueBasedRule `json:"cyclomaticComplexity,omitempty" yaml:"cyclomaticComplexity,omitempty" toml:"cyclomaticComplexity,omitempty"`
}

type NamingRulesGroup struct {
	Use                 bool                 `json:"use" yaml:"use" toml:"use"`
	ReceiverNames       *ReceiverNamesRule   `json:"receiverNames,omitempty" yaml:"receiverNames,omitempty" toml:"receiverNames,omitempty"`
	ExportedIdentifiers *AnyPatternBasedRule `json:"exportedIdentifiers,omitempty" yaml:"exportedIdentifiers,omitempty" toml:"exportedIdentifiers,omitempty"`
	ImportedIdentifiers *AnyPatternBasedRule `json:"importedIdentifiers,omitempty" yaml:"importedIdentifiers,omitempty" toml:"importedIdentifiers,omitempty"`
	BannedChars         *BannedCharsRule     `json:"bannedChars,omitempty" yaml:"bannedChars,omitempty" toml:"bannedChars,omitempty"`
}

type StyleRulesGroup struct {
	Use             bool                  `json:"use" yaml:"use" toml:"use"`
	PreferIncDec    *LinterBaseRule       `json:"preferIncDec,omitempty" yaml:"preferIncDec,omitempty" toml:"preferIncDec,omitempty"`
	MaxLineLength   *AnyMaxValueBasedRule `json:"maxLineLength,omitempty" yaml:"maxLineLength,omitempty" toml:"maxLineLength,omitempty"`
	PackageComments *PackageCommentsRule  `json:"packageComments,omitempty" yaml:"packageComments,omitempty" toml:"packageComments,omitempty"`
	CommentSpacing  *CommentSpacingRule   `json:"commentSpacing,omitempty" yaml:"commentSpacing,omitempty" toml:"commentSpacing,omitempty"`
	FileHeader      *FileHeaderRule       `json:"fileHeader,omitempty" yaml:"fileHeader,omitempty" toml:"fileHeader,omitempty"`
}

// SINGLE RULES STRUCTS

type FileHeaderRule struct {
	Severity     string `json:"severity" yaml:"severity" toml:"severity"`
	Header       string `json:"header" yaml:"header" toml:"header"`
	AllowShebang *bool  `json:"allowShebang,omitempty" yaml:"allowShebang,omitempty" toml:"allowShebang,omitempty"`
}

type CommentSpacingRule struct {
	Severity   string    `json:"severity" yaml:"severity" toml:"severity"`
	Exceptions *[]string `json:"exceptions,omitempty" yaml:"exceptions,omitempty" toml:"exceptions,omitempty"`
}

type PackageCommentsRule struct {
	Severity         string `json:"severity" yaml:"severity" toml:"severity"`
	RequireTopOfFile *bool  `json:"requireTopOfFile,omitempty" yaml:"requireTopOfFile,omitempty" toml:"requireTopOfFile,omitempty"`
}

type BannedCharsRule struct {
	Severity string   `json:"severity" yaml:"severity" toml:"severity"`
	Chars    []string `json:"chars" yaml:"chars" toml:"chars"`
}

type AmbiguousReturnsRule struct {
	Severity           string `json:"severity" yaml:"severity" toml:"severity"`
	MaxUnnamedSameType *int   `json:"maxUnnamedSameType,omitempty" yaml:"maxUnnamedSameType,omitempty" toml:"maxUnnamedSameType,omitempty"`
}

type DisallowedPackagesRule struct {
	Severity string   `json:"severity" yaml:"severity" toml:"severity"`
	Packages []string `json:"packages" yaml:"packages" toml:"packages"`
}

type ReceiverNamesRule struct {
	Severity string `json:"severity" yaml:"severity" toml:"severity"`
	MaxSize  *int   `json:"maxSize,omitempty" yaml:"maxSize,omitempty" toml:"maxSize,omitempty"`
}

const (
	// ERRORS

	NoErrorShadowingID uint16 = iota
	ErrorStringFormatID
	ErrorNotWrappedID

	// IMPORTS

	NoDotImportsID
	DisallowedPackagesID
	RedundantImportAliasID

	// BEST PRACTICES

	NoDeferInLoopID
	UseContextInFirstParamID
	NoBareReturnsID
	NoMagicNumbersID
	UseSliceCapacityID
	MaxParamsID
	AvoidEmptyStructsID
	AlwaysPreferConstID
	GetMustReturnValueID

	// CORRECTNESS

	UnusedReceiverID
	UnusedParamsID
	EmptyBlockID
	AmbiguousReturnID

	// COMPLEXITY

	MaxFuncLinesID
	MaxNestingDepthID
	CyclomaticComplexityID

	// NAMING

	ReceiverNameID
	ExportedIdentifiersID
	ImportedIdentifiersID

	// STYLE

	PreferIncDecID
)
