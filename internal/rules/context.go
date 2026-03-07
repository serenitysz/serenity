package rules

import (
	"go/ast"
	"strings"
)

const messageContextSeparator = "\x1f"

func CurrentFunctionName(runner *Runner) string {
	if runner == nil || runner.CurrentFunc == nil {
		return ""
	}

	name := runner.CurrentFunc.Name
	if name == "" {
		return ""
	}

	if name == "anonymous" {
		return "anonymous function"
	}

	return name
}

func PackContext2(first, second string) string {
	if first == "" && second == "" {
		return ""
	}

	return first + messageContextSeparator + second
}

func SplitContext2(value string) (string, string) {
	idx := strings.IndexByte(value, messageContextSeparator[0])
	if idx < 0 {
		return value, ""
	}

	return value[:idx], value[idx+1:]
}

func ExprName(expr ast.Expr) string {
	switch n := expr.(type) {
	case *ast.Ident:
		return n.Name
	case *ast.SelectorExpr:
		left := ExprName(n.X)
		if left == "" {
			return n.Sel.Name
		}

		return left + "." + n.Sel.Name
	case *ast.StarExpr:
		return ExprName(n.X)
	default:
		return ""
	}
}

func AssignmentTargetName(parent ast.Node, value ast.Expr) string {
	switch n := parent.(type) {
	case *ast.AssignStmt:
		for i, rhs := range n.Rhs {
			if rhs != value {
				continue
			}

			if i < len(n.Lhs) {
				return ExprName(n.Lhs[i])
			}

			return ""
		}
	case *ast.ValueSpec:
		for i, rhs := range n.Values {
			if rhs != value {
				continue
			}

			if i < len(n.Names) {
				return n.Names[i].Name
			}

			return ""
		}
	}

	return ""
}
