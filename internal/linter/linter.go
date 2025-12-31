package linter

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/serenitysz/serenity/internal/rules"
)

type Linter struct {
	Write       bool
	Unsafe      bool
	MaxIssues   int // 0 = unlimited
	MaxFileSize int64
	Config      *rules.LinterOptions
}

func New(write, unsafe bool, config *rules.LinterOptions, maxIssues int, maxFileSize int64) *Linter {
	return &Linter{
		Write:       write,
		Unsafe:      unsafe,
		Config:      config,
		MaxIssues:   maxIssues,
		MaxFileSize: maxFileSize,
	}
}

const (
	defaultFileMode     = os.FileMode(0o644)
	initialFileIssueCap = 0
	finalFileIssueCap   = 32
)

type packageJob struct {
	dirPath string
	files   []string
}

type analysisParams struct {
	pkgFiles   []*ast.File
	pkgPaths   []string
	fset       *token.FileSet
	rules      map[reflect.Type][]rules.Rule
	shouldStop func(int) bool
}

func (l *Linter) analyzePackage(params analysisParams) ([]rules.Issue, error) {
	conf := types.Config{
		Importer: nil,
		Error:    func(err error) {},
	}

	info := &types.Info{
		Defs: make(map[*ast.Ident]types.Object),
		Uses: make(map[*ast.Ident]types.Object),
	}

	if len(params.pkgFiles) > 0 {
		conf.Check(params.pkgPaths[0], params.fset, params.pkgFiles, info)
	}

	mutatedObjects := make(map[types.Object]bool)

	if l.Config.Linter.Use == nil || *l.Config.Linter.Use {
		for _, f := range params.pkgFiles {
			ast.Inspect(f, func(n ast.Node) bool {
				switch t := n.(type) {
				case *ast.AssignStmt:
					for _, lhs := range t.Lhs {
						if id, ok := lhs.(*ast.Ident); ok {
							if obj := info.Uses[id]; obj != nil {
								mutatedObjects[obj] = true
							}
						}
					}
				case *ast.IncDecStmt:
					if id, ok := t.X.(*ast.Ident); ok {
						if obj := info.Uses[id]; obj != nil {
							mutatedObjects[obj] = true
						}
					}
				case *ast.UnaryExpr:
					if t.Op == token.AND {
						if id, ok := t.X.(*ast.Ident); ok {
							if obj := info.Uses[id]; obj != nil {
								mutatedObjects[obj] = true
							}
						}
					}
				}
				return true
			})
		}
	}

	estimatedIssues := len(params.pkgFiles) * 8
	allIssues := make([]rules.Issue, 0, estimatedIssues)

	for i, f := range params.pkgFiles {
		filePath := params.pkgPaths[i]

		issues := make([]rules.Issue, 0, finalFileIssueCap)

		runner := rules.Runner{
			File:           f,
			Fset:           params.fset,
			Cfg:            l.Config,
			Autofix:        l.Write || rules.CanAutoFix(l.Config),
			Unsafe:         l.Unsafe,
			Issues:         &issues,
			MutatedObjects: mutatedObjects,
			ShouldStop: func() bool {
				return params.shouldStop != nil && params.shouldStop(len(allIssues)+len(issues))
			},
			TypesInfo: info,
		}

		ast.Inspect(f, func(n ast.Node) bool {
			if n == nil {
				return true
			}

			nodeType := reflect.TypeOf(n)
			if specificRules, found := params.rules[nodeType]; found {
				for _, rule := range specificRules {
					rule.Run(&runner, n)
				}
			}

			return true
		})

		allIssues = append(allIssues, issues...)

		if runner.Modified {
			var buf bytes.Buffer
			if err := format.Node(&buf, params.fset, f); err == nil {
				if err := os.WriteFile(filePath, buf.Bytes(), defaultFileMode); err != nil {
					return allIssues, fmt.Errorf("failed to write file %s: %w", filePath, err)
				}
			}
		}
	}

	return allIssues, nil
}

func (l *Linter) ProcessPath(root string) ([]rules.Issue, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}

	activeRules := GetActiveRulesMap(l.Config)

	if !info.IsDir() {
		if l.MaxFileSize > 0 && info.Size() > l.MaxFileSize {
			return nil, nil
		}

		fset := token.NewFileSet()
		src, err := os.ReadFile(root)
		if err != nil {
			return nil, err
		}
		f, err := parser.ParseFile(fset, root, src, 0)
		if err != nil {
			return nil, err
		}

		return l.analyzePackage(analysisParams{
			pkgFiles: []*ast.File{f},
			pkgPaths: []string{root},
			fset:     fset,
			rules:    activeRules,
			shouldStop: func(current int) bool {
				return l.MaxIssues > 0 && current >= l.MaxIssues
			},
		})
	}

	workers := runtime.GOMAXPROCS(0)

	pkgJobs := make(chan packageJob, workers*2)
	results := make(chan []rules.Issue, workers)
	done := make(chan struct{})
	var stopOnce sync.Once

	var totalIssues int64
	var wg sync.WaitGroup
	wg.Add(workers)

	for range workers {
		go func() {
			defer wg.Done()

			for {
				select {
				case <-done:
					return
				case job, ok := <-pkgJobs:
					if !ok {
						return
					}

					fset := token.NewFileSet()
					var pkgFiles []*ast.File
					var pkgPaths []string

					for _, path := range job.files {
						src, err := os.ReadFile(path)
						if err != nil {
							continue
						}

						f, err := parser.ParseFile(fset, path, src, 0)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Parse error in %s: %v\n", path, err)
							continue
						}

						pkgFiles = append(pkgFiles, f)
						pkgPaths = append(pkgPaths, path)
					}

					if len(pkgFiles) == 0 {
						continue
					}

					localIssues, err := l.analyzePackage(analysisParams{
						pkgFiles: pkgFiles,
						pkgPaths: pkgPaths,
						fset:     fset,
						rules:    activeRules,
						shouldStop: func(current int) bool {
							return l.MaxIssues > 0 && int(atomic.LoadInt64(&totalIssues)) >= l.MaxIssues
						},
					})
					if err != nil {
						fmt.Fprintf(os.Stderr, "Package analysis error: %v\n", err)
						continue
					}

					if len(localIssues) == 0 {
						continue
					}

					if l.MaxIssues > 0 {
						atomic.AddInt64(&totalIssues, int64(len(localIssues)))
					}

					out := make([]rules.Issue, len(localIssues))
					copy(out, localIssues)

					select {
					case results <- out:
					case <-done:
						return
					}
				}
			}
		}()
	}

	go func() {
		pendingPackages := make(map[string][]string)

		filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			select {
			case <-done:
				return filepath.SkipAll
			default:
			}

			if d.IsDir() {
				if d.Name() == "vendor" || d.Name() == ".git" {
					return filepath.SkipDir
				}
				return nil
			}

			if !strings.HasSuffix(path, ".go") {
				return nil
			}

			if l.MaxFileSize > 0 {
				info, err := d.Info()
				if err == nil && info.Size() > l.MaxFileSize {
					return nil
				}
			}

			dir := filepath.Dir(path)
			pendingPackages[dir] = append(pendingPackages[dir], path)
			return nil
		})

		for dir, files := range pendingPackages {
			select {
			case pkgJobs <- packageJob{dirPath: dir, files: files}:
			case <-done:
				break
			}
		}

		close(pkgJobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	capHint := 128
	if l.MaxIssues > 0 && l.MaxIssues < capHint {
		capHint = l.MaxIssues
	}

	final := make([]rules.Issue, 0, capHint)

	for batch := range results {
		if l.MaxIssues == 0 {
			final = append(final, batch...)
			continue
		}

		remaining := l.MaxIssues - len(final)

		if remaining <= 0 {
			stopOnce.Do(func() { close(done) })
			break
		}

		if len(batch) > remaining {
			final = append(final, batch[:remaining]...)
			stopOnce.Do(func() {
				close(done)
			})

			break
		}

		final = append(final, batch...)
	}

	return final, nil
}
