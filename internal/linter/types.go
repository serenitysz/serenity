package linter

import (
	"go/ast"
	"go/token"
	"os"

	"github.com/serenitysz/serenity/internal/rules"
)

const (
	FINAL_FILE_ISSUE_CAP = 32
	DEFAULT_FILE_MODE    = os.FileMode(0o644)
)

type PackageJob struct {
	dirPath string
	files   []string
	inputs  []packageInput
}

type cachedBatch struct {
	data       []byte
	issueCount int
	issueStart int
	inputs     []packageInput
}

type issueBatch struct {
	issues []rules.Issue
	cached *cachedBatch
}

func (b issueBatch) count() int {
	if b.cached != nil {
		return b.cached.issueCount
	}

	return len(b.issues)
}

type AnalysisParams struct {
	pkgFiles     []*ast.File
	pkgPaths     []string
	fset         *token.FileSet
	maxIssues    int
	autofix      bool
	shouldStop   func(int) bool
	rules        *ActiveRules
	suppressions map[string][]rules.Suppression
}

type ActiveRules struct {
	File       []rules.Rule
	FuncDecl   []rules.Rule
	FuncLit    []rules.Rule
	TypeSpec   []rules.Rule
	ValueSpec  []rules.Rule
	Field      []rules.Rule
	ImportSpec []rules.Rule
	ReturnStmt []rules.Rule
	BasicLit   []rules.Rule
	CallExpr   []rules.Rule
	BlockStmt  []rules.Rule
	BinaryExpr []rules.Rule
	ForStmt    []rules.Rule
	RangeStmt  []rules.Rule
	AssignStmt []rules.Rule
	DeferStmt  []rules.Rule

	NeedsConstAnalysis bool
	HasAutofixRules    bool
}

func (a *ActiveRules) Run(runner *rules.Runner, node ast.Node) {
	switch n := node.(type) {
	case *ast.File:
		runRules(a.File, runner, n)
	case *ast.FuncDecl:
		runRules(a.FuncDecl, runner, n)
	case *ast.FuncLit:
		runRules(a.FuncLit, runner, n)
	case *ast.TypeSpec:
		runRules(a.TypeSpec, runner, n)
	case *ast.ValueSpec:
		runRules(a.ValueSpec, runner, n)
	case *ast.Field:
		runRules(a.Field, runner, n)
	case *ast.ImportSpec:
		runRules(a.ImportSpec, runner, n)
	case *ast.ReturnStmt:
		runRules(a.ReturnStmt, runner, n)
	case *ast.BasicLit:
		runRules(a.BasicLit, runner, n)
	case *ast.CallExpr:
		runRules(a.CallExpr, runner, n)
	case *ast.BlockStmt:
		runRules(a.BlockStmt, runner, n)
	case *ast.BinaryExpr:
		runRules(a.BinaryExpr, runner, n)
	case *ast.ForStmt:
		runRules(a.ForStmt, runner, n)
	case *ast.RangeStmt:
		runRules(a.RangeStmt, runner, n)
	case *ast.AssignStmt:
		runRules(a.AssignStmt, runner, n)
	case *ast.DeferStmt:
		runRules(a.DeferStmt, runner, n)
	}
}

func runRules(active []rules.Rule, runner *rules.Runner, node ast.Node) {
	for _, rule := range active {
		rule.Run(runner, node)

		if runner.ReachedMax() || (runner.ShouldStop != nil && runner.ShouldStop()) {
			return
		}
	}
}
