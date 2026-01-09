package naming

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

type ReceiverNamesRule struct{}

func (r *ReceiverNamesRule) Name() string {
	return "receiver-names"
}

func (r *ReceiverNamesRule) Targets() []ast.Node {
	return []ast.Node{(*ast.FuncDecl)(nil)}
}

func (r *ReceiverNamesRule) Run(runner *rules.Runner, node ast.Node) {
	if runner.ShouldStop != nil && runner.ShouldStop() {
		return
	}

	naming := runner.Cfg.Linter.Rules.Naming

	if naming == nil || !naming.Use || naming.ReceiverNames == nil {
		return
	}

	if max := runner.Cfg.GetMaxIssues(); max > 0 && *runner.IssuesCount >= max {
		return
	}

	fn := node.(*ast.FuncDecl)

	if fn.Recv == nil && len(fn.Recv.List) != 1 {
		return
	}

	recv := fn.Recv.List[0]

	if len(recv.Names) > 0 && len(recv.Names[0].Name) > *naming.ReceiverNames.MaxSize {
		*runner.IssuesCount++

		*runner.Issues = append(*runner.Issues, rules.Issue{
			ArgStr1:  recv.Names[0].Name,
			ID:       rules.ReceiverNameID,
			Pos:      runner.Fset.Position(recv.Pos()),
			Severity: rules.ParseSeverity(naming.ReceiverNames.Severity),
		})
	}
}
