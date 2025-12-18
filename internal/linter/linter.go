package linter

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
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

func (l *Linter) ProcessPath(root string) ([]rules.Issue, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		if l.MaxFileSize > 0 && info.Size() > l.MaxFileSize {
			return nil, nil
		}

		return l.processSingleFile(root)
	}

	workers := runtime.GOMAXPROCS(0)

	paths := make(chan string, workers*4)
	results := make(chan []rules.Issue, workers)

	done := make(chan struct{})
	var stopOnce sync.Once

	var total int64
	var wg sync.WaitGroup
	wg.Add(workers)

	for range workers {
		go func() {
			defer wg.Done()

			fset := token.NewFileSet()
			local := make([]rules.Issue, 0, 32)

			for {
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

					f, err := parser.ParseFile(
						fset,
						path,
						src,
						parser.SkipObjectResolution,
					)
					if err != nil {
						continue
					}

					local = rules.CheckContextFirstParam(f, fset, local[:0])
					local = rules.CheckNoDotImports(f, fset, local)

					if len(local) == 0 {
						continue
					}

					if l.Write {
						for i := range local {
							if local[i].Fix != nil {
								local[i].Fix()
							}
						}

						var buf bytes.Buffer

						if err := format.Node(&buf, fset, f); err == nil {
							_ = os.WriteFile(path, buf.Bytes(), 0o644)
						}
					}

					out := make([]rules.Issue, len(local))
					copy(out, local)

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
				switch d.Name() {
				case "vendor", ".git":
					return filepath.SkipDir
				}
				return nil
			}

			if len(path) < 3 || path[len(path)-3:] != ".go" {
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

		cur := int(atomic.LoadInt64(&total))
		remaining := l.MaxIssues - cur

		if remaining <= 0 {
			stopOnce.Do(func() { close(done) })
			break
		}

		if len(batch) <= remaining {
			final = append(final, batch...)
			atomic.AddInt64(&total, int64(len(batch)))
		} else {
			final = append(final, batch[:remaining]...)
			atomic.StoreInt64(&total, int64(l.MaxIssues))
			stopOnce.Do(func() { close(done) })
			break
		}
	}

	return final, nil
}

func (l *Linter) processSingleFile(path string) ([]rules.Issue, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	issues := rules.CheckContextFirstParam(f, fset, nil)
	issues = rules.CheckNoDotImports(f, fset, issues)

	if l.Write && len(issues) > 0 {
		for i := range issues {
			if issues[i].Fix != nil {
				issues[i].Fix()
			}
		}

		var buf bytes.Buffer

		if err := format.Node(&buf, fset, f); err == nil {
			_ = os.WriteFile(path, buf.Bytes(), 0o644)
		}

	}

	return issues, nil
}
