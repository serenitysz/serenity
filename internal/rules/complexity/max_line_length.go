package complexity

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

type CheckMaxLineLengthRule struct {
	Limit    int
	Severity rules.Severity
}

func (c *CheckMaxLineLengthRule) Name() string {
	return "max-line-length"
}

func (c *CheckMaxLineLengthRule) Targets() []ast.Node {
	return []ast.Node{(*ast.File)(nil)}
}

func (c *CheckMaxLineLengthRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	if runner.ReachedMax() {
		return
	}

	tokFile := runner.Fset.File(node.Pos())
	if tokFile == nil {
		return
	}

	offsets := tokFile.Lines()
	fSize := tokFile.Size()
	lineCount := len(offsets)
	eof := tokFile.Position(tokFile.Pos(fSize))

	for i := range lineCount {
		if runner.ReachedMax() {
			return
		}

		var lineLen int

		isLastLine := i == lineCount-1
		if isLastLine {
			if eof.Line == i+1 {
				lineLen = eof.Column - 1
			} else {
				lineLen = fSize - offsets[i]
				if lineLen > 0 {
					lineLen--
				}
			}
		} else {
			lineLen = offsets[i+1] - offsets[i] - 1
		}

		if lineLen > c.Limit {
			runner.Report(tokFile.LineStart(i+1), rules.Issue{
				ID:       rules.MaxLineLengthID,
				Severity: c.Severity,
				ArgInt1:  uint32(c.Limit),
				ArgInt2:  uint32(lineLen),
			})
		}
	}
}
