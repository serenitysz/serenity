package linter

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
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

	if !info.IsDir() {
		return l.processFile(root, info.Size())
	}

	workers := l.Workers
	if workers < 1 {
		workers = 1
	}

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

					pkgFiles, pkgPaths, fset, suppressions := l.parsePackage(job.files)
					if len(pkgFiles) == 0 {
						continue
					}

					localIssues, err := l.Analyze(AnalysisParams{
						pkgFiles:     pkgFiles,
						pkgPaths:     pkgPaths,
						fset:         fset,
						rules:        l.ActiveRules,
						suppressions: suppressions,
						shouldStop: func(current int) bool {
							if l.MaxIssues <= 0 {
								return false
							}

							return int(atomic.LoadInt64(&totalIssues))+current >= l.MaxIssues
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
						if int(atomic.AddInt64(&totalIssues, int64(len(localIssues)))) >= l.MaxIssues {
							stopOnce.Do(func() { close(done) })
						}
					}

					select {
					case results <- localIssues:
					case <-done:
						return
					}
				}
			}
		}()
	}

	go func() {
		defer close(pkgJobs)
		_ = l.walkPackages(root, done, func(job PackageJob) bool {
			select {
			case pkgJobs <- job:
				return true
			case <-done:
				return false
			}
		})
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
			stopOnce.Do(func() { close(done) })
			break
		}

		final = append(final, batch...)
	}

	return final, nil
}

func (l *Linter) processFile(path string, size int64) ([]rules.Issue, error) {
	if l.MaxFileSize > 0 && size > l.MaxFileSize {
		return nil, nil
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, l.ParseMode)
	if err != nil {
		return nil, exception.InternalError("%v", err)
	}

	return l.Analyze(AnalysisParams{
		pkgFiles: []*ast.File{file},
		pkgPaths: []string{path},
		fset:     fset,
		rules:    l.ActiveRules,
		suppressions: map[string][]rules.Suppression{
			path: rules.ProcessSuppressions(file.Comments, fset, file.Decls, file.Package),
		},
		shouldStop: func(current int) bool {
			return l.MaxIssues > 0 && current >= l.MaxIssues
		},
	})
}

func (l *Linter) parsePackage(paths []string) ([]*ast.File, []string, *token.FileSet, map[string][]rules.Suppression) {
	fset := token.NewFileSet()
	pkgFiles := make([]*ast.File, 0, len(paths))
	pkgPaths := make([]string, 0, len(paths))
	suppressions := make(map[string][]rules.Suppression, len(paths))

	for _, path := range paths {
		file, err := parser.ParseFile(fset, path, nil, l.ParseMode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Parse error in %s: %v\n", path, err)
			continue
		}

		suppressions[path] = rules.ProcessSuppressions(file.Comments, fset, file.Decls, file.Package)
		pkgFiles = append(pkgFiles, file)
		pkgPaths = append(pkgPaths, path)
	}

	return pkgFiles, pkgPaths, fset, suppressions
}

func (l *Linter) walkPackages(root string, done <-chan struct{}, enqueue func(PackageJob) bool) error {
	var walk func(string) error

	walk = func(dir string) error {
		select {
		case <-done:
			return nil
		default:
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil
		}

		files := make([]string, 0, len(entries))
		dirs := make([]string, 0, len(entries))

		for _, entry := range entries {
			select {
			case <-done:
				return nil
			default:
			}

			name := entry.Name()
			path := filepath.Join(dir, name)

			if entry.IsDir() {
				if name == "vendor" || name == ".git" {
					continue
				}

				dirs = append(dirs, path)
				continue
			}

			if !strings.HasSuffix(name, ".go") {
				continue
			}

			if l.MaxFileSize > 0 {
				info, err := entry.Info()
				if err == nil && info.Size() > l.MaxFileSize {
					continue
				}
			}

			files = append(files, path)
		}

		if len(files) > 0 && !enqueue(PackageJob{dirPath: dir, files: files}) {
			return nil
		}

		for _, subdir := range dirs {
			if err := walk(subdir); err != nil {
				return err
			}
		}

		return nil
	}

	return walk(root)
}
