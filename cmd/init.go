package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/serenitysz/serenity/internal/config"
	"github.com/serenitysz/serenity/internal/exception"
	"github.com/serenitysz/serenity/internal/prompts"
	"github.com/serenitysz/serenity/internal/rules"
	"github.com/spf13/cobra"
)

var (
	interactive bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new Serenity project by creating a serenity config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if interactive {
			return createSerenityInteractive()
		}

		return createSerenity()
	},
}

func init() {
	initCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Use an interactive mode to init Serenity project")

	rootCmd.AddCommand(initCmd)
}

var supportedConfigExts = map[string]struct{}{
	".json": {},
	".yaml": {},
	".yml":  {},
	".toml": {},
}

func isSupportedConfig(path string) bool {
	_, ok := supportedConfigExts[filepath.Ext(path)]

	return ok
}

func createSerenity() error {
	path := "serenity.json"

	if ok, _ := config.Exists(path); ok {
		return exception.InternalError("config file already exists: %s", path)
	}

	autofix := false
	cfg := config.GenDefaultConfig(&autofix)

	return createConfig(path, cfg)
}

func createSerenityInteractive() error {
	format, err := prompts.Input(
		"Which config format do you want to use? (JSON, YAML, TOML)", "JSON")

	if err != nil {
		return exception.InternalError("%v", err)
	}

	ext, ok := map[string]string{
		"JSON": ".json",
		"YAML": ".yaml",
		"TOML": ".toml",
	}[format]

	if !ok {
		return exception.InternalError("unsupported config format: %s", format)
	}

	defaultPath := "serenity" + ext

	path, err := prompts.Input(
		"Config file path",
		defaultPath,
	)

	if err != nil {
		return exception.InternalError("%v", err)
	}

	if !isSupportedConfig(path) {
		return exception.InternalError("unsupported config format: %s", filepath.Ext(path))
	}

	if ok, _ := config.Exists(path); ok {
		overwrite, err := prompts.Confirm("Config already exists. Overwrite?")

		if err != nil {
			return exception.InternalError("%v", err)
		}

		if !overwrite {
			return nil
		}
	}

	strict, err := prompts.Confirm("Enable strict preset?")

	if err != nil {
		return err
	}

	autofix, err := prompts.Confirm("Enable autofix when possible?")

	if err != nil {
		return exception.InternalError("%v", err)
	}

	cfg := config.GenDefaultConfig(&autofix)

	if strict {
		cfg = config.GenStrictDefaultConfig(&autofix)
	}

	return createConfig(path, cfg)
}

func createConfig(path string, cfg *rules.LinterOptions) error {
	fmt.Printf("Creating %s...\n", path)

	if err := config.CreateConfigFile(cfg, path); err != nil {
		return exception.InternalError("%v", err)
	}

	fmt.Println("Config file created successfully.")

	return nil
}
