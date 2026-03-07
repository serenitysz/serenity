package linter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/serenitysz/serenity/internal/config"
	"github.com/serenitysz/serenity/internal/rules"
	"github.com/serenitysz/serenity/internal/utils"
)

type benchmarkRuleMode string

const (
	benchmarkRuleModeStrict        benchmarkRuleMode = "strict"
	benchmarkRuleModeStrictNoConst benchmarkRuleMode = "strict-no-const"
	benchmarkRuleModeConstOnly     benchmarkRuleMode = "const-only"
)

type benchmarkCorpusSpec struct {
	name         string
	packages     int
	filesPerPkg  int
	funcsPerFile int
	lineLength   int
}

type benchmarkCorpus struct {
	root           string
	singleFile     string
	packageFiles   [][]string
	totalPackages  int
	totalFiles     int
	totalFuncs     int
	totalBytes     int64
	singleFileSize int64
}

type processBenchmarkScenario struct {
	name      string
	target    string
	workers   int
	cache     bool
	prewarm   bool
	maxIssues int
	ruleMode  benchmarkRuleMode
}

type analyzeBenchmarkScenario struct {
	name      string
	ruleMode  benchmarkRuleMode
	maxIssues int
}

type analyzeFixture struct {
	linter    *Linter
	params    AnalysisParams
	fileCount int
	byteCount int64
}

type cacheFixture struct {
	store        *cacheStore
	probedInputs []packageInput
	loadedInputs []packageInput
	issues       []rules.Issue
	data         []byte
	header       cacheHeader
	byteCount    int64
}

var (
	benchmarkSmallCorpus = benchmarkCorpusSpec{
		name:         "small",
		packages:     4,
		filesPerPkg:  4,
		funcsPerFile: 4,
		lineLength:   96,
	}
	benchmarkMediumCorpus = benchmarkCorpusSpec{
		name:         "medium",
		packages:     12,
		filesPerPkg:  8,
		funcsPerFile: 6,
		lineLength:   120,
	}
	benchmarkLargeCorpus = benchmarkCorpusSpec{
		name:         "large",
		packages:     24,
		filesPerPkg:  12,
		funcsPerFile: 10,
		lineLength:   160,
	}
)

func BenchmarkProcessPathSequential(b *testing.B) {
	corpus := writeBenchmarkCorpus(b, benchmarkMediumCorpus)
	runProcessPathScenario(b, corpus, processBenchmarkScenario{
		name:     "repo/cold/sequential",
		target:   "repo",
		workers:  1,
		cache:    false,
		prewarm:  false,
		ruleMode: benchmarkRuleModeStrict,
	})
}

func BenchmarkProcessPathParallel(b *testing.B) {
	corpus := writeBenchmarkCorpus(b, benchmarkMediumCorpus)
	runProcessPathScenario(b, corpus, processBenchmarkScenario{
		name:     "repo/cold/parallel",
		target:   "repo",
		workers:  runtime.GOMAXPROCS(0),
		cache:    false,
		prewarm:  false,
		ruleMode: benchmarkRuleModeStrict,
	})
}

func BenchmarkProcessPathSequentialWarmCache(b *testing.B) {
	corpus := writeBenchmarkCorpus(b, benchmarkMediumCorpus)
	runProcessPathScenario(b, corpus, processBenchmarkScenario{
		name:     "repo/warm/sequential",
		target:   "repo",
		workers:  1,
		cache:    true,
		prewarm:  true,
		ruleMode: benchmarkRuleModeStrict,
	})
}

func BenchmarkProcessPathParallelWarmCache(b *testing.B) {
	corpus := writeBenchmarkCorpus(b, benchmarkMediumCorpus)
	runProcessPathScenario(b, corpus, processBenchmarkScenario{
		name:     "repo/warm/parallel",
		target:   "repo",
		workers:  runtime.GOMAXPROCS(0),
		cache:    true,
		prewarm:  true,
		ruleMode: benchmarkRuleModeStrict,
	})
}

func BenchmarkProcessPathMatrix(b *testing.B) {
	corpora := []benchmarkCorpusSpec{
		benchmarkSmallCorpus,
		benchmarkMediumCorpus,
		benchmarkLargeCorpus,
	}
	scenarios := []processBenchmarkScenario{
		{
			name:     "repo/cold/sequential",
			target:   "repo",
			workers:  1,
			cache:    false,
			prewarm:  false,
			ruleMode: benchmarkRuleModeStrict,
		},
		{
			name:     "repo/cold/parallel",
			target:   "repo",
			workers:  runtime.GOMAXPROCS(0),
			cache:    false,
			prewarm:  false,
			ruleMode: benchmarkRuleModeStrict,
		},
		{
			name:     "repo/warm/sequential",
			target:   "repo",
			workers:  1,
			cache:    true,
			prewarm:  true,
			ruleMode: benchmarkRuleModeStrict,
		},
		{
			name:     "repo/warm/parallel",
			target:   "repo",
			workers:  runtime.GOMAXPROCS(0),
			cache:    true,
			prewarm:  true,
			ruleMode: benchmarkRuleModeStrict,
		},
		{
			name:      "repo/warm/parallel/max-1",
			target:    "repo",
			workers:   runtime.GOMAXPROCS(0),
			cache:     true,
			prewarm:   true,
			maxIssues: 1,
			ruleMode:  benchmarkRuleModeStrict,
		},
		{
			name:      "repo/warm/parallel/max-32",
			target:    "repo",
			workers:   runtime.GOMAXPROCS(0),
			cache:     true,
			prewarm:   true,
			maxIssues: 32,
			ruleMode:  benchmarkRuleModeStrict,
		},
		{
			name:     "file/cold/sequential",
			target:   "file",
			workers:  1,
			cache:    false,
			prewarm:  false,
			ruleMode: benchmarkRuleModeStrict,
		},
		{
			name:     "file/warm/sequential",
			target:   "file",
			workers:  1,
			cache:    true,
			prewarm:  true,
			ruleMode: benchmarkRuleModeStrict,
		},
	}

	for _, spec := range corpora {
		spec := spec
		b.Run(spec.name, func(b *testing.B) {
			corpus := writeBenchmarkCorpus(b, spec)
			for _, scenario := range scenarios {
				scenario := scenario
				b.Run(scenario.name, func(b *testing.B) {
					runProcessPathScenario(b, corpus, scenario)
				})
			}
		})
	}
}

func BenchmarkAnalyzeMatrix(b *testing.B) {
	spec := benchmarkCorpusSpec{
		name:         "analyze-heavy",
		packages:     1,
		filesPerPkg:  16,
		funcsPerFile: 12,
		lineLength:   160,
	}
	scenarios := []analyzeBenchmarkScenario{
		{name: "strict/unlimited", ruleMode: benchmarkRuleModeStrict},
		{name: "strict/max-32", ruleMode: benchmarkRuleModeStrict, maxIssues: 32},
		{name: "strict-no-const/unlimited", ruleMode: benchmarkRuleModeStrictNoConst},
		{name: "const-only/unlimited", ruleMode: benchmarkRuleModeConstOnly},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		b.Run(scenario.name, func(b *testing.B) {
			fixture := prepareAnalyzeFixture(b, spec, scenario)
			runAnalyzeBenchmark(b, fixture, scenario.maxIssues)
		})
	}
}

func BenchmarkCacheMatrix(b *testing.B) {
	fixture := prepareCacheFixture(b, benchmarkCorpusSpec{
		name:         "cache-heavy",
		packages:     1,
		filesPerPkg:  16,
		funcsPerFile: 12,
		lineLength:   160,
	})

	b.Run("encode", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(fixture.byteCount)
		b.ResetTimer()

		var payload []byte
		for range b.N {
			payload = encodeCache(fixture.loadedInputs, fixture.issues, fixture.store.configHash)
		}

		if len(payload) == 0 {
			b.Fatal("expected cache payload")
		}

		b.StopTimer()
		reportBenchmarkMetrics(b, fixture.byteCount, len(fixture.loadedInputs), 1, len(fixture.issues))
	})

	b.Run("validate-header", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(fixture.data)))
		b.ResetTimer()

		var header cacheHeader
		var err error
		for range b.N {
			header, err = validateCacheHeader(fixture.data, clonePackageInputs(fixture.probedInputs), fixture.store.configHash)
			if err != nil {
				b.Fatalf("validateCacheHeader failed: %v", err)
			}
		}

		if header.IssueCount == 0 {
			b.Fatal("expected cached issues")
		}

		b.StopTimer()
		reportBenchmarkMetrics(b, fixture.byteCount, len(fixture.probedInputs), 1, header.IssueCount)
	})

	b.Run("load-hot/unlimited", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(fixture.data)))
		b.ResetTimer()

		var issues []rules.Issue
		for range b.N {
			var ok bool
			issues, ok = fixture.store.load(clonePackageInputs(fixture.probedInputs), 0)
			if !ok {
				b.Fatal("expected cache hit")
			}
		}

		if len(issues) == 0 {
			b.Fatal("expected issues from cache")
		}

		b.StopTimer()
		reportBenchmarkMetrics(b, fixture.byteCount, len(fixture.probedInputs), 1, len(issues))
	})

	b.Run("load-hot/limit-32", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(fixture.data)))
		b.ResetTimer()

		var issues []rules.Issue
		for range b.N {
			var ok bool
			issues, ok = fixture.store.load(clonePackageInputs(fixture.probedInputs), 32)
			if !ok {
				b.Fatal("expected cache hit")
			}
		}

		if len(issues) != 32 {
			b.Fatalf("expected 32 cached issues, got %d", len(issues))
		}

		b.StopTimer()
		reportBenchmarkMetrics(b, fixture.byteCount, len(fixture.probedInputs), 1, len(issues))
	})

	b.Run("decode/unlimited", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(fixture.data)))
		b.ResetTimer()

		var issues []rules.Issue
		for range b.N {
			var ok bool
			issues, ok = decodeCachedIssues(fixture.data, fixture.header, fixture.probedInputs, 0)
			if !ok {
				b.Fatal("expected decode success")
			}
		}

		if len(issues) == 0 {
			b.Fatal("expected decoded issues")
		}

		b.StopTimer()
		reportBenchmarkMetrics(b, fixture.byteCount, len(fixture.probedInputs), 1, len(issues))
	})

	b.Run("decode/limit-32", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(fixture.data)))
		b.ResetTimer()

		var issues []rules.Issue
		for range b.N {
			var ok bool
			issues, ok = decodeCachedIssues(fixture.data, fixture.header, fixture.probedInputs, 32)
			if !ok {
				b.Fatal("expected decode success")
			}
		}

		if len(issues) != 32 {
			b.Fatalf("expected 32 decoded issues, got %d", len(issues))
		}

		b.StopTimer()
		reportBenchmarkMetrics(b, fixture.byteCount, len(fixture.probedInputs), 1, len(issues))
	})
}

func BenchmarkWalkPackagesMatrix(b *testing.B) {
	corpora := []benchmarkCorpusSpec{
		benchmarkSmallCorpus,
		benchmarkMediumCorpus,
		benchmarkLargeCorpus,
	}
	for _, spec := range corpora {
		spec := spec
		b.Run(spec.name, func(b *testing.B) {
			corpus := writeBenchmarkCorpus(b, spec)
			for _, cacheEnabled := range []bool{false, true} {
				cacheEnabled := cacheEnabled
				name := "cache-off"
				if cacheEnabled {
					name = "cache-on"
				}

				b.Run(name, func(b *testing.B) {
					if cacheEnabled {
						b.Setenv("SERENITY_CACHE_DIR", filepath.Join(b.TempDir(), "cache"))
					}

					cfg := newBenchmarkConfig(benchmarkRuleModeStrict, runtime.GOMAXPROCS(0), cacheEnabled)
					l := New(false, false, cfg, 0, 0)

					b.ReportAllocs()
					b.SetBytes(corpus.totalBytes)
					b.ResetTimer()

					var jobCount int
					var fileCount int

					for range b.N {
						jobCount = 0
						fileCount = 0

						if err := l.walkPackages(corpus.root, nil, func(job PackageJob) bool {
							jobCount++
							fileCount += len(job.files)
							return true
						}); err != nil {
							b.Fatalf("walkPackages failed: %v", err)
						}
					}

					b.StopTimer()
					reportBenchmarkMetrics(b, corpus.totalBytes, fileCount, jobCount, 0)
				})
			}
		})
	}
}

func runProcessPathScenario(b *testing.B, corpus benchmarkCorpus, scenario processBenchmarkScenario) {
	b.Helper()
	b.ReportAllocs()

	if scenario.cache {
		b.Setenv("SERENITY_CACHE_DIR", filepath.Join(b.TempDir(), "cache"))
	}

	cfg := newBenchmarkConfig(scenario.ruleMode, scenario.workers, scenario.cache)
	l := New(false, false, cfg, scenario.maxIssues, 0)

	targetPath, fileCount, packageCount, byteCount := corpus.target(scenario.target)
	expectedIssues := -1

	if scenario.prewarm {
		issues, err := l.ProcessPath(targetPath)
		if err != nil {
			b.Fatalf("benchmark warm-up failed: %v", err)
		}
		if len(issues) == 0 {
			b.Fatal("expected issues from benchmark fixture during warm-up")
		}
		expectedIssues = len(issues)
	}

	b.SetBytes(byteCount)
	b.ResetTimer()

	for range b.N {
		issues, err := l.ProcessPath(targetPath)
		if err != nil {
			b.Fatalf("ProcessPath failed: %v", err)
		}
		if len(issues) == 0 {
			b.Fatal("expected issues from benchmark fixture")
		}

		if expectedIssues < 0 {
			expectedIssues = len(issues)
			continue
		}
		if len(issues) != expectedIssues {
			b.Fatalf("issue count changed during benchmark: expected %d, got %d", expectedIssues, len(issues))
		}
	}

	b.StopTimer()
	reportBenchmarkMetrics(b, byteCount, fileCount, packageCount, expectedIssues)
}

func prepareAnalyzeFixture(b *testing.B, spec benchmarkCorpusSpec, scenario analyzeBenchmarkScenario) analyzeFixture {
	b.Helper()

	corpus := writeBenchmarkCorpus(b, spec)
	cfg := newBenchmarkConfig(scenario.ruleMode, 1, false)
	l := New(false, false, cfg, scenario.maxIssues, 0)

	inputs, err := loadPackageInputs(corpus.packageFiles[0])
	if err != nil {
		b.Fatalf("loadPackageInputs failed: %v", err)
	}

	pkgFiles, pkgPaths, fset, suppressions, complete := l.parsePackageInputs(inputs)
	if !complete {
		b.Fatal("expected benchmark package to parse without errors")
	}

	params := AnalysisParams{
		pkgFiles:     pkgFiles,
		pkgPaths:     pkgPaths,
		fset:         fset,
		maxIssues:    scenario.maxIssues,
		rules:        l.ActiveRules,
		suppressions: suppressions,
	}
	if scenario.maxIssues > 0 {
		params.shouldStop = func(current int) bool {
			return current >= scenario.maxIssues
		}
	}

	byteCount := int64(0)
	for _, input := range inputs {
		byteCount += int64(len(input.Src))
	}

	return analyzeFixture{
		linter:    l,
		params:    params,
		fileCount: len(pkgFiles),
		byteCount: byteCount,
	}
}

func runAnalyzeBenchmark(b *testing.B, fixture analyzeFixture, maxIssues int) {
	b.Helper()
	b.ReportAllocs()
	b.SetBytes(fixture.byteCount)
	b.ResetTimer()

	expectedIssues := -1
	for range b.N {
		issues, err := fixture.linter.Analyze(fixture.params)
		if err != nil {
			b.Fatalf("Analyze failed: %v", err)
		}

		if expectedIssues < 0 {
			expectedIssues = len(issues)
		} else if len(issues) != expectedIssues {
			b.Fatalf("issue count changed during benchmark: expected %d, got %d", expectedIssues, len(issues))
		}

		if maxIssues == 0 && len(issues) == 0 {
			b.Fatal("expected Analyze to report issues")
		}
	}

	b.StopTimer()
	reportBenchmarkMetrics(b, fixture.byteCount, fixture.fileCount, 1, expectedIssues)
}

func prepareCacheFixture(b *testing.B, spec benchmarkCorpusSpec) cacheFixture {
	b.Helper()

	corpus := writeBenchmarkCorpus(b, spec)
	b.Setenv("SERENITY_CACHE_DIR", filepath.Join(b.TempDir(), "cache"))

	cfg := newBenchmarkConfig(benchmarkRuleModeStrict, 1, true)
	l := New(false, false, cfg, 0, 0)

	loadedInputs, err := loadPackageInputs(corpus.packageFiles[0])
	if err != nil {
		b.Fatalf("loadPackageInputs failed: %v", err)
	}

	pkgFiles, pkgPaths, fset, suppressions, complete := l.parsePackageInputs(loadedInputs)
	if !complete {
		b.Fatal("expected cache benchmark package to parse without errors")
	}

	issues, err := l.analyzePackage(pkgFiles, pkgPaths, fset, suppressions, 0, nil)
	if err != nil {
		b.Fatalf("analyzePackage failed: %v", err)
	}
	if len(issues) == 0 {
		b.Fatal("expected cache fixture to produce issues")
	}

	if err := l.Cache.save(loadedInputs, issues); err != nil {
		b.Fatalf("cache save failed: %v", err)
	}

	probedInputs, err := probePackageInputs(corpus.packageFiles[0])
	if err != nil {
		b.Fatalf("probePackageInputs failed: %v", err)
	}

	entryPath, err := l.Cache.entryPath(probedInputs)
	if err != nil {
		b.Fatalf("entryPath failed: %v", err)
	}

	data, err := os.ReadFile(entryPath)
	if err != nil {
		b.Fatalf("read cache entry failed: %v", err)
	}

	header, err := validateCacheHeader(data, clonePackageInputs(probedInputs), l.Cache.configHash)
	if err != nil {
		b.Fatalf("validateCacheHeader failed: %v", err)
	}

	byteCount := int64(0)
	for _, input := range loadedInputs {
		byteCount += int64(len(input.Src))
	}

	return cacheFixture{
		store:        l.Cache,
		probedInputs: probedInputs,
		loadedInputs: loadedInputs,
		issues:       issues,
		data:         data,
		header:       header,
		byteCount:    byteCount,
	}
}

func clonePackageInputs(inputs []packageInput) []packageInput {
	cloned := make([]packageInput, len(inputs))
	copy(cloned, inputs)
	return cloned
}

func reportBenchmarkMetrics(b *testing.B, byteCount int64, fileCount int, packageCount int, issueCount int) {
	b.Helper()

	if byteCount > 0 {
		b.ReportMetric(float64(byteCount), "source-bytes/op")
	}
	if fileCount > 0 {
		b.ReportMetric(float64(fileCount), "files/op")
	}
	if packageCount > 0 {
		b.ReportMetric(float64(packageCount), "pkgs/op")
	}
	if issueCount >= 0 {
		b.ReportMetric(float64(issueCount), "issues/op")
	}
}

func newBenchmarkConfig(mode benchmarkRuleMode, workers int, cacheEnabled bool) *rules.LinterOptions {
	cfg := config.GenStrictDefaultConfig(utils.Ptr(false))
	cfg.Linter.Issues.Max = 0
	cfg.Performance.Threads = utils.Ptr(workers)
	cfg.Performance.Caching = utils.Ptr(cacheEnabled)

	switch mode {
	case benchmarkRuleModeStrictNoConst:
		if cfg.Linter.Rules.BestPractices != nil {
			cfg.Linter.Rules.BestPractices.AlwaysPreferConst = nil
		}
	case benchmarkRuleModeConstOnly:
		cfg.Linter.Rules = rules.LinterRulesGroup{
			BestPractices: &rules.BestPracticesRulesGroup{
				Use: true,
				AlwaysPreferConst: &rules.LinterBaseRule{
					Severity: "error",
				},
			},
		}
	}

	return cfg
}

func writeBenchmarkCorpus(b *testing.B, spec benchmarkCorpusSpec) benchmarkCorpus {
	b.Helper()

	root := b.TempDir()
	corpus := benchmarkCorpus{
		root:          root,
		packageFiles:  make([][]string, 0, spec.packages),
		totalPackages: spec.packages,
		totalFiles:    spec.packages * spec.filesPerPkg,
		totalFuncs:    spec.packages * spec.filesPerPkg * spec.funcsPerFile * 2,
	}

	for pkg := range spec.packages {
		dir := filepath.Join(root, fmt.Sprintf("pkg%02d", pkg))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			b.Fatalf("mkdir %s: %v", dir, err)
		}

		files := make([]string, 0, spec.filesPerPkg)
		for file := range spec.filesPerPkg {
			path := filepath.Join(dir, fmt.Sprintf("file%02d.go", file))
			src := benchmarkFileSource(spec, pkg, file)
			if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
				b.Fatalf("write %s: %v", path, err)
			}

			if corpus.singleFile == "" {
				corpus.singleFile = path
				corpus.singleFileSize = int64(len(src))
			}

			corpus.totalBytes += int64(len(src))
			files = append(files, path)
		}

		corpus.packageFiles = append(corpus.packageFiles, files)
	}

	return corpus
}

func (c benchmarkCorpus) target(kind string) (string, int, int, int64) {
	if kind == "file" {
		return c.singleFile, 1, 1, c.singleFileSize
	}

	return c.root, c.totalFiles, c.totalPackages, c.totalBytes
}

func benchmarkFileSource(spec benchmarkCorpusSpec, pkgIndex, fileIndex int) string {
	var src strings.Builder

	src.Grow(max(8192, spec.funcsPerFile*1024))
	fmt.Fprintf(&src, "package pkg%02d\n\n", pkgIndex)
	src.WriteString("import (\n")
	src.WriteString("\t\"context\"\n")
	src.WriteString("\tfmt \"fmt\"\n")
	src.WriteString(")\n\n")

	fmt.Fprintf(&src, "var neverMutated%d_%d = 42\n", pkgIndex, fileIndex)
	fmt.Fprintf(&src, "var mutated%d_%d = 7\n", pkgIndex, fileIndex)
	fmt.Fprintf(&src, "type empty%d_%d struct{}\n", pkgIndex, fileIndex)
	fmt.Fprintf(&src, "type sampleReceiver%d_%d struct{}\n\n", pkgIndex, fileIndex)

	fmt.Fprintf(&src, "func (receiver sampleReceiver%d_%d) ExportedMethod(sample string, count int) {\n", pkgIndex, fileIndex)
	src.WriteString("\t_ = sample\n")
	src.WriteString("\t_ = count\n")
	src.WriteString("}\n\n")

	for fn := range spec.funcsPerFile {
		fmt.Fprintf(&src, "func mutate%d_%d_%d() {\n", pkgIndex, fileIndex, fn)
		fmt.Fprintf(&src, "\tmutated%d_%d = %d\n", pkgIndex, fileIndex, fn+2)
		src.WriteString("}\n\n")

		fmt.Fprintf(&src, "func GetThing%d_%d_%d(ctx context.Context, id int, extra int, flag bool, sample string, count int) (first string, second string, err error) {\n", pkgIndex, fileIndex, fn)
		src.WriteString("\titems := make([]int, 8)\n")
		src.WriteString("\t_ = ctx\n")
		src.WriteString("\t_ = sample\n")
		src.WriteString("\tfor i := 0; i < count; i += 1 {\n")
		src.WriteString("\t\tdefer fmt.Println(i)\n")
		src.WriteString("\t\titems = append(items, i)\n")
		src.WriteString("\t\tif flag == true {\n")
		src.WriteString("\t\t\treturn \"\", \"\", fmt.Errorf(\"Bad value.\")\n")
		src.WriteString("\t\t}\n")
		src.WriteString("\t}\n")
		src.WriteString("\tif id > extra {\n")
		src.WriteString("\t\treturn\n")
		src.WriteString("\t}\n")
		src.WriteString("\tfmt.Println(\"")
		src.WriteString(strings.Repeat("x", spec.lineLength))
		src.WriteString("\")\n")
		src.WriteString("\treturn \"a\", \"b\", fmt.Errorf(\"Another Error.\")\n")
		src.WriteString("}\n\n")
	}

	return src.String()
}
