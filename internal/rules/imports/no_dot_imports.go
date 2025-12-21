package imports

import (
	"github.com/serenitysz/serenity/internal/rules"
)

func CheckNoDotImports(runner *rules.Runner) []rules.Issue {
	issues := make([]rules.Issue, 0, len(runner.File.Imports))
	imports := runner.Cfg.Linter.Rules.Imports

	if imports == nil ||
		(imports.Use != nil && !*imports.Use) ||
		imports.NoDotImports == nil {
		return issues
	}

	maxIssues := rules.GetMaxIssues(runner.Cfg)

	for _, i := range runner.File.Imports {
		if i.Name != nil && i.Name.Name == "." {

			if maxIssues > 0 && int16(len(issues)) >= maxIssues {
				break
			}

			issues = append(issues, rules.Issue{
				ID:      rules.NoDotImportsID,
				Pos:     runner.Fset.Position(i.Name.NamePos),
				Message: "Imports should not be named with '.'",
			})
		}
	}

	return issues
}
