package cmd

import (
	"fmt"

	"github.com/serenitysz/serenity/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes serenity, creating a serenity.json configuration file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit() error {
	path, err := config.GetConfigFilePath()
	if err != nil {
		return err
	}

	exists, err := config.CheckHasConfigFile(path)
	if err != nil {
		return fmt.Errorf("error checking for config file: %w", err)
	}

	if exists {
		fmt.Println("Config file serenity.json already exists.")
		return nil
	}

	fmt.Println("Creating default serenity.json config file...")
	cfg := config.GenDefaultConfig()

	if err := config.CreateConfigFile(cfg, path); err != nil {
		return err
	}
	fmt.Println("Config file created successfully.")

	return nil
}
