package rules

import (
	"go/ast"
	"reflect"
)

type Rule interface {
	Name() string
	Targets() []ast.Node
	Run(runner *Runner, node ast.Node)
}

func GetNodeType(node ast.Node) reflect.Type {
	return reflect.TypeOf(node)
}
