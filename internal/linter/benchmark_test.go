package linter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/serenitysz/serenity/internal/config"
	"github.com/serenitysz/serenity/internal/utils"
)

func BenchmarkProcessPathSequential(b *testing.B) {
	runProcessPathBenchmark(b, 1)
}

func BenchmarkProcessPathParallel(b *testing.B) {
	runProcessPathBenchmark(b, runtime.GOMAXPROCS(0))
}

func runProcessPathBenchmark(b *testing.B, workers int) {
	b.Helper()
	b.ReportAllocs()

	root := writeBenchmarkRepo(b)
	cfg := config.GenStrictDefaultConfig(utils.Ptr(false))
	cfg.Linter.Issues.Max = 0
	cfg.Performance.Threads = utils.Ptr(workers)

	l := New(false, false, cfg, 0, 0)

	b.ResetTimer()

	for range b.N {
		issues, err := l.ProcessPath(root)
		if err != nil {
			b.Fatalf("ProcessPath failed: %v", err)
		}

		if len(issues) == 0 {
			b.Fatal("expected issues from benchmark fixture")
		}
	}
}

func writeBenchmarkRepo(b *testing.B) string {
	b.Helper()

	root := b.TempDir()

	for pkg := range 12 {
		dir := filepath.Join(root, fmt.Sprintf("pkg%02d", pkg))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			b.Fatalf("mkdir %s: %v", dir, err)
		}

		for file := range 8 {
			path := filepath.Join(dir, fmt.Sprintf("file%02d.go", file))
			src := benchmarkFileSource(pkg, file)

			if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
				b.Fatalf("write %s: %v", path, err)
			}
		}
	}

	return root
}

func benchmarkFileSource(pkgIndex, fileIndex int) string {
	var src strings.Builder

	src.Grow(8192)
	fmt.Fprintf(&src, "package pkg%02d\n\n", pkgIndex)
	src.WriteString("import (\n")
	src.WriteString("\t\"context\"\n")
	src.WriteString("\tfmt \"fmt\"\n")
	src.WriteString(")\n\n")

	fmt.Fprintf(&src, "var neverMutated%d_%d = 42\n", pkgIndex, fileIndex)
	fmt.Fprintf(&src, "var mutated%d_%d = 7\n\n", pkgIndex, fileIndex)
	fmt.Fprintf(&src, "type empty%d_%d struct{}\n\n", pkgIndex, fileIndex)

	for fn := range 6 {
		fmt.Fprintf(&src, "func mutate%d_%d_%d() {\n", pkgIndex, fileIndex, fn)
		fmt.Fprintf(&src, "\tmutated%d_%d = %d\n", pkgIndex, fileIndex, fn+2)
		src.WriteString("}\n\n")

		fmt.Fprintf(&src, "func GetThing%d_%d_%d(ctx context.Context, id int, extra int, flag bool, sample string, count int) (err error) {\n", pkgIndex, fileIndex, fn)
		src.WriteString("\titems := make([]int, 8)\n")
		src.WriteString("\t_ = ctx\n")
		src.WriteString("\t_ = sample\n")
		src.WriteString("\tfor i := 0; i < count; i += 1 {\n")
		src.WriteString("\t\tdefer fmt.Println(i)\n")
		src.WriteString("\t\titems = append(items, i)\n")
		src.WriteString("\t\tif flag == true {\n")
		src.WriteString("\t\t\treturn fmt.Errorf(\"Bad value.\")\n")
		src.WriteString("\t\t}\n")
		src.WriteString("\t}\n")
		src.WriteString("\tif id > extra {\n")
		src.WriteString("\t\treturn\n")
		src.WriteString("\t}\n")
		src.WriteString("\tfmt.Println(\"")
		src.WriteString(strings.Repeat("x", 120))
		src.WriteString("\")\n")
		src.WriteString("\treturn fmt.Errorf(\"Another Error.\")\n")
		src.WriteString("}\n\n")
	}

	return src.String()
}
