package linter

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/almeidazs/gowther/internal/rules"
)

type Linter struct {
	Write  bool
	Unsafe bool
	Fset   *token.FileSet
	Config *rules.Config
	Issues []rules.Issue
}

func New(write, unsafe bool, config *rules.Config) *Linter {
	return &Linter{
		Write:  write,
		Unsafe: unsafe,
		Fset:   token.NewFileSet(),
		Config: config,
		Issues: []rules.Issue{},
	}
}

func (l *Linter) ProcessPath(path string) ([]rules.Issue, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("error to read file info: %w", err)
	}

	var issues []rules.Issue
	if info.IsDir() {
		err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("error to read %v: %w", p, err)
			}

			if info.IsDir() {
				name := info.Name()
				if name == "vendor" || name == ".git" {
					return filepath.SkipDir
				}

				return nil
			}

			if strings.HasSuffix(p, ".go") && !strings.Contains(p, "vendor/") {
				issues, err = l.ProcessFile(p)
				if err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return issues, nil
}

func (l *Linter) ProcessFile(filename string) ([]rules.Issue, error) {
	src, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error to read file: %w", err)
	}

	f, err := parser.ParseFile(l.Fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse error in %s: %v", filename, err)
	}

	l.Issues = append(l.Issues, rules.CheckContextFirstParam(f, l.Fset)...)

	return l.Issues, nil
}
