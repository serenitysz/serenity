package linter

import (
	"bytes"
	"go/ast"
	"go/format"
	"os"

	"github.com/serenitysz/serenity/internal/exception"
	"github.com/serenitysz/serenity/internal/rules"
)

type visitFrame struct {
	prevFunc      *rules.FunctionContext
	prevLoopDepth int
	switchedFunc  bool
	enteredLoop   bool
}

func (l *Linter) Analyze(params AnalysisParams) ([]rules.Issue, error) {
	constCandidates := l.buildConstCandidates(params)

	estimatedIssues := len(params.pkgFiles) * 8
	allIssues := make([]rules.Issue, 0, estimatedIssues)

	for i, file := range params.pkgFiles {
		filePath := params.pkgPaths[i]
		issues := make([]rules.Issue, 0, FINAL_FILE_ISSUE_CAP)
		suppressions := params.suppressions[filePath]

		runner := rules.Runner{
			File:            file,
			Fset:            params.fset,
			Cfg:             l.Config,
			Unsafe:          l.Unsafe,
			Issues:          &issues,
			IssuesCount:     new(uint16),
			ConstCandidates: constCandidates,
			Autofix:         l.Write || l.Config.ShouldAutofix(),
			ShouldStop: func() bool {
				return params.shouldStop != nil && params.shouldStop(len(allIssues)+len(issues))
			},
			Suppressions: suppressions,
			MaxIssues:    params.maxIssues,
		}

		l.runFile(&runner, file, params.rules)

		unusedWarnings := rules.CheckUnusedSuppressions(filePath, issues, suppressions)

		issues = rules.FilterSuppressedIssues(issues, suppressions)
		issues = append(issues, unusedWarnings...)
		allIssues = append(allIssues, issues...)

		if runner.Modified {
			var buf bytes.Buffer

			if err := format.Node(&buf, params.fset, file); err == nil {
				if err := os.WriteFile(filePath, buf.Bytes(), DEFAULT_FILE_MODE); err != nil {
					return allIssues, exception.InternalError("failed to write file %s: %w", filePath, err)
				}
			}
		}
	}

	return allIssues, nil
}

func (l *Linter) runFile(runner *rules.Runner, file *ast.File, active *ActiveRules) {
	if active == nil {
		return
	}

	stack := make([]visitFrame, 0, 64)

	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			if len(stack) == 0 {
				return true
			}

			frame := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			restoreTraversalState(runner, frame)

			return true
		}

		frame := applyTraversalState(runner, n)
		stack = append(stack, frame)

		active.Run(runner, n)

		if runner.ReachedMax() || (runner.ShouldStop != nil && runner.ShouldStop()) {
			stack = stack[:len(stack)-1]
			restoreTraversalState(runner, frame)

			return false
		}

		return true
	})
}

func applyTraversalState(runner *rules.Runner, node ast.Node) visitFrame {
	frame := visitFrame{}

	switch n := node.(type) {
	case *ast.FuncDecl:
		frame.prevFunc = runner.CurrentFunc
		frame.prevLoopDepth = runner.LoopDepth
		frame.switchedFunc = true
		runner.CurrentFunc = functionContextForDecl(n)
		runner.LoopDepth = 0
	case *ast.FuncLit:
		frame.prevFunc = runner.CurrentFunc
		frame.prevLoopDepth = runner.LoopDepth
		frame.switchedFunc = true
		runner.CurrentFunc = functionContextForLit(n)
		runner.LoopDepth = 0
	case *ast.ForStmt, *ast.RangeStmt:
		frame.enteredLoop = true
		runner.LoopDepth++
	}

	return frame
}

func restoreTraversalState(runner *rules.Runner, frame visitFrame) {
	if frame.enteredLoop {
		runner.LoopDepth--
	}

	if frame.switchedFunc {
		runner.CurrentFunc = frame.prevFunc
		runner.LoopDepth = frame.prevLoopDepth
	}
}

func functionContextForDecl(fn *ast.FuncDecl) *rules.FunctionContext {
	name := "anonymous"

	if fn.Name != nil && fn.Name.Name != "" {
		name = fn.Name.Name
	}

	return &rules.FunctionContext{
		Name:            name,
		HasNamedResults: hasNamedResults(fn.Type),
	}
}

func functionContextForLit(fn *ast.FuncLit) *rules.FunctionContext {
	return &rules.FunctionContext{
		Name:            "anonymous",
		HasNamedResults: hasNamedResults(fn.Type),
	}
}

func hasNamedResults(fnType *ast.FuncType) bool {
	if fnType == nil || fnType.Results == nil {
		return false
	}

	for _, field := range fnType.Results.List {
		if len(field.Names) > 0 {
			return true
		}
	}

	return false
}
