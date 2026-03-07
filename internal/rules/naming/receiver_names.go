package naming

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

type ReceiverNamesRule struct {
	Severity rules.Severity
	MaxSize  int
}

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

	if runner.ReachedMax() {
		return
	}

	fn := node.(*ast.FuncDecl)

	if fn.Recv == nil || len(fn.Recv.List) != 1 {
		return
	}

	recv := fn.Recv.List[0]

	if len(recv.Names) > 0 && len(recv.Names[0].Name) > r.MaxSize {
		runner.Report(recv.Pos(), rules.Issue{
			ArgStr1:  rules.PackContext2(recv.Names[0].Name, fn.Name.Name),
			ArgInt1:  uint32(r.MaxSize),
			ID:       rules.ReceiverNameID,
			Severity: r.Severity,
		})
	}
}
