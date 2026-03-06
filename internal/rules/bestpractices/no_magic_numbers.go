package bestpractices

import (
	"go/ast"
	"go/token"

	"github.com/serenitysz/serenity/internal/rules"
)

type NoMagicNumbersRule struct {
	Severity rules.Severity
}

func (n *NoMagicNumbersRule) Name() string {
	return "no-magic-number"
}

func (n *NoMagicNumbersRule) Targets() []ast.Node {
	return []ast.Node{(*ast.BasicLit)(nil)}
}

func (n *NoMagicNumbersRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	if runner.ReachedMax() {
		return
	}

	lit := node.(*ast.BasicLit)

	if lit == nil {
		return
	}

	if lit.Kind != token.INT && lit.Kind != token.FLOAT {
		return
	}

	if lit.Value == "0" || lit.Value == "1" || lit.Value == "-1" {
		return
	}

	runner.Report(lit.Pos(), rules.Issue{
		ID:       rules.NoMagicNumbersID,
		Severity: n.Severity,
		ArgStr1:  lit.Value,
	})
}
