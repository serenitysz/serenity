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

	for _, cg := range comments {
		if cg == nil {
			continue
		}

		for _, comment := range cg.List {
			if comment == nil {
				continue
			}

			line := fset.Position(comment.Pos()).Line

			sup := parseSuppression(comment.Text, line, pkgLine)
			if sup != nil {
				suppressions = append(suppressions, *sup)
			}
		}
	}

	return suppressions
}

func parseSuppression(text string, line int, pkgLine int) *Suppression {
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
	index := buildSuppressionIndex(suppressions)
	filtered := issues[:0]

	for _, issue := range issues {
		ruleName := GetRuleName(issue.ID)
		if ruleName == "" {
			filtered = append(filtered, issue)
			continue
		}

		if !index.matches(ruleName, issue.LineNumber()) {
			filtered = append(filtered, issue)
		}
	}

	return filtered
}

func CheckUnusedSuppressions(issues []Issue, suppressions []Suppression) []Issue {
	issueIndex := buildIssueIndex(issues)
	warnings := make([]Issue, 0, len(suppressions))

	for _, sup := range suppressions {
		if sup.IsMisplaced {
			warnings = append(warnings, Issue{
				ID:       MisplacedFileWideIgnoreID,
				Line:     uint32(sup.Line),
				Severity: SeverityWarn,
				ArgStr1:  sup.RuleName,
			})
			continue
		}

		used := issueIndex.matches(sup.RuleName, sup.Line)

		if !used {
			line := sup.Line
			if sup.IsFileWide {
				line = 1
			}

			warnings = append(warnings, Issue{
				ID:       UnusedSuppressionID,
				Line:     uint32(line),
				Severity: SeverityWarn,
				ArgStr1:  sup.RuleName,
			})
		}
	}

	return warnings
}

type suppressionIndex map[string]suppressionLines

type suppressionLines struct {
	fileWide bool
	inline   map[int]struct{}
}

func buildSuppressionIndex(suppressions []Suppression) suppressionIndex {
	index := make(suppressionIndex, len(suppressions))

	for _, sup := range suppressions {
		entry := index[sup.RuleName]
		if sup.IsFileWide {
			entry.fileWide = true
			index[sup.RuleName] = entry
			continue
		}

		if entry.inline == nil {
			entry.inline = make(map[int]struct{}, 2)
		}

		entry.inline[sup.Line] = struct{}{}
		entry.inline[sup.Line+1] = struct{}{}
		index[sup.RuleName] = entry
	}

	return index
}

func buildIssueIndex(issues []Issue) suppressionIndex {
	index := make(suppressionIndex, len(issues))

	for _, issue := range issues {
		ruleName := GetRuleName(issue.ID)
		if ruleName == "" {
			continue
		}

		entry := index[ruleName]
		if entry.inline == nil {
			entry.inline = make(map[int]struct{}, 1)
		}

		entry.inline[issue.LineNumber()] = struct{}{}
		index[ruleName] = entry
	}

	return index
}

func (s suppressionIndex) matches(ruleName string, line int) bool {
	entry, ok := s[ruleName]
	if !ok {
		return false
	}

	if entry.fileWide {
		return true
	}

	_, ok = entry.inline[line]

	return ok
}
