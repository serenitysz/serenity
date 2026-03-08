package errs

import (
	"go/ast"
	"go/token"
	"strconv"
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

func ensureImport(file *ast.File, path string) string {
	if file == nil {
		return ""
	}

	quoted := strconv.Quote(path)

	for _, spec := range file.Imports {
		if spec.Path == nil || spec.Path.Value != quoted {
			continue
		}

		if spec.Name == nil {
			return importDefaultName(path)
		}

		switch spec.Name.Name {
		case ".", "_":
			spec.Name = nil
			return importDefaultName(path)
		default:
			return spec.Name.Name
		}
	}

	newSpec := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: quoted,
		},
	}
	newDecl := &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: []ast.Spec{newSpec},
	}

	insertAt := 0
	for insertAt < len(file.Decls) {
		gen, ok := file.Decls[insertAt].(*ast.GenDecl)
		if !ok || gen.Tok != token.IMPORT {
			break
		}

		insertAt++
	}

	file.Decls = append(file.Decls, nil)
	copy(file.Decls[insertAt+1:], file.Decls[insertAt:])
	file.Decls[insertAt] = newDecl
	file.Imports = append(file.Imports, newSpec)

	return importDefaultName(path)
}

func importDefaultName(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx >= 0 {
		return path[idx+1:]
	}

	return path
}
