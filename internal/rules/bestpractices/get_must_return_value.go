package bestpractices

import (
	"go/ast"
	"strings"

	"github.com/serenitysz/serenity/internal/rules"
)

type GetMustReturnValueRule struct {
	Severity rules.Severity
}

func (r *GetMustReturnValueRule) Name() string {
	return "get-must-return-value"
}

func (r *GetMustReturnValueRule) Targets() []ast.Node {
	return []ast.Node{(*ast.FuncDecl)(nil)}
}

func (r *GetMustReturnValueRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	if runner.ReachedMax() {
		return
	}

	fn := node.(*ast.FuncDecl)

	if fn.Name == nil || !strings.HasPrefix(fn.Name.Name, "Get") {
		return
	}

	results := fn.Type.Results

	if results == nil || len(results.List) == 0 {
		r.report(runner, fn)
		return
	}

	nonErrorReturns := 0

	for _, field := range results.List {
		if isErrorType(field.Type) {
			continue
		}

		count := max(1, len(field.Names))

		nonErrorReturns += count

		if nonErrorReturns > 0 {
			return
		}
	}

	r.report(runner, fn)
}

func (r *GetMustReturnValueRule) report(runner *rules.Runner, fn *ast.FuncDecl) {
	runner.Report(fn.Name.Pos(), rules.Issue{
		ID:       rules.GetMustReturnValueID,
		Severity: r.Severity,
	})
}

func isErrorType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name == "error"
	case *ast.SelectorExpr:
		return false
	}

	return false
}
