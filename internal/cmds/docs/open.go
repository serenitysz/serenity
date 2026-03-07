package docs

import (
	"os/exec"
	"runtime"

	"github.com/serenitysz/serenity/internal/exception"
)

const DOCS_URL = "https://docs-blond.vercel.app/"

func Open() error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", DOCS_URL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", DOCS_URL)
	default:
		cmd = exec.Command("xdg-open", DOCS_URL)
	}

	if err := cmd.Start(); err != nil {
		return exception.InternalError("could not open the Serenity documentation in your browser: %w", err)
	}

	return nil
}
