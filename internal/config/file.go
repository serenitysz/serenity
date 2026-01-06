package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/pelletier/go-toml/v2"
	"github.com/serenitysz/serenity/internal/rules"
)

var paths = [4]string{
	"serenity.json",
	"serenity.yaml",
	"serenity.yml",
	"serenity.toml",
}

func GetConfigFilePath() (string, error) {
	if path, err := getFromEnv(); err != nil {
		return path, nil
	}

	wd, err := os.Getwd()

	if err != nil {
		return "", fmt.Errorf("cannot get working directory: %w", err)
	}

	if path, ok := findConfigUpwards(wd); ok {
		return path, nil
	}

	return "", fmt.Errorf(
		"no serenity config file found (looked for %v). Run `serenity init`",
		paths,
	)
}

func getFromEnv() (string, error) {
	env := os.Getenv("SERENITY_CONFIG_PATH")

	if env == "" {
		return "", nil
	}

	if ok, err := fileExists(env); err != nil {
		return "", err
	} else if !ok {
		return "", fmt.Errorf("SERENITY_CONFIG_PATH points to a non-existent file: %s", env)
	}

	return env, nil

}

func findConfigUpwards(start string) (string, bool) {
	dir := start

	for {
		for _, name := range paths {
			path := filepath.Join(dir, name)

			if ok, _ := fileExists(path); ok {
				return path, true
			}
		}

		parent := filepath.Dir(dir)

		if parent == dir {
			return "", false
		}

		dir = parent
	}
}

func fileExists(path string) (bool, error) {
	info, err := os.Stat(path)

	if err == nil {
		return !info.IsDir(), nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func CreateConfigFile(config *rules.LinterOptions, path string) error {
	data, err := marshalConfigByExt(config, path)

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func marshalConfigByExt(config *rules.LinterOptions, path string) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(path))

	if ext == "" {
		return nil, fmt.Errorf("config file has no extension: %s", path)
	}

	cfg := *config

	if ext != ".json" {
		cfg.Schema = ""
	}

	switch ext {
	case ".json":
		return json.MarshalIndent(&cfg, "", "\t")

	case ".toml":
		return toml.Marshal(&cfg)

	case ".yml", ".yaml":
		return yaml.Marshal(&cfg)

	default:
		return nil, fmt.Errorf(
			"unsupported config format %q (supported: JSON, TOML, YAML, YML)",
			ext,
		)
	}
}

func CheckHasConfigFile(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}
