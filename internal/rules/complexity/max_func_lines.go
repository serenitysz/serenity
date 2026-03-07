package complexity

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

type CheckMaxFuncLinesRule struct {
	Limit    int16
	Severity rules.Severity
}

func (c *CheckMaxFuncLinesRule) Name() string {
	return "max-func-lines"
}

func (c *CheckMaxFuncLinesRule) Targets() []ast.Node {
	return []ast.Node{(*ast.FuncDecl)(nil)}
}

func (c *CheckMaxFuncLinesRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	if runner.ReachedMax() {
		return
	}

	fn := node.(*ast.FuncDecl)

	if fn.Body == nil {
		return
	}

	end := runner.Fset.Position(fn.End()).Line
	start := runner.Fset.Position(fn.Pos()).Line

	linesCount := end - start + 1

	if int16(linesCount) <= c.Limit {
		return
	}

	runner.Report(fn.Pos(), rules.Issue{
		ArgStr1:  fn.Name.Name,
		ArgInt1:  uint32(c.Limit),
		ArgInt2:  uint32(linesCount),
		ID:       rules.MaxFuncLinesID,
		Severity: c.Severity,
	})
}
