package linter

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/serenitysz/serenity/internal/exception"
	"github.com/serenitysz/serenity/internal/rules"
)

func (l *Linter) ProcessPath(root string) ([]rules.Issue, error) {
	info, err := os.Stat(root)

	if err != nil {
		return nil, exception.InternalError("%v", err)
	}

	activeRules := GetActiveRulesMap(l.Config)

	if !info.IsDir() {
		if l.MaxFileSize > 0 && info.Size() > l.MaxFileSize {
			return nil, nil
		}

		fset := token.NewFileSet()
		src, err := os.ReadFile(root)

		if err != nil {
			return nil, exception.InternalError("%v", err)
		}

		f, err := parser.ParseFile(fset, root, src, 0)

		if err != nil {
			return nil, exception.InternalError("%v", err)
		}

		return l.Analyze(AnalysisParams{
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

	done := make(chan struct{})

	pkgJobs := make(chan PackageJob, workers*2)
	results := make(chan []rules.Issue, workers)

	var wg sync.WaitGroup
	var stopOnce sync.Once

	var totalIssues int64

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

					var pkgPaths []string
					var pkgFiles []*ast.File

					// TODO: Review this later
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

					localIssues, err := l.Analyze(AnalysisParams{
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
			case pkgJobs <- PackageJob{dirPath: dir, files: files}:
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
