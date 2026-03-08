package linter

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/serenitysz/serenity/internal/exception"
	"github.com/serenitysz/serenity/internal/render"
	"github.com/serenitysz/serenity/internal/rules"
)

func (l *Linter) ProcessPath(root string) ([]rules.Issue, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, exception.InternalError("could not access %q: %w", root, err)
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
	results := make(chan issueBatch, workers)

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

					batch, err := l.processPackageJob(job, &totalIssues)
					if err != nil {
						render.Warnf("%s  %s", job.dirPath, exception.Message(err))
						continue
					}
					if batch.count() == 0 {
						continue
					}

					reachedLimit := false
					if l.MaxIssues > 0 {
						reachedLimit = int(atomic.AddInt64(&totalIssues, int64(batch.count()))) >= l.MaxIssues
					}

					if reachedLimit {
						select {
						case results <- batch:
							stopOnce.Do(func() { close(done) })
						case <-done:
						}
						return
					}

					select {
					case results <- batch:
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
	batches := make([]issueBatch, 0, workers)
	totalFinal := 0

	for batch := range results {
		if l.MaxIssues == 0 {
			if batch.count() == 0 {
				continue
			}

			batches = append(batches, batch)
			totalFinal += batch.count()
			continue
		}

		remaining := l.MaxIssues - len(final)
		if remaining <= 0 {
			stopOnce.Do(func() { close(done) })
			break
		}

		if len(batch.issues) > remaining {
			final = append(final, batch.issues[:remaining]...)
			stopOnce.Do(func() { close(done) })
			break
		}

		final = append(final, batch.issues...)
	}

	if l.MaxIssues == 0 {
		if len(batches) == 1 {
			if batches[0].cached != nil {
				final = make([]rules.Issue, batches[0].cached.issueCount)
				n, ok := decodeCachedIssuesInto(final, batches[0].cached)
				if !ok {
					return nil, exception.InternalError("could not decode cached issues")
				}
				return final[:n], nil
			}

			return batches[0].issues, nil
		}

		final = make([]rules.Issue, totalFinal)
		offset := 0
		for _, batch := range batches {
			if batch.cached != nil {
				n, ok := decodeCachedIssuesInto(final[offset:offset+batch.cached.issueCount], batch.cached)
				if !ok {
					return nil, exception.InternalError("could not decode cached issues")
				}
				offset += n
				continue
			}

			offset += copy(final[offset:], batch.issues)
		}
		final = final[:offset]
	}

	return final, nil
}

func (l *Linter) processFile(path string, size int64) ([]rules.Issue, error) {
	if l.MaxFileSize > 0 && size > l.MaxFileSize {
		return nil, nil
	}

	if l.Cache.enabledForRun() {
		inputs, err := probePackageInputs([]string{path})
		if err != nil {
			return nil, exception.InternalError("could not inspect %q: %w", path, err)
		}

		if cached, ok := l.Cache.load(inputs, l.MaxIssues); ok {
			if l.MaxIssues > 0 {
				return truncateIssues(cached, l.MaxIssues), nil
			}

			return cached, nil
		}

		if err := loadPackageSources(inputs); err != nil {
			return nil, exception.InternalError("could not read %q: %w", path, err)
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, inputs[0].Src, l.ParseMode)
		if err != nil {
			return nil, exception.InternalError("could not parse Go file %q: %w", path, err)
		}

		issues, err := l.analyzePackage([]*ast.File{file}, []string{path}, fset, map[string][]rules.Suppression{
			path: rules.ProcessSuppressions(file.Comments, fset, file.Decls, file.Package),
		}, 0, nil)
		if err != nil {
			return nil, err
		}

		issues, err = l.refreshCacheForInputs(inputs, issues)
		if err != nil {
			return nil, err
		}

		if l.MaxIssues > 0 {
			return truncateIssues(issues, l.MaxIssues), nil
		}

		return issues, nil
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, l.ParseMode)
	if err != nil {
		return nil, exception.InternalError("could not parse Go file %q: %w", path, err)
	}

	return l.analyzePackage([]*ast.File{file}, []string{path}, fset, map[string][]rules.Suppression{
		path: rules.ProcessSuppressions(file.Comments, fset, file.Decls, file.Package),
	}, l.MaxIssues, func(current int) bool {
		return l.MaxIssues > 0 && current >= l.MaxIssues
	})
}

func (l *Linter) processPackageJob(job PackageJob, totalIssues *int64) (issueBatch, error) {
	if l.MaxIssues > 0 && atomic.LoadInt64(totalIssues) >= int64(l.MaxIssues) {
		return issueBatch{}, nil
	}

	if l.Cache.enabledForRun() {
		return l.processCachedPackageJob(job, totalIssues)
	}

	pkgFiles, pkgPaths, fset, suppressions := l.parsePackage(job.files)
	if len(pkgFiles) == 0 {
		return issueBatch{}, nil
	}

	issues, err := l.analyzePackage(pkgFiles, pkgPaths, fset, suppressions, l.issueBudgetFromTotal(totalIssues), func(current int) bool {
		if l.MaxIssues <= 0 {
			return false
		}

		return int(atomic.LoadInt64(totalIssues))+current >= l.MaxIssues
	})
	if err != nil {
		return issueBatch{}, err
	}

	return issueBatch{issues: issues}, nil
}

func (l *Linter) processCachedPackageJob(job PackageJob, totalIssues *int64) (issueBatch, error) {
	inputs := job.inputs
	if len(inputs) == 0 {
		var err error
		inputs, err = probePackageInputs(job.files)
		if err != nil {
			return issueBatch{}, exception.InternalError("could not inspect package inputs in %q: %w", job.dirPath, err)
		}
	}

	if l.MaxIssues <= 0 {
		if cached, ok := l.Cache.loadRaw(inputs); ok {
			return issueBatch{cached: cached}, nil
		}
	} else if cached, ok := l.Cache.load(inputs, l.issueBudgetFromTotal(totalIssues)); ok {
		return issueBatch{issues: l.limitIssuesByTotal(cached, totalIssues)}, nil
	}

	if err := loadPackageSources(inputs); err != nil {
		return issueBatch{}, exception.InternalError("could not read package sources in %q: %w", job.dirPath, err)
	}

	pkgFiles, pkgPaths, fset, suppressions, complete := l.parsePackageInputs(inputs)
	if len(pkgFiles) == 0 {
		return issueBatch{}, nil
	}

	issues, err := l.analyzePackage(pkgFiles, pkgPaths, fset, suppressions, 0, nil)
	if err != nil {
		return issueBatch{}, err
	}

	if complete {
		issues, err = l.refreshCacheForInputs(inputs, issues)
		if err != nil {
			return issueBatch{}, err
		}
	}

	return issueBatch{issues: l.limitIssuesByTotal(issues, totalIssues)}, nil
}

func (l *Linter) analyzePackage(
	pkgFiles []*ast.File,
	pkgPaths []string,
	fset *token.FileSet,
	suppressions map[string][]rules.Suppression,
	maxIssues int,
	shouldStop func(int) bool,
) ([]rules.Issue, error) {
	return l.Analyze(AnalysisParams{
		pkgFiles:     pkgFiles,
		pkgPaths:     pkgPaths,
		fset:         fset,
		maxIssues:    maxIssues,
		autofix:      l.ActiveRules.HasAutofixRules && (l.Write || l.Config.ShouldAutofix()),
		rules:        l.ActiveRules,
		suppressions: suppressions,
		shouldStop:   shouldStop,
	})
}

func (l *Linter) analyzePackageReadonly(
	pkgFiles []*ast.File,
	pkgPaths []string,
	fset *token.FileSet,
	suppressions map[string][]rules.Suppression,
) ([]rules.Issue, error) {
	return l.Analyze(AnalysisParams{
		pkgFiles:     pkgFiles,
		pkgPaths:     pkgPaths,
		fset:         fset,
		maxIssues:    0,
		autofix:      false,
		rules:        l.ActiveRules,
		suppressions: suppressions,
	})
}

func (l *Linter) refreshCacheForInputs(inputs []packageInput, issues []rules.Issue) ([]rules.Issue, error) {
	if !l.Cache.enabledForRun() || len(inputs) == 0 {
		return issues, nil
	}

	if !l.Cache.mutating {
		_ = l.Cache.save(inputs, issues)
		return issues, nil
	}

	refreshedInputs, changed, err := reloadPackageInputs(inputs)
	if err != nil {
		return issues, exception.InternalError("could not refresh package inputs after applying fixes: %w", err)
	}

	if !changed {
		_ = l.Cache.save(refreshedInputs, issues)
		return issues, nil
	}

	pkgFiles, pkgPaths, fset, suppressions, err := l.parsePackageInputsStrict(refreshedInputs)
	if err != nil {
		return issues, err
	}

	finalIssues, err := l.analyzePackageReadonly(pkgFiles, pkgPaths, fset, suppressions)
	if err != nil {
		return issues, err
	}

	_ = l.Cache.save(refreshedInputs, finalIssues)

	return issues, nil
}

func (l *Linter) parsePackage(paths []string) ([]*ast.File, []string, *token.FileSet, map[string][]rules.Suppression) {
	fset := token.NewFileSet()
	pkgFiles := make([]*ast.File, 0, len(paths))
	pkgPaths := make([]string, 0, len(paths))
	suppressions := make(map[string][]rules.Suppression, len(paths))

	for _, path := range paths {
		file, err := parser.ParseFile(fset, path, nil, l.ParseMode)
		if err != nil {
			render.Warnf("%s  could not parse Go file: %v", path, err)
			continue
		}

		suppressions[path] = rules.ProcessSuppressions(file.Comments, fset, file.Decls, file.Package)
		pkgFiles = append(pkgFiles, file)
		pkgPaths = append(pkgPaths, path)
	}

	return pkgFiles, pkgPaths, fset, suppressions
}

func (l *Linter) parsePackageInputs(inputs []packageInput) ([]*ast.File, []string, *token.FileSet, map[string][]rules.Suppression, bool) {
	fset := token.NewFileSet()
	pkgFiles := make([]*ast.File, 0, len(inputs))
	pkgPaths := make([]string, 0, len(inputs))
	suppressions := make(map[string][]rules.Suppression, len(inputs))
	complete := true

	for _, input := range inputs {
		file, err := parser.ParseFile(fset, input.Path, input.Src, l.ParseMode)
		if err != nil {
			complete = false
			render.Warnf("%s  could not parse Go file: %v", input.Path, err)
			continue
		}

		suppressions[input.Path] = rules.ProcessSuppressions(file.Comments, fset, file.Decls, file.Package)
		pkgFiles = append(pkgFiles, file)
		pkgPaths = append(pkgPaths, input.Path)
	}

	return pkgFiles, pkgPaths, fset, suppressions, complete
}

func (l *Linter) parsePackageInputsStrict(inputs []packageInput) ([]*ast.File, []string, *token.FileSet, map[string][]rules.Suppression, error) {
	fset := token.NewFileSet()
	pkgFiles := make([]*ast.File, 0, len(inputs))
	pkgPaths := make([]string, 0, len(inputs))
	suppressions := make(map[string][]rules.Suppression, len(inputs))

	for _, input := range inputs {
		file, err := parser.ParseFile(fset, input.Path, input.Src, l.ParseMode)
		if err != nil {
			return nil, nil, nil, nil, exception.InternalError("applied fixes left %q invalid: %w", input.Path, err)
		}

		suppressions[input.Path] = rules.ProcessSuppressions(file.Comments, fset, file.Decls, file.Package)
		pkgFiles = append(pkgFiles, file)
		pkgPaths = append(pkgPaths, input.Path)
	}

	return pkgFiles, pkgPaths, fset, suppressions, nil
}

func (l *Linter) limitIssuesByTotal(issues []rules.Issue, totalIssues *int64) []rules.Issue {
	if l.MaxIssues <= 0 {
		return issues
	}

	return truncateIssues(issues, l.issueBudgetFromTotal(totalIssues))
}

func (l *Linter) issueBudgetFromTotal(totalIssues *int64) int {
	if l.MaxIssues <= 0 {
		return 0
	}

	remaining := l.MaxIssues - int(atomic.LoadInt64(totalIssues))
	if remaining <= 0 {
		return 0
	}

	return remaining
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
		inputs := make([]packageInput, 0, len(entries))
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

			if l.Cache.enabledForRun() {
				info, err := entry.Info()
				if err != nil {
					continue
				}
				if l.MaxFileSize > 0 && info.Size() > l.MaxFileSize {
					continue
				}

				inputs = append(inputs, packageProbeFromInfo(path, info))
				files = append(files, path)
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

		if len(inputs) > 1 {
			sort.Slice(inputs, func(i, j int) bool {
				return inputs[i].NormalizedPath < inputs[j].NormalizedPath
			})
		}

		if len(files) > 0 {
			if !enqueue(PackageJob{dirPath: dir, files: files, inputs: inputs}) {
				return nil
			}
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

func reloadPackageInputs(inputs []packageInput) ([]packageInput, bool, error) {
	paths := make([]string, len(inputs))
	for i := range inputs {
		paths[i] = inputs[i].Path
	}

	refreshed, err := loadPackageInputs(paths)
	if err != nil {
		return nil, false, err
	}

	return refreshed, packageInputsChanged(inputs, refreshed), nil
}

func packageInputsChanged(before, after []packageInput) bool {
	if len(before) != len(after) {
		return true
	}

	for i := range before {
		if before[i].NormalizedPath != after[i].NormalizedPath || before[i].Hash != after[i].Hash {
			return true
		}
	}

	return false
}
