package rules

import "fmt"

type RuleMetadata struct {
	ID       uint16
	Name     string
	Template string
}

var registry = map[uint16]RuleMetadata{
	// --- ERRORS ---
	NoErrorShadowingID:  {ID: NoErrorShadowingID, Name: "no-error-shadowing", Template: "declaration of %q shadows an error variable"},
	ErrorStringFormatID: {ID: ErrorStringFormatID, Name: "error-string-format", Template: "error strings should not be capitalized or end with punctuation"},
	ErrorNotWrappedID:   {ID: ErrorNotWrappedID, Name: "error-not-wrapped", Template: "error returned from external package is not wrapped"},

	// --- IMPORTS ---
	NoDotImportsID:       {ID: NoDotImportsID, Name: "no-dot-imports", Template: "imports should not be named with '.'"},
	DisallowedPackagesID: {ID: DisallowedPackagesID, Name: "disallowed-packages", Template: "usage of package %q is disallowed"},

	// --- BEST PRACTICES ---
	NoDeferInLoopID:          {ID: NoDeferInLoopID, Name: "no-defer-in-loop", Template: "defer statements should not be used inside loops"},
	UseContextInFirstParamID: {ID: UseContextInFirstParamID, Name: "context-first-param", Template: "context.Context should be the first parameter"},
	NoBareReturnsID:          {ID: NoBareReturnsID, Name: "no-bare-returns", Template: "avoid bare returns in function %q"},
	NoMagicNumbersID:         {ID: NoMagicNumbersID, Name: "no-magic-numbers", Template: "magic number %v detected, assign it to a constant"},
	UseSliceCapacityID:       {ID: UseSliceCapacityID, Name: "use-slice-capacity", Template: "slice make call should specify capacity when length is known"},
	MaxParamsID:              {ID: MaxParamsID, Name: "max-params", Template: "function exceeds the maximum parameter limit of %d (actual: %d)"},
	AvoidEmptyStructsID:      {ID: AvoidEmptyStructsID, Name: "avoid-empty-structs", Template: "empty struct declaration found"},
	AlwaysPreferConstID:      {ID: AlwaysPreferConstID, Name: "always-prefer-const", Template: "variable %q should be a constant"},

	// --- CORRECTNESS ---
	UnusedReceiverID: {ID: UnusedReceiverID, Name: "unused-receiver", Template: "method receiver %q is not used inside the function"},
	UnusedParamsID:   {ID: UnusedParamsID, Name: "unused-params", Template: "parameter %q is unused"},
	EmptyBlockID:     {ID: EmptyBlockID, Name: "empty-block", Template: "this block is empty, consider removing or adding a comment"},

	// --- COMPLEXITY ---
	MaxFuncLinesID:         {ID: MaxFuncLinesID, Name: "max-func-lines", Template: "function exceeds the maximum line limit of %d (actual: %d)"},
	MaxNestingDepthID:      {ID: MaxNestingDepthID, Name: "max-nesting-depth", Template: "function exceeds maximum nesting depth of %d"},
	CyclomaticComplexityID: {ID: CyclomaticComplexityID, Name: "cyclomatic-complexity", Template: "function has cyclomatic complexity of %d (max: %d)"},

	// --- NAMING ---
	ReceiverNameID:        {ID: ReceiverNameID, Name: "receiver-name", Template: "receiver name %q does not match the standard"},
	ExportedIdentifiersID: {ID: ExportedIdentifiersID, Name: "exported-identifiers", Template: "exported identifier %q should have a comment"},
	ImportedIdentifiersID: {ID: ImportedIdentifiersID, Name: "imported-identifiers", Template: "package alias %q differs from default name"},
}

func GetMetadata(id uint16) (RuleMetadata, bool) {
	m, ok := registry[id]
	return m, ok
}

func FormatMessage(issue Issue) string {
	meta, ok := GetMetadata(issue.ID)
	if !ok {
		return fmt.Sprintf("issue found (unknown rule id: %d)", issue.ID)
	}

	switch issue.ID {
	case MaxParamsID, MaxFuncLinesID, CyclomaticComplexityID:
		return fmt.Sprintf(meta.Template, issue.ArgInt1, issue.ArgInt2)

	case MaxNestingDepthID:
		return fmt.Sprintf(meta.Template, issue.ArgInt1)

	case NoBareReturnsID,
		AlwaysPreferConstID,
		NoErrorShadowingID,
		UnusedReceiverID,
		UnusedParamsID,
		ReceiverNameID,
		ExportedIdentifiersID,
		ImportedIdentifiersID,
		NoMagicNumbersID,
		DisallowedPackagesID:
		return fmt.Sprintf(meta.Template, issue.ArgStr1)

	default:
		return meta.Template
	}
}
