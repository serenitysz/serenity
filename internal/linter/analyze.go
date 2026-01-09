package linter

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"os"
	"reflect"

	"github.com/serenitysz/serenity/internal/rules"
)

func (l *Linter) Analyze(params AnalysisParams) ([]rules.Issue, error) {
	conf := types.Config{
		Importer: nil,
		Error:    func(err error) {},
	}

	info := &types.Info{
		Defs: make(map[*ast.Ident]types.Object),
		Uses: make(map[*ast.Ident]types.Object),
	}

	if len(params.pkgFiles) > 0 {
		conf.Check(params.pkgPaths[0], params.fset, params.pkgFiles, info)
	}

	mutatedObjects := make(map[types.Object]bool)

	for _, f := range params.pkgFiles {
		ast.Inspect(f, func(n ast.Node) bool {
			switch t := n.(type) {
			case *ast.AssignStmt:
				for _, lhs := range t.Lhs {
					if id, ok := lhs.(*ast.Ident); ok {
						if obj := info.Uses[id]; obj != nil {
							mutatedObjects[obj] = true
						}
					}
				}
			case *ast.IncDecStmt:
				if id, ok := t.X.(*ast.Ident); ok {
					if obj := info.Uses[id]; obj != nil {
						mutatedObjects[obj] = true
					}
				}
			case *ast.UnaryExpr:
				if t.Op == token.AND {
					if id, ok := t.X.(*ast.Ident); ok {
						if obj := info.Uses[id]; obj != nil {
							mutatedObjects[obj] = true
						}
					}
				}
			}
			return true
		})
	}

	estimatedIssues := len(params.pkgFiles) * 8
	allIssues := make([]rules.Issue, 0, estimatedIssues)

	for i, f := range params.pkgFiles {
		filePath := params.pkgPaths[i]

		issues := make([]rules.Issue, 0, FINAL_FILE_ISSUE_CAP)

		runner := rules.Runner{
			File:           f,
			Fset:           params.fset,
			Cfg:            l.Config,
			Unsafe:         l.Unsafe,
			Issues:         &issues,
			IssuesCount:    new(uint16),
			MutatedObjects: mutatedObjects,
			Autofix:        l.Write || l.Config.ShouldAutofix(),
			ShouldStop: func() bool {
				return params.shouldStop != nil && params.shouldStop(len(allIssues)+len(issues))
			},
			TypesInfo: info,
		}

		ast.Inspect(f, func(n ast.Node) bool {
			if n == nil {
				return true
			}

			nodeType := reflect.TypeOf(n)
			if specificRules, found := params.rules[nodeType]; found {
				for _, rule := range specificRules {
					rule.Run(&runner, n)
				}
			}

			return true
		})

		allIssues = append(allIssues, issues...)

		if runner.Modified {
			var buf bytes.Buffer

			if err := format.Node(&buf, params.fset, f); err == nil {
				if err := os.WriteFile(filePath, buf.Bytes(), DEFAULT_FILE_MODE); err != nil {
					return allIssues, fmt.Errorf("failed to write file %s: %w", filePath, err)
				}
			}
		}
	}

	return allIssues, nil
}
