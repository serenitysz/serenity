package imports

import (
	"github.com/serenitysz/serenity/internal/rules"
)

func CheckNoDotImports(runner *rules.Runner) {
	imports := runner.Cfg.Linter.Rules.Imports

	if imports == nil ||
		(imports.Use != nil && !*imports.Use) ||
		imports.NoDotImports == nil {
		return
	}

	maxIssues := rules.GetMaxIssues(runner.Cfg)

	for _, i := range runner.File.Imports {
		if i.Name != nil && i.Name.Name == "." {

			if maxIssues > 0 && int16(len(*runner.Issues)) >= maxIssues {
				break
			}

			*runner.Issues = append(*runner.Issues, rules.Issue{
				ID:       rules.NoDotImportsID,
				Pos:      runner.Fset.Position(i.Name.NamePos),
				Severity: rules.ParseSeverity(imports.NoDotImports.Severity),
			})
		}
	}
}
