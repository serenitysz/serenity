package rules

import "testing"

func TestFormatMessageUsesReadableTemplates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		issue Issue
		want  string
	}{
		{
			name:  "max params",
			issue: Issue{ID: MaxParamsID, ArgStr1: "Handle", ArgInt1: 5, ArgInt2: 7},
			want:  "function \"Handle\" has 7 parameters; limit is 5",
		},
		{
			name:  "max line length",
			issue: Issue{ID: MaxLineLengthID, ArgInt1: 100, ArgInt2: 188},
			want:  "line has 188 characters; limit is 100",
		},
		{
			name:  "context first param",
			issue: Issue{ID: UseContextInFirstParamID, ArgStr1: PackContext2("ctx", "Handle"), ArgInt1: 2},
			want:  "parameter \"ctx\" in function \"Handle\" has type context.Context and must be the first parameter",
		},
		{
			name:  "slice capacity target",
			issue: Issue{ID: UseSliceCapacityID, ArgStr1: PackContext2("buf", "Handle")},
			want:  "provide slice capacity when initializing \"buf\" in function \"Handle\"",
		},
		{
			name:  "cyclomatic complexity",
			issue: Issue{ID: CyclomaticComplexityID, ArgInt1: 12, ArgInt2: 10},
			want:  "cyclomatic complexity is 12; limit is 10",
		},
		{
			name:  "ambiguous return",
			issue: Issue{ID: AmbiguousReturnID, ArgStr1: PackContext2("Read", "string"), ArgInt1: 3, ArgInt2: 1},
			want:  "function \"Read\" returns 3 unnamed values of type \"string\"; limit is 1",
		},
		{
			name:  "exported identifier rule",
			issue: Issue{ID: ExportedIdentifiersID, ArgStr1: PackContext2("type", "Widget")},
			want:  "exported type \"Widget\" should have a doc comment",
		},
		{
			name:  "backward compatible exported identifier rule",
			issue: Issue{ID: ExportedIdentifiersID, ArgStr1: "Widget"},
			want:  "exported identifier \"Widget\" should have a doc comment",
		},
		{
			name:  "error string format",
			issue: Issue{ID: ErrorStringFormatID, ArgStr1: PackContext2("Bad.", "Create")},
			want:  "error message \"Bad.\" in function \"Create\" should start with a lowercase letter and should not end with punctuation",
		},
		{
			name:  "prefer inc dec",
			issue: Issue{ID: PreferIncDecID, ArgStr1: PackContext2("count", "Handle")},
			want:  "use ++ or -- instead of += 1 or -= 1 for \"count\" in function \"Handle\"",
		},
		{
			name:  "unknown rule",
			issue: Issue{ID: 65535},
			want:  "unknown rule (id 65535)",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := FormatMessage(tt.issue); got != tt.want {
				t.Fatalf("unexpected message: got %q want %q", got, tt.want)
			}
		})
	}
}
