package linter

import (
	"go/ast"
	"go/token"
	"os"
	"reflect"

	"github.com/serenitysz/serenity/internal/rules"
)

const (
	FINAL_FILE_ISSUE_CAP = 32
	DEFAULT_FILE_MODE    = os.FileMode(0o644)
)

type PackageJob struct {
	dirPath string
	files   []string
}

type AnalysisParams struct {
	pkgFiles   []*ast.File
	pkgPaths   []string
	fset       *token.FileSet
	shouldStop func(int) bool
	rules      map[reflect.Type][]rules.Rule
}
