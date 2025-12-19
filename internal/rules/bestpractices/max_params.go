package bestpractices

import (
	"go/ast"
	"go/token"
	"unicode"

	"github.com/serenitysz/serenity/internal/rules"
)

func CheckMaxParamsNode(runner *rules.Runner) []rules.Issue {
	bestPractices := runner.Cfg.Linter.Rules.BestPractices

	if bestPractices == nil {
		return nil
	}

	if bestPractices.Use != nil && !*bestPractices.Use {
		return nil
	}

	var limit int8 = 5

	if bestPractices.MaxParams != nil &&
		bestPractices.MaxParams.Quantity != nil {
		limit = *bestPractices.MaxParams.Quantity
	}

	fn, ok := runner.Node.(*ast.FuncDecl)
	if !ok {
		return nil
	}

	params := fn.Type.Params
	if params == nil {
		return nil
	}

	var count int16 = 0

	for _, field := range params.List {
		count += int16(len(field.Names))

		if len(field.Names) == 0 {
			count++
		}
	}

	if limit > 0 && int8(count) <= limit {
		return nil
	}

	return []rules.Issue{{
		Pos:     runner.Fset.Position(fn.Pos()),
		Message: "functions exceed the maximum parameter limit",
		Fix: func() {
			FixMaxParams(runner, fn, params)
		},
	}}
}

func FixMaxParams(runner *rules.Runner, fn *ast.FuncDecl, params *ast.FieldList) {
	structName := fn.Name.Name + "Params"

	var newFields []*ast.Field
	for _, param := range params.List {
		var names []*ast.Ident

		for _, name := range param.Names {
			runes := []rune(name.Name)
			if len(runes) > 0 {
				runes[0] = unicode.ToUpper(runes[0])
			}
			names = append(names, ast.NewIdent(string(runes)))
		}

		newFields = append(newFields, &ast.Field{
			Names: names,
			Type:  param.Type,
		})
	}

	typeSpec := &ast.TypeSpec{
		Name: ast.NewIdent(structName),
		Type: &ast.StructType{
			Fields: &ast.FieldList{List: newFields},
		},
	}

	decl := &ast.GenDecl{
		Tok:   token.TYPE,
		Specs: []ast.Spec{typeSpec},
	}

	insertIdx := 0
	for i, d := range runner.File.Decls {
		if g, ok := d.(*ast.GenDecl); ok && g.Tok == token.IMPORT {
			insertIdx = i + 1
		}
	}

	runner.File.Decls = append(runner.File.Decls[:insertIdx], append([]ast.Decl{decl}, runner.File.Decls[insertIdx:]...)...)

	newParam := &ast.Field{
		Names: []*ast.Ident{ast.NewIdent("params")},
		Type:  ast.NewIdent(structName),
	}

	fn.Type.Params.List = []*ast.Field{newParam}
}
