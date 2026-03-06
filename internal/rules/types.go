package rules

import (
	"go/ast"
)

type Rule interface {
	Name() string
	Targets() []ast.Node
	Run(runner *Runner, node ast.Node)
}
