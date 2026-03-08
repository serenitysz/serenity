package check

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/serenitysz/serenity/internal/render"
	"github.com/serenitysz/serenity/internal/rules"
)

const issueSeparator = "──────────────────"
const (
	frameContextLines = 1
	frameMaxWidth     = 96
	tabWidth          = 4
	maxHighlightWidth = 12
)

type issueRenderer struct {
	cwd         string
	out         io.Writer
	sourceCache map[string][]string
}

func newIssueRenderer(out io.Writer) issueRenderer {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = ""
	}

	return issueRenderer{
		cwd:         cwd,
		out:         out,
		sourceCache: make(map[string][]string, 8),
	}
}

func (r *issueRenderer) write(issue rules.Issue, msg string) {
	meta, ok := rules.GetMetadata(issue.ID)
	ruleName := "unknown-rule"
	if ok && meta.Name != "" {
		ruleName = meta.Name
	}

	var b strings.Builder
	b.Grow(len(msg) + len(ruleName) + 96)

	sevLabel, sevColor := severityPresentation(issue.Severity)

	b.WriteString(render.Paint(sevLabel, sevColor, false))
	b.WriteByte(' ')
	b.WriteString(render.Paint("•", render.Gray, false))
	b.WriteByte(' ')
	b.WriteString(render.Paint(ruleName, render.Gray, false))
	b.WriteString(render.Paint(fmt.Sprintf(" (%d)", issue.ID), render.Gray, false))

	if rules.IsFixable(issue.ID) {
		b.WriteByte(' ')
		b.WriteString(render.Paint("[fixable]", render.Green, false))
	}

	b.WriteByte('\n')
	b.WriteString(render.Paint(r.displayPath(issue.Filename()), render.Gray, false))
	b.WriteByte(':')
	b.WriteString(render.Paint(fmt.Sprintf("%d:%d", issue.LineNumber(), issue.ColumnNumber()), render.Gray, false))
	b.WriteByte('\n')
	b.WriteByte('\n')

	r.writeFrame(&b, issue, sevColor)

	b.WriteString(render.Paint("Hint:", render.Bold+render.Gray, false))
	b.WriteByte('\n')
	b.WriteString("  ")
	b.WriteString(msg)
	b.WriteString("\n\n")
	b.WriteString(render.Paint(issueSeparator, render.Gray, false))
	b.WriteByte('\n')

	_, _ = io.WriteString(r.out, b.String())
}

func (r *issueRenderer) writeFrame(b *strings.Builder, issue rules.Issue, sevColor string) {
	lines, ok := r.loadLines(issue.Filename())
	if !ok {
		return
	}

	lineNo := issue.LineNumber()
	if lineNo < 1 || lineNo > len(lines) {
		return
	}

	startLine := max(1, lineNo-frameContextLines)
	endLine := min(len(lines), lineNo+frameContextLines)
	gutterWidth := len(strconv.Itoa(endLine))

	targetRaw := strings.TrimRight(lines[lineNo-1], "\r")
	expandedTarget := expandTabs(targetRaw)
	startColumn, spanWidth := highlightColumns(issue, targetRaw)
	window := cropWindow(expandedTarget, startColumn, spanWidth)

	for current := startLine; current <= endLine; current++ {
		raw := strings.TrimRight(lines[current-1], "\r")
		display := cropExpandedLine(expandTabs(raw), window)

		marker := " "
		markerColor := render.Gray
		if current == lineNo {
			marker = ">"
			markerColor = sevColor
		}

		b.WriteString(render.Paint(marker, markerColor, false))
		b.WriteByte(' ')
		b.WriteString(render.Paint(fmt.Sprintf("%*d", gutterWidth, current), render.Gray, false))
		b.WriteByte(' ')
		b.WriteString(render.Paint("│", render.Gray, false))
		b.WriteByte(' ')
		b.WriteString(display)
		b.WriteByte('\n')

		if current == lineNo {
			caretColumn := window.highlightColumn
			caretWidth := window.highlightWidth
			if caretWidth < 1 {
				caretWidth = 1
			}

			b.WriteByte(' ')
			b.WriteByte(' ')
			b.WriteString(strings.Repeat(" ", gutterWidth))
			b.WriteByte(' ')
			b.WriteString(render.Paint("│", render.Gray, false))
			b.WriteByte(' ')
			b.WriteString(strings.Repeat(" ", max(0, caretColumn-1)))
			b.WriteString(render.Paint(highlightMarker(caretWidth), sevColor, false))
			b.WriteByte('\n')
		}
	}
	b.WriteByte('\n')
}

func (r issueRenderer) displayPath(path string) string {
	if path == "" || r.cwd == "" {
		if filepath.IsAbs(path) {
			return path
		}

		return filepath.ToSlash(filepath.Clean(path))
	}

	if !filepath.IsAbs(path) {
		return filepath.ToSlash(filepath.Clean(path))
	}

	rel, err := filepath.Rel(r.cwd, path)
	if err != nil {
		return path
	}

	if rel == "." {
		return path
	}

	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return path
	}

	return filepath.ToSlash(rel)
}

func severityPresentation(severity rules.Severity) (string, string) {
	switch severity {
	case rules.SeverityError:
		return "Error", render.Red
	case rules.SeverityInfo:
		return "Info", render.Blue
	case rules.SeverityWarn:
		return "Warn", render.Yellow
	default:
		return "Issue", render.Gray
	}
}

func (r *issueRenderer) loadLines(path string) ([]string, bool) {
	if path == "" {
		return nil, false
	}

	if r.sourceCache == nil {
		r.sourceCache = make(map[string][]string, 4)
	}

	if lines, ok := r.sourceCache[path]; ok {
		return lines, true
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	r.sourceCache[path] = lines

	return lines, true
}

type frameWindowState struct {
	startColumn     int
	endColumn       int
	highlightColumn int
	highlightWidth  int
}

func cropWindow(line string, startColumn, highlightWidth int) frameWindowState {
	runes := []rune(line)
	lineWidth := len(runes)

	if lineWidth == 0 {
		return frameWindowState{
			startColumn:     1,
			endColumn:       0,
			highlightColumn: 1,
			highlightWidth:  highlightWidth,
		}
	}

	if startColumn < 1 {
		startColumn = 1
	}
	if highlightWidth < 1 {
		highlightWidth = 1
	}

	endHighlight := startColumn + highlightWidth - 1
	if endHighlight > lineWidth {
		endHighlight = lineWidth
	}

	if lineWidth <= frameMaxWidth {
		return frameWindowState{
			startColumn:     1,
			endColumn:       lineWidth,
			highlightColumn: startColumn,
			highlightWidth:  max(1, endHighlight-startColumn+1),
		}
	}

	windowStart := max(1, startColumn-18)
	windowEnd := windowStart + frameMaxWidth - 1

	if windowEnd < endHighlight {
		windowEnd = endHighlight
		windowStart = windowEnd - frameMaxWidth + 1
	}

	if windowEnd > lineWidth {
		windowEnd = lineWidth
		windowStart = max(1, windowEnd-frameMaxWidth+1)
	}

	highlightColumn := startColumn - windowStart + 1
	if windowStart > 1 {
		highlightColumn++
	}

	return frameWindowState{
		startColumn:     windowStart,
		endColumn:       windowEnd,
		highlightColumn: highlightColumn,
		highlightWidth:  max(1, endHighlight-startColumn+1),
	}
}

func cropExpandedLine(line string, window frameWindowState) string {
	runes := []rune(line)
	if len(runes) == 0 {
		return ""
	}

	if len(runes) < window.startColumn {
		return line
	}

	start := clamp(window.startColumn, 1, len(runes)+1)
	end := clamp(window.endColumn, 0, len(runes))
	if end < start {
		end = start - 1
	}

	var b strings.Builder
	b.Grow((end - start + 1) + 2)

	if start > 1 {
		b.WriteRune('…')
	}

	if end >= start {
		b.WriteString(string(runes[start-1 : end]))
	}

	if end < len(runes) {
		b.WriteRune('…')
	}

	return b.String()
}

func highlightColumns(issue rules.Issue, rawLine string) (int, int) {
	switch issue.ID {
	case rules.MaxLineLengthID:
		startRaw := int(issue.ArgInt1) + 1
		if startRaw < 1 {
			startRaw = 1
		}

		endRaw := int(issue.ArgInt2) + 1
		if endRaw <= startRaw {
			endRaw = startRaw + 1
		}

		startDisplay := displayColumn(rawLine, startRaw)
		endDisplay := displayColumn(rawLine, endRaw)

		return startDisplay, max(1, endDisplay-startDisplay)

	default:
		startRaw := issue.ColumnNumber()
		if startRaw < 1 {
			startRaw = 1
		}

		return displayColumn(rawLine, startRaw), 1
	}
}

func displayColumn(rawLine string, rawColumn int) int {
	if rawColumn <= 1 {
		return 1
	}

	display := 1
	current := 1

	for len(rawLine) > 0 && current < rawColumn {
		r, size := utf8.DecodeRuneInString(rawLine)
		rawLine = rawLine[size:]
		current++

		if r == '\t' {
			display += tabWidth - ((display - 1) % tabWidth)
			continue
		}

		display++
	}

	return display
}

func expandTabs(rawLine string) string {
	var b strings.Builder
	b.Grow(len(rawLine) + 8)

	column := 1

	for len(rawLine) > 0 {
		r, size := utf8.DecodeRuneInString(rawLine)
		rawLine = rawLine[size:]

		if r == '\t' {
			spaces := tabWidth - ((column - 1) % tabWidth)
			b.WriteString(strings.Repeat(" ", spaces))
			column += spaces
			continue
		}

		b.WriteRune(r)
		column++
	}

	return b.String()
}

func highlightMarker(width int) string {
	if width <= 1 {
		return "^"
	}

	if width > maxHighlightWidth {
		return strings.Repeat("^", maxHighlightWidth) + "…"
	}

	return strings.Repeat("^", width)
}

func clamp(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}

	return value
}
