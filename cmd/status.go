package cmd

import (
	"fmt"
	"runtime"

	"github.com/serenitysz/serenity/internal/config"
	"github.com/serenitysz/serenity/internal/version"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use: "status",
	Short: "Display the current status of Serenity",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getStatus()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func getStatus() error {
	fmt.Println("Serenity:")
	fmt.Printf("  Version:                      %s\n", version.Version)
	fmt.Printf("  Commit:            		%s\n", version.Commit)

	fmt.Println("\nPlatform:")
	fmt.Printf("  CPU Architecture:             %s\n", runtime.GOARCH)
	fmt.Printf("  OS:                           %s\n", runtime.GOOS)
	fmt.Printf("  GO_VERSION:                   %s\n", runtime.Version())

	fmt.Println("\nSerenity Configuration:")

	path, _ := config.GetConfigFilePath()
	exists, err := config.CheckHasConfigFile(path)

	status := "Not found"

	if err == nil && exists {
		status = "Loaded successfully"
	}

	fmt.Printf("  Status:                       %s\n", status)
	fmt.Printf("  Path:                         %s\n", path)

	return nil
}
