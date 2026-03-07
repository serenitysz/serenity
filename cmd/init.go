package cmd

import (
	"path/filepath"

	"github.com/serenitysz/serenity/internal/config"
	"github.com/serenitysz/serenity/internal/exception"
	"github.com/serenitysz/serenity/internal/prompts"
	"github.com/serenitysz/serenity/internal/render"
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
			noColor, err := cmd.Flags().GetBool("no-color")

			if err != nil {
				return exception.InternalError("could not read --no-color: %w", err)
			}

			return createSerenityInteractive(noColor)
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

	ok, err := config.Exists(path)
	if err != nil {
		return err
	}

	if ok {
		return exception.CommandError("config file %q already exists", path)
	}

	autofix := false
	cfg := config.GenDefaultConfig(&autofix)

	return createConfig(path, cfg)
}

func createSerenityInteractive(noColor bool) error {
	format, err := prompts.Input(
		"Which config format do you want to use? (JSON, YAML, TOML)", "JSON", noColor)

	if err != nil {
		return exception.InternalError("could not read the selected config format: %w", err)
	}

	ext, ok := map[string]string{
		"JSON": ".json",
		"YAML": ".yaml",
		"TOML": ".toml",
	}[format]

	if !ok {
		return exception.CommandError("unsupported config format %q; choose JSON, YAML, or TOML", format)
	}

	defaultPath := "serenity" + ext

	path, err := prompts.Input(
		"Config file path",
		defaultPath,
		noColor,
	)

	if err != nil {
		return exception.InternalError("could not read the config path: %w", err)
	}

	if !isSupportedConfig(path) {
		return exception.CommandError("unsupported config format %q; supported formats: .json, .yaml, .yml, .toml", filepath.Ext(path))
	}

	ok, err = config.Exists(path)
	if err != nil {
		return err
	}

	if ok {
		overwrite, err := prompts.Confirm("Config already exists. Overwrite?", noColor)

		if err != nil {
			return exception.InternalError("could not confirm whether to overwrite %q: %w", path, err)
		}

		if !overwrite {
			return nil
		}
	}

	strict, err := prompts.Confirm("Enable strict preset?", noColor)

	if err != nil {
		return exception.InternalError("could not read the strict preset option: %w", err)
	}

	autofix, err := prompts.Confirm("Enable autofix when possible?", noColor)

	if err != nil {
		return exception.InternalError("could not read the autofix option: %w", err)
	}

	cfg := config.GenDefaultConfig(&autofix)

	if strict {
		cfg = config.GenStrictDefaultConfig(&autofix)
	}

	return createConfig(path, cfg)
}

func createConfig(path string, cfg *rules.LinterOptions) error {
	render.Infof("creating config file at %s", path)

	if err := config.CreateConfigFile(cfg, path); err != nil {
		return err
	}

	render.Successf("config file created at %s", path)

	return nil
}
