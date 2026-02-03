package cmd

import (
	"os/exec"
	"runtime"

	"github.com/serenitysz/serenity/internal/exception"
	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Open Serenity documentation in the browser quickly",
	RunE: func(cmd *cobra.Command, args []string) error {
		return open("https://docs-blond.vercel.app/")
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
}

func open(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	if err := cmd.Start(); err != nil {
		return exception.InternalError("%v", err)
	}

	return nil
}
