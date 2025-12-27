package linter

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
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

func (l *Linter) processSingleFile(path string, r map[reflect.Type][]rules.Rule) ([]rules.Issue, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	return l.analyze(analysisParams{
		path:  path,
		src:   src,
		fset:  fset,
		rules: r,
		shouldStop: func(currentLocalCount int) bool {
			return l.MaxIssues > 0 && currentLocalCount >= l.MaxIssues
		},
	})
}

type analysisParams struct {
	path       string
	src        []byte
	rules      map[reflect.Type][]rules.Rule
	fset       *token.FileSet
	shouldStop func(int) bool
}

func (l *Linter) analyze(params analysisParams) ([]rules.Issue, error) {
	f, err := parser.ParseFile(params.fset, params.path, params.src, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	if l.Config.Linter.Use != nil && !*l.Config.Linter.Use {
		return nil, nil
	}

	issues := make([]rules.Issue, initialFileIssueCap, finalFileIssueCap)

	runner := rules.Runner{
		File:    f,
		Fset:    params.fset,
		Cfg:     l.Config,
		Autofix: l.Write || rules.CanAutoFix(l.Config),
		Unsafe:  l.Unsafe,
		Issues:  &issues,
		ShouldStop: func() bool {
			return params.shouldStop != nil && params.shouldStop(len(issues))
		},
	}

	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		runner.Node = n
		nodeType := reflect.TypeOf(n)
		if specificRules, found := params.rules[nodeType]; found {
			for _, rule := range specificRules {
				rule.Run(&runner, n)
			}
		}

		return true
	})

	if runner.Modified {
		var buf bytes.Buffer

		if err := format.Node(&buf, params.fset, f); err == nil {
			if err := os.WriteFile(params.path, buf.Bytes(), defaultFileMode); err != nil {
				return issues, fmt.Errorf("failed to write file %s: %w", params.path, err)
			}
		}
	}

	return issues, nil
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

		return l.processSingleFile(root, activeRules)
	}

	workers := runtime.GOMAXPROCS(0)

	channelBufferMultiplier := workers * 4
	paths := make(chan string, channelBufferMultiplier)
	results := make(chan []rules.Issue, workers)

	done := make(chan struct{})
	var stopOnce sync.Once

	var total int64
	var wg sync.WaitGroup
	wg.Add(workers)

	for range workers {
		go func() {
			defer wg.Done()

			for {
				fset := token.NewFileSet()

				select {
				case <-done:
					return
				case path, ok := <-paths:
					if !ok {
						return
					}

					src, err := os.ReadFile(path)
					if err != nil {
						continue
					}

					localIssues, err := l.analyze(analysisParams{
						path:  path,
						src:   src,
						fset:  fset,
						rules: activeRules,
						shouldStop: func(currentLocalCount int) bool {
							return l.MaxIssues > 0 && int(atomic.LoadInt64(&total)) >= l.MaxIssues
						},
					})
					if err != nil {
						if localIssues != nil {
							fmt.Fprintf(os.Stderr, "%v\n", err)
						}
						continue
					}

					if len(localIssues) == 0 {
						continue
					}

					if l.MaxIssues > 0 {
						atomic.AddInt64(&total, int64(len(localIssues)))
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

			select {
			case paths <- path:
			case <-done:
				return filepath.SkipAll
			}

			return nil
		})

		close(paths)
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
