package rules

import "fmt"

type RuleMetadata struct {
	ID       uint16
	Name     string
	Template string
	Fixable  bool
}

var registry = map[uint16]RuleMetadata{
	// --- ERRORS ---
	NoErrorShadowingID:  {ID: NoErrorShadowingID, Name: "no-error-shadowing", Template: "identifier %q shadows an existing error variable"},
	ErrorStringFormatID: {ID: ErrorStringFormatID, Name: "error-string-format", Template: "error message should start with a lowercase letter and should not end with punctuation", Fixable: true},
	ErrorNotWrappedID:   {ID: ErrorNotWrappedID, Name: "error-not-wrapped", Template: "error should be wrapped before it is returned", Fixable: true},

	// --- IMPORTS ---
	NoDotImportsID:         {ID: NoDotImportsID, Name: "no-dot-imports", Template: "dot import is not allowed"},
	DisallowedPackagesID:   {ID: DisallowedPackagesID, Name: "disallowed-packages", Template: "package %q is disallowed by configuration"},
	RedundantImportAliasID: {ID: RedundantImportAliasID, Name: "redundant-import-alias", Template: "import alias is redundant", Fixable: true},

	// --- BEST PRACTICES ---
	NoDeferInLoopID:          {ID: NoDeferInLoopID, Name: "no-defer-in-loop", Template: "avoid defer inside loops"},
	UseContextInFirstParamID: {ID: UseContextInFirstParamID, Name: "context-first-param", Template: "context.Context should be the first parameter"},
	NoBareReturnsID:          {ID: NoBareReturnsID, Name: "no-bare-returns", Template: "avoid bare returns"},
	NoMagicNumbersID:         {ID: NoMagicNumbersID, Name: "no-magic-numbers", Template: "extract magic number into a named constant"},
	UseSliceCapacityID:       {ID: UseSliceCapacityID, Name: "use-slice-capacity", Template: "provide slice capacity when the length is known upfront"},
	MaxParamsID:              {ID: MaxParamsID, Name: "max-params", Template: "function exceeds the parameter limit"},
	AvoidEmptyStructsID:      {ID: AvoidEmptyStructsID, Name: "avoid-empty-structs", Template: "empty struct declarations are not allowed"},
	AlwaysPreferConstID:      {ID: AlwaysPreferConstID, Name: "always-prefer-const", Template: "replace variable with a constant"},
	GetMustReturnValueID:     {ID: GetMustReturnValueID, Name: "get-must-return-value", Template: `functions whose names start with "Get" should return at least one non-error value`},

	// --- CORRECTNESS ---
	UnusedReceiverID:         {ID: UnusedReceiverID, Name: "unused-receiver", Template: "receiver %q is never used"},
	UnusedParamsID:           {ID: UnusedParamsID, Name: "unused-params", Template: "parameter %q is never used"},
	EmptyBlockID:             {ID: EmptyBlockID, Name: "empty-block", Template: "empty block; remove it or add a clarifying comment"},
	AmbiguousReturnID:        {ID: AmbiguousReturnID, Name: "ambiguous-return", Template: "function returns too many unnamed values of the same type"},
	BoolLiteralExpressionsID: {ID: BoolLiteralExpressionsID, Name: "boolean-literal-expressions", Template: "simplify boolean literal expressions"},

	// --- COMPLEXITY ---
	MaxFuncLinesID:         {ID: MaxFuncLinesID, Name: "max-func-lines", Template: "function exceeds the line limit"},
	MaxLineLengthID:        {ID: MaxLineLengthID, Name: "max-line-length", Template: "line has %d characters; limit is %d"},
	MaxNestingDepthID:      {ID: MaxNestingDepthID, Name: "max-nesting-depth", Template: "nesting depth exceeds the limit of %d"},
	CyclomaticComplexityID: {ID: CyclomaticComplexityID, Name: "cyclomatic-complexity", Template: "cyclomatic complexity is %d; limit is %d"},

	// --- NAMING ---
	ReceiverNameID:        {ID: ReceiverNameID, Name: "receiver-name", Template: "receiver name does not follow the configured convention"},
	ExportedIdentifiersID: {ID: ExportedIdentifiersID, Name: "exported-identifiers", Template: "exported identifier should have a doc comment"},
	ImportedIdentifiersID: {ID: ImportedIdentifiersID, Name: "imported-identifiers", Template: "import alias does not match the configured naming rule"},

	// ---- STYLE ---
	PreferIncDecID: {ID: PreferIncDecID, Name: "prefer-inc-dec", Template: "use ++ or -- instead of += 1 or -= 1"},

	// ---- SUPPRESSION ----
	UnusedSuppressionID:       {ID: UnusedSuppressionID, Name: "unused-suppression", Template: "suppression for rule %q does not match any issue"},
	MisplacedFileWideIgnoreID: {ID: MisplacedFileWideIgnoreID, Name: "misplaced-file-wide-ignore", Template: "file-wide suppression for rule %q must appear before the package declaration"},
}

func GetMetadata(id uint16) (RuleMetadata, bool) {
	m, ok := registry[id]
	return m, ok
}

func IsFixable(id uint16) bool {
	meta, ok := GetMetadata(id)
	return ok && meta.Fixable
}

func FormatMessage(issue Issue) string {
	meta, ok := GetMetadata(issue.ID)

	if !ok {
		return fmt.Sprintf("unknown rule (id %d)", issue.ID)
	}

	switch issue.ID {
	case ErrorStringFormatID:
		return formatErrorStringMessage(issue)

	case ErrorNotWrappedID:
		return formatErrorNotWrappedMessage(issue)

	case NoDotImportsID:
		if issue.ArgStr1 != "" {
			return fmt.Sprintf("dot import of package %q is not allowed", issue.ArgStr1)
		}

		return meta.Template

	case RedundantImportAliasID:
		alias, path := SplitContext2(issue.ArgStr1)
		if alias != "" && path != "" {
			return fmt.Sprintf("import alias %q is redundant for package %q", alias, path)
		}
		if alias != "" {
			return fmt.Sprintf("import alias %q is redundant", alias)
		}

		return meta.Template

	case NoDeferInLoopID:
		return withFunction(issue.ArgStr1, "avoid defer inside loops")

	case UseContextInFirstParamID:
		return formatContextFirstParamMessage(issue)

	case NoBareReturnsID:
		return withFunction(issue.ArgStr1, "avoid bare returns")

	case NoMagicNumbersID:
		return formatMagicNumberMessage(issue)

	case UseSliceCapacityID:
		return formatUseSliceCapacityMessage(issue)

	case MaxParamsID, MaxFuncLinesID, MaxLineLengthID:
		return formatCountLimitMessage(issue)

	case CyclomaticComplexityID, AmbiguousReturnID:
		return formatComplexityMessage(issue)

	case MaxNestingDepthID:
		return fmt.Sprintf(meta.Template, issue.ArgInt1)

	case AvoidEmptyStructsID:
		if issue.ArgStr1 != "" {
			return fmt.Sprintf("struct %q is empty; add fields or use an explicit marker type", issue.ArgStr1)
		}

		return meta.Template

	case AlwaysPreferConstID:
		name, fn := SplitContext2(issue.ArgStr1)
		if name != "" && fn != "" {
			return fmt.Sprintf("replace variable %q with a constant in function %q", name, fn)
		}
		if name != "" {
			return fmt.Sprintf("replace variable %q with a constant", name)
		}

		return meta.Template

	case GetMustReturnValueID:
		if issue.ArgStr1 != "" {
			return fmt.Sprintf("function %q starts with %q and should return at least one non-error value", issue.ArgStr1, "Get")
		}

		return meta.Template

	case EmptyBlockID:
		if issue.ArgStr1 != "" {
			return fmt.Sprintf("empty block in function %q; remove it or add a clarifying comment", issue.ArgStr1)
		}

		return meta.Template

	case BoolLiteralExpressionsID:
		return withFunction(issue.ArgStr1, "simplify boolean literal expressions")

	case NoErrorShadowingID,
		UnusedReceiverID,
		UnusedParamsID,
		DisallowedPackagesID,
		UnusedSuppressionID,
		MisplacedFileWideIgnoreID:
		return fmt.Sprintf(meta.Template, issue.ArgStr1)

	case ReceiverNameID:
		return formatReceiverNameMessage(issue)

	case ExportedIdentifiersID:
		return formatExportedIdentifierMessage(issue)

	case ImportedIdentifiersID:
		alias, path := SplitContext2(issue.ArgStr1)
		if alias != "" && path != "" {
			return fmt.Sprintf("import alias %q for package %q does not match the configured naming rule", alias, path)
		}
		if alias != "" {
			return fmt.Sprintf("import alias %q does not match the configured naming rule", alias)
		}

		return meta.Template

	case PreferIncDecID:
		name, fn := SplitContext2(issue.ArgStr1)
		if name != "" && fn != "" {
			return fmt.Sprintf("use ++ or -- instead of += 1 or -= 1 for %q in function %q", name, fn)
		}
		if name != "" {
			return fmt.Sprintf("use ++ or -- instead of += 1 or -= 1 for %q", name)
		}

		return meta.Template

	default:
		return meta.Template
	}
}

func withFunction(functionName, message string) string {
	if functionName == "" {
		return message
	}

	return fmt.Sprintf("%s in function %q", message, functionName)
}

func formatErrorStringMessage(issue Issue) string {
	msg, fn := SplitContext2(issue.ArgStr1)
	if msg == "" {
		return registry[ErrorStringFormatID].Template
	}
	if fn != "" {
		return fmt.Sprintf("error message %q in function %q should start with a lowercase letter and should not end with punctuation", msg, fn)
	}

	return fmt.Sprintf("error message %q should start with a lowercase letter and should not end with punctuation", msg)
}

func formatErrorNotWrappedMessage(issue Issue) string {
	name, fn := SplitContext2(issue.ArgStr1)
	if name == "" {
		return registry[ErrorNotWrappedID].Template
	}
	if fn != "" {
		return fmt.Sprintf("error value %q in function %q should be wrapped before it is returned", name, fn)
	}

	return fmt.Sprintf("error value %q should be wrapped before it is returned", name)
}

func formatContextFirstParamMessage(issue Issue) string {
	paramName, fn := SplitContext2(issue.ArgStr1)
	position := issue.ArgInt1

	if paramName != "" && fn != "" {
		return fmt.Sprintf("parameter %q in function %q has type context.Context and must be the first parameter", paramName, fn)
	}
	if paramName != "" && position > 0 {
		return fmt.Sprintf("parameter %q has type context.Context and must be the first parameter (currently at position #%d)", paramName, position)
	}
	if fn != "" && position > 0 {
		return fmt.Sprintf("context.Context is parameter #%d in function %q; move it to the first position", position, fn)
	}
	if fn != "" {
		return fmt.Sprintf("context.Context should be the first parameter in function %q", fn)
	}

	return registry[UseContextInFirstParamID].Template
}

func formatMagicNumberMessage(issue Issue) string {
	value, fn := SplitContext2(issue.ArgStr1)
	if value == "" {
		return registry[NoMagicNumbersID].Template
	}
	if fn != "" {
		return fmt.Sprintf("extract magic number %q into a named constant in function %q", value, fn)
	}

	return fmt.Sprintf("extract magic number %q into a named constant", value)
}

func formatUseSliceCapacityMessage(issue Issue) string {
	target, fn := SplitContext2(issue.ArgStr1)
	if target != "" && fn != "" {
		return fmt.Sprintf("provide slice capacity when initializing %q in function %q", target, fn)
	}
	if target != "" {
		return fmt.Sprintf("provide slice capacity when initializing %q", target)
	}
	if fn != "" {
		return fmt.Sprintf("provide slice capacity in function %q when the length is known upfront", fn)
	}

	return registry[UseSliceCapacityID].Template
}

func formatCountLimitMessage(issue Issue) string {
	switch issue.ID {
	case MaxParamsID:
		if issue.ArgStr1 != "" {
			return fmt.Sprintf("function %q has %d parameters; limit is %d", issue.ArgStr1, issue.ArgInt2, issue.ArgInt1)
		}

		return fmt.Sprintf("function has %d parameters; limit is %d", issue.ArgInt2, issue.ArgInt1)

	case MaxFuncLinesID:
		if issue.ArgStr1 != "" {
			return fmt.Sprintf("function %q has %d lines; limit is %d", issue.ArgStr1, issue.ArgInt2, issue.ArgInt1)
		}

		return fmt.Sprintf("function has %d lines; limit is %d", issue.ArgInt2, issue.ArgInt1)

	default:
		return fmt.Sprintf(registry[MaxLineLengthID].Template, issue.ArgInt2, issue.ArgInt1)
	}
}

func formatComplexityMessage(issue Issue) string {
	switch issue.ID {
	case AmbiguousReturnID:
		fn, typ := SplitContext2(issue.ArgStr1)
		if fn != "" && typ != "" {
			return fmt.Sprintf("function %q returns %d unnamed values of type %q; limit is %d", fn, issue.ArgInt1, typ, issue.ArgInt2)
		}
		if fn != "" {
			return fmt.Sprintf("function %q returns %d unnamed values of the same type; limit is %d", fn, issue.ArgInt1, issue.ArgInt2)
		}

		return fmt.Sprintf("function returns %d unnamed values of the same type; limit is %d", issue.ArgInt1, issue.ArgInt2)

	default:
		return fmt.Sprintf(registry[CyclomaticComplexityID].Template, issue.ArgInt1, issue.ArgInt2)
	}
}

func formatReceiverNameMessage(issue Issue) string {
	name, method := SplitContext2(issue.ArgStr1)
	if name == "" {
		return registry[ReceiverNameID].Template
	}
	if method != "" && issue.ArgInt1 > 0 {
		return fmt.Sprintf("receiver name %q in method %q exceeds the configured limit of %d characters", name, method, issue.ArgInt1)
	}
	if method != "" {
		return fmt.Sprintf("receiver name %q in method %q does not follow the configured convention", name, method)
	}

	return fmt.Sprintf("receiver name %q does not follow the configured convention", name)
}

func formatExportedIdentifierMessage(issue Issue) string {
	kind, name := SplitContext2(issue.ArgStr1)
	if name == "" {
		name = kind
		kind = ""
	}

	if kind != "" && name != "" {
		return fmt.Sprintf("exported %s %q should have a doc comment", kind, name)
	}
	if name != "" {
		return fmt.Sprintf("exported identifier %q should have a doc comment", name)
	}

	return registry[ExportedIdentifiersID].Template
}
