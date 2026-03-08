package errs

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
	"unicode"

	"github.com/serenitysz/serenity/internal/rules"
)

type ErrorStringFormatRule struct {
	Severity rules.Severity
}

func (r *ErrorStringFormatRule) Name() string {
	return "error-string-format"
}

func (r *ErrorStringFormatRule) Targets() []ast.Node {
	return []ast.Node{(*ast.ReturnStmt)(nil)}
}

func (r *ErrorStringFormatRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	if runner.ReachedMax() {
		return
	}

	ret := node.(*ast.ReturnStmt)

	for _, res := range ret.Results {
		call, ok := res.(*ast.CallExpr)

		if !ok || !isErrorConstructor(call) || len(call.Args) == 0 {
			continue
		}

		lit, ok := call.Args[0].(*ast.BasicLit)

		if !ok || lit.Kind != token.STRING {
			continue
		}

		msg := strings.Trim(lit.Value, `"`)

		if isValidErrorString(msg) {
			continue
		}

		issue := rules.Issue{
			ArgStr1:  rules.PackContext2(msg, rules.CurrentFunctionName(runner)),
			ID:       rules.ErrorStringFormatID,
			Severity: r.Severity,
		}

		if runner.ShouldAutofix() {
			runner.Modified = true
			lit.Value = strconv.Quote(fixErrorString(msg))
			runner.ReportFixed(lit.Pos(), issue)
			return
		}

		runner.ReportFixable(lit.Pos(), issue)
	}
}

func fixErrorString(s string) string {
	if s == "" {
		return s
	}

	for len(s) > 0 {
		last := s[len(s)-1]

		if strings.ContainsRune(".!?:;", rune(last)) {
			s = s[:len(s)-1]
		} else {
			break
		}
	}

	r := []rune(s)
	r[0] = unicode.ToLower(r[0])

	return string(r)
}
