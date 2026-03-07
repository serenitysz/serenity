package render

import (
	"fmt"
	"io"
	"os"
)

func Logf(w io.Writer, label, color, format string, args ...any) {
	_, _ = fmt.Fprintf(w, "%s %s\n", Tag(label, color, false), fmt.Sprintf(format, args...))
}

func Infof(format string, args ...any) {
	Logf(os.Stdout, "info", Blue, format, args...)
}

func Successf(format string, args ...any) {
	Logf(os.Stdout, "done", Green, format, args...)
}

func Warnf(format string, args ...any) {
	Logf(os.Stderr, "warn", Yellow, format, args...)
}
