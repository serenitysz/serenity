package errs

import (
	"go/ast"
	"strings"
	"unicode"
)

func isErrorConstructor(call *ast.CallExpr) bool {
	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		if pkg, ok := fun.X.(*ast.Ident); ok {
			return (pkg.Name == "errors" && fun.Sel.Name == "New") ||
				(pkg.Name == "fmt" && fun.Sel.Name == "Errorf")
		}
	}

	return false
}

func isValidErrorString(s string) bool {
	if s == "" {
		return true
	}

	last := s[len(s)-1]

	if strings.ContainsRune(".!?:;", rune(last)) {
		return false
	}

	first := rune(s[0])

	if unicode.IsUpper(first) && !isAllCapsWord(s) {
		return false
	}

	return true
}

func isAllCapsWord(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return false
		}
	}

	return true
}
