package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Serenity to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		return update(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func update(ctx context.Context) error {
	if rootCmd.Version == "" {
		return errors.New("current version is not set")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)

	defer cancel()

	const slug = "serenitysz/serenity"

	latest, found, err := selfupdate.DetectLatest(ctx, selfupdate.ParseSlug(slug))

	if err != nil {
		return fmt.Errorf("failed to check latest version: %w", err)
	}

	if !found || latest.LessOrEqual(rootCmd.Version) {
		fmt.Printf("You're already running the latest version (%s)\n", rootCmd.Version)
	
		return nil
	}

	exe, err := os.Executable()

	if err != nil {
		return fmt.Errorf("failed to locate executable: %w", err)
	}

	fmt.Printf("Updating Serenity from %s to %s...\n", rootCmd.Version, latest.Version())

	if err := selfupdate.UpdateTo(ctx, latest.AssetURL, latest.AssetName, exe); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	fmt.Printf("You'are now on the %s of Serenity!\n", latest.Version())

	return nil
}
