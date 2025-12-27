package bestpractices

import (
	"go/ast"
	"go/token"

	"github.com/serenitysz/serenity/internal/rules"
)

type NoMagicNumbersRule struct{}

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

	lit := node.(*ast.BasicLit)
	if lit == nil {
		return
	}

	bp := runner.Cfg.Linter.Rules.BestPractices
	if bp == nil || (bp.Use != nil && !*bp.Use) || bp.NoMagicNumbers == nil {
		return
	}

	if lit.Kind != token.INT && lit.Kind != token.FLOAT {
		return
	}

	maxIssues := rules.GetMaxIssues(runner.Cfg)
	if maxIssues > 0 && int16(len(*runner.Issues)) >= maxIssues {
		return
	}

	if lit.Value == "0" || lit.Value == "1" || lit.Value == "-1" {
		return
	}

	*runner.Issues = append(*runner.Issues, rules.Issue{
		ID:       rules.NoMagicNumbersID,
		Pos:      runner.Fset.Position(lit.Pos()),
		Severity: rules.ParseSeverity(bp.NoMagicNumbers.Severity),
		ArgStr1:  lit.Value,
	})
}
