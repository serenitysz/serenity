package status

import (
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/serenitysz/serenity/internal/config"
	"github.com/serenitysz/serenity/internal/exception"
	"github.com/serenitysz/serenity/internal/render"
	"github.com/serenitysz/serenity/internal/version"
)

func Get(noColor bool) error {
	path, err := config.SearchConfigPath()

	if err != nil {
		return exception.InternalError("%v", err)
	}

	status := "Not found"

	if path != "" {
		status = "Loaded successfully"
	}

	var b strings.Builder

	b.Grow(256)

	b.WriteString("Serenity:\n")
	b.WriteString("  Commit:                       ")
	b.WriteString(version.Commit)
	b.WriteByte('\n')
	b.WriteString("  Version:                      ")
	b.WriteString(version.Version)
	b.WriteString("\n\n")

	b.WriteString("Platform:\n")
	b.WriteString("  OS:                           ")
	b.WriteString(runtime.GOOS)
	b.WriteByte('\n')
	b.WriteString("  CPU Architecture:             ")
	b.WriteString(runtime.GOARCH)
	b.WriteByte('\n')
	b.WriteString("  GO_VERSION:                   ")
	b.WriteString(runtime.Version())
	b.WriteString("\n\n")

	b.WriteString("Serenity Configuration:\n")
	b.WriteString("  Status:                       ")
	b.WriteString(status)
	b.WriteByte('\n')
	b.WriteString("  Path:                         ")
	b.WriteString(render.Paint(path, render.Purple, noColor))
	b.WriteByte('\n')

	if _, err := io.WriteString(os.Stdout, b.String()); err != nil {
		return exception.InternalError("%v", err)
	}

	return nil
}
