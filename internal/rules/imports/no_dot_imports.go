package imports

import (
	"go/ast"

	"github.com/serenitysz/serenity/internal/rules"
)

func CheckNoDotImports(runner *rules.Runner) []rules.Issue {
	var issues []rules.Issue
	imports := runner.Cfg.Linter.Rules.Imports

	if err := rules.VerifyIssues(runner.Cfg, issues); err != nil {
		return issues
	}

	if imports == nil ||
		(imports.Use != nil && !*imports.Use) ||
		imports.NoDotImports == nil {
		return issues
	}

	for _, i := range runner.File.Imports {
		if i.Name != nil && i.Name.Name == "." {
			issues = append(issues, rules.Issue{
				Pos:     runner.Fset.Position(i.Name.NamePos),
				Message: "Imports should not be named with '.' ",
				Fix: func() {
					FixNoDotImports(runner, i)
				},
			})
		}
	}

	return issues
}

func FixNoDotImports(runner *rules.Runner, i *ast.ImportSpec) {
	i.Name = nil
}
