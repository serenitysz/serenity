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
	// TODO: Change to uint16 (unsigned)
	IssuesCount *uint16
}

type LinterOptions struct {
	Linter      LinterRules         `json:"linter"`
	Schema      string              `json:"$schema" yaml:"$schema,omitempty" toml:"$schema,omitempty"`
	File        *GoFileOptions      `json:"go,omitempty"`
	Extends     *[]string           `json:"extends,omitempty"`
	Assistance  *AssistanceOptions  `json:"assistance,omitempty"`
	Performance *PerformanceOptions `json:"performance,omitempty"`
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

type GoFileOptions struct {
	Exclude     *[]string `json:"exclude,omitempty"`
	MaxFileSize *int64    `json:"maxFileSize,omitempty"`
}

type PerformanceOptions struct {
	Use     *bool `json:"use,omitempty"`
	Threads *int  `json:"threads,omitempty"`
	Caching *bool `json:"caching,omitempty"`
}

type AssistanceOptions struct {
	Use     *bool `json:"use,omitempty"`
	AutoFix *bool `json:"autofix,omitempty"`
}

type LinterRules struct {
	Use    bool                 `json:"use,omitempty"`
	Rules  LinterRulesGroup     `json:"rules,omitempty"`
	Issues *LinterIssuesOptions `json:"issues,omitempty"`
}

type LinterIssuesOptions struct {
	Use bool   `json:"use,omitempty"`
	Max uint16 `json:"max,omitempty"`
}

type LinterBaseRule struct {
	Severity string `json:"severity"`
}

type AnyMaxValueBasedRule struct {
	Severity string `json:"severity"`
	Max      *int   `json:"max,omitempty"`
}

type AnyPatternBasedRule struct {
	Severity string  `json:"severity"`
	Pattern  *string `json:"pattern,omitempty"`
}

type LinterRulesGroup struct {
	UseRecommended *bool                    `json:"recommended,omitempty"`
	Errors         *ErrorHandlingRulesGroup `json:"errors,omitempty"`
	Imports        *ImportRulesGroup        `json:"imports,omitempty"`
	BestPractices  *BestPracticesRulesGroup `json:"bestPractices,omitempty"`
	Correctness    *CorrectnessRulesGroup   `json:"correctness,omitempty"`
	Complexity     *ComplexityRulesGroup    `json:"complexity,omitempty"`
	Naming         *NamingRulesGroup        `json:"naming,omitempty"`
}

type ErrorHandlingRulesGroup struct {
	Use               bool            `json:"use"`
	NoErrorShadowing  *LinterBaseRule `json:"noErrorShadowing,omitempty"`
	ErrorStringFormat *LinterBaseRule `json:"errorStringFormat,omitempty"`
	ErrorNotWrapped   *LinterBaseRule `json:"errorNotWrapped,omitempty"`
}

type ImportRulesGroup struct {
	Use                  bool                    `json:"use"`
	NoDotImports         *LinterBaseRule         `json:"noDotImports,omitempty"`
	DisallowedPackages   *DisallowedPackagesRule `json:"disallowedPackages,omitempty"`
	RedundantImportAlias *LinterBaseRule         `json:"redundantImportAlias,omitempty"`
}

type BestPracticesRulesGroup struct {
	Use                    bool                 `json:"use"`
	SimplifyBooleanReturn  *LinterBaseRule      `json:"simplifyBooleanReturn,omitempty"`
	GetMustReturnValue     *LinterBaseRule      `json:"getMustReturnValue,omitempty"`
	PreferEarlyReturn      *LinterBaseRule      `json:"preferEarlyReturn,omitempty"`
	RedundantErrorCheck    *LinterBaseRule      `json:"redundantErrorCheck,omitempty"`
	NoDeferInLoop          *LinterBaseRule      `json:"noDeferInLoop,omitempty"`
	UseContextInFirstParam *LinterBaseRule      `json:"useContextInFirstParam,omitempty"`
	NoBareReturns          *LinterBaseRule      `json:"noBareReturns,omitempty"`
	NoMagicNumbers         *LinterBaseRule      `json:"noMagicNumbers,omitempty"`
	UseSliceCapacity       *LinterBaseRule      `json:"useSliceCapacity,omitempty"`
	MaxParams              *LinterIssuesOptions `json:"maxParams,omitempty"` // NOTE: aqui ficou MaxParams.Max, talvez melhor ser Quantity?
	AvoidEmptyStructs      *LinterBaseRule      `json:"avoidEmptyStructs,omitempty"`
	AlwaysPreferConst      *LinterBaseRule      `json:"alwaysPreferConst,omitempty"`
}

type CorrectnessRulesGroup struct {
	Use                    bool                  `json:"use,omitempty"`
	UnusedReceiver         *LinterBaseRule       `json:"unusedReceiver,omitempty"`
	UnusedParams           *LinterBaseRule       `json:"ununsedParams,omitempty"`
	EmptyBlock             *LinterBaseRule       `json:"emptyBlock,omitempty"`
	BoolLiteralExpressions *LinterBaseRule       `json:"boolLiteralExpressions,omitempty"`
	AmbiguousReturns       *AmbiguousReturnsRule `json:"ambiguousReturns,omitempty"`
}

type ComplexityRulesGroup struct {
	Use                  bool                  `json:"use,omitempty"`
	MaxFuncLines         *AnyMaxValueBasedRule `json:"maxFuncLines,omitempty"`
	MaxNestingDepth      *AnyMaxValueBasedRule `json:"maxNestingDepth,omitempty"`
	CyclomaticComplexity *AnyMaxValueBasedRule `json:"cyclomaticComplexity,omitempty"`
}

type NamingRulesGroup struct {
	Use                 bool                 `json:"use,omitempty"`
	ReceiverNames       *ReceiverNamesRule   `json:"receiverNames,omitempty"`
	ExportedIdentifiers *AnyPatternBasedRule `json:"exportedIdentifiers,omitempty"`
	ImportedIdentifiers *AnyPatternBasedRule `json:"importedIdentifiers,omitempty"`
	BannedChars         *BannedCharsRule     `json:"bannedChars,omitempty"`
}

type StyleRulesGroup struct {
	Use             bool                  `json:"use,omitempty"`
	PreferIncDec    *LinterBaseRule       `json:"preferIncDec,omitempty"`
	MaxLineLength   *AnyMaxValueBasedRule `json:"maxLineLength,omitempty"`
	PackageComments *PackageCommentsRule  `json:"packageComments,omitempty"`
	CommentSpacing  *CommentSpacingRule   `json:"commentSpacing,omitempty"`
	FileHeader      *FileHeaderRule       `json:"fileHeader,omitempty"`
}

// SINGLE RULES STRUCTS

type FileHeaderRule struct {
	Severity     string `json:"severity,omitempty"`
	Header       string `json:"header"`
	AllowShebang *bool  `json:"allowShebang,omitempty"`
}

type CommentSpacingRule struct {
	Severity   string    `json:"severity"`
	Exceptions *[]string `json:"exceptions,omitempty"`
}

type PackageCommentsRule struct {
	Severity         string `json:"severity"`
	RequireTopOfFile *bool  `json:"requireTopOfFile,omitempty"`
}

type BannedCharsRule struct {
	Severity string   `json:"severity"`
	Chars    []string `json:"chars"`
}

type AmbiguousReturnsRule struct {
	Severity           string `json:"severity"`
	MaxUnnamedSameType *uint8 `json:"maxUnnamedSameType,omitempty"`
}

type DisallowedPackagesRule struct {
	Severity string   `json:"severity"`
	Packages []string `json:"packages"`
}

type ReceiverNamesRule struct {
	Severity string `json:"severity"`
	MaxSize  *int   `json:"maxSize,omitempty"`
}

const (
	// ERRORS

	NoErrorShadowingID uint16 = iota
	ErrorStringFormatID
	ErrorNotWrappedID

	// IMPORTS

	NoDotImportsID
	DisallowedPackagesID

	// BEST PRACTICES

	NoDeferInLoopID
	UseContextInFirstParamID
	NoBareReturnsID
	NoMagicNumbersID
	UseSliceCapacityID
	MaxParamsID
	AvoidEmptyStructsID
	AlwaysPreferConstID

	// CORRECTNESS

	UnusedReceiverID
	UnusedParamsID
	EmptyBlockID

	// COMPLEXITY

	MaxFuncLinesID
	MaxNestingDepthID
	CyclomaticComplexityID

	// NAMING

	ReceiverNameID
	ExportedIdentifiersID
	ImportedIdentifiersID
)
