package rules

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"
)

type Suppression struct {
	RuleName    string
	Reason      string
	Line        int
	IsFileWide  bool // If true, applies to all occurrences in the file
	IsMisplaced bool // If true, file-wide ignore is placed after package declaration
}

var (
	inlineRegex   = regexp.MustCompile(`@serenity-ignore\s+([\w-]+)(?::\s*(.+))?`)
	fileWideRegex = regexp.MustCompile(`@serenity-ignore-all\s+([\w-]+)(?::\s*(.+))?`)
)

func ProcessSuppressions(comments []*ast.CommentGroup, fset *token.FileSet, decls []ast.Decl, pkgPos token.Pos) []Suppression {
	var suppressions []Suppression

	pkgLine := fset.Position(pkgPos).Line

	firstDeclLine := 0
	if len(decls) > 0 {
		firstDeclLine = fset.Position(decls[0].Pos()).Line
	}

	for _, cg := range comments {
		if cg == nil {
			continue
		}

		for _, comment := range cg.List {
			if comment == nil {
				continue
			}

			line := fset.Position(comment.Pos()).Line

			sup := parseSuppression(comment.Text, line, firstDeclLine, pkgLine)
			if sup != nil {
				suppressions = append(suppressions, *sup)
			}
		}
	}

	return suppressions
}

func parseSuppression(text string, line int, firstDeclLine int, pkgLine int) *Suppression {
	text = strings.TrimSpace(text)

	if match := fileWideRegex.FindStringSubmatch(text); match != nil {
		misplaced := line >= pkgLine

		ruleName := match[1]
		reason := ""
		if len(match) > 2 {
			reason = match[2]
		}

		return &Suppression{
			RuleName:    ruleName,
			Reason:      reason,
			Line:        line,
			IsFileWide:  true,
			IsMisplaced: misplaced,
		}
	}

	if match := inlineRegex.FindStringSubmatch(text); match != nil {
		ruleName := match[1]
		reason := ""
		if len(match) > 2 {
			reason = match[2]
		}

		return &Suppression{
			RuleName:   ruleName,
			Reason:     reason,
			Line:       line,
			IsFileWide: false,
		}
	}

	return nil
}

func GetRuleName(id uint16) string {
	meta, ok := GetMetadata(id)
	if !ok {
		return ""
	}
	return meta.Name
}

func FilterSuppressedIssues(issues []Issue, suppressions []Suppression) []Issue {
	var filtered []Issue

	for _, issue := range issues {
		ruleName := GetRuleName(issue.ID)
		if ruleName == "" {
			filtered = append(filtered, issue)
			continue
		}

		suppressed := false

		for _, sup := range suppressions {
			if sup.RuleName != ruleName {
				continue
			}

			if sup.IsFileWide {
				suppressed = true
				break
			}

			// Inline suppression applies to the same line or the next line
			if sup.Line == issue.Pos.Line || sup.Line+1 == issue.Pos.Line {
				suppressed = true
				break
			}
		}

		if !suppressed {
			filtered = append(filtered, issue)
		}
	}

	return filtered
}

func CheckUnusedSuppressions(issues []Issue, suppressions []Suppression) []Issue {
	var warnings []Issue

	for _, sup := range suppressions {
		if sup.IsMisplaced {
			warnings = append(warnings, Issue{
				ID:       MisplacedFileWideIgnoreID,
				Pos:      token.Position{Line: sup.Line},
				Severity: SeverityWarn,
				ArgStr1:  sup.RuleName,
			})
			continue
		}

		used := false

		for _, issue := range issues {
			ruleName := GetRuleName(issue.ID)
			if ruleName != sup.RuleName {
				continue
			}

			if sup.IsFileWide {
				used = true
				break
			}

			// Inline suppression applies to the same line or the next line
			if sup.Line == issue.Pos.Line || sup.Line+1 == issue.Pos.Line {
				used = true
				break
			}
		}

		if !used {
			pos := token.Position{Line: sup.Line}
			if sup.IsFileWide {
				pos.Line = 1
			}

			warnings = append(warnings, Issue{
				ID:       UnusedSuppressionID,
				Pos:      pos,
				Severity: SeverityWarn,
				ArgStr1:  sup.RuleName,
			})
		}
	}

	return warnings
}
