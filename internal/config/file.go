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

var CANDIDATES = [4]string{
	"serenity.json",
	"serenity.yaml",
	"serenity.yml",
	"serenity.toml",
}

func Exists(path string) (bool, error) {
	info, err := os.Stat(path)

	if err == nil {
		return !info.IsDir(), nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func SearchConfigPath() (string, error) {
	if path, err := getFromEnv(); err != nil {
		return path, nil
	}

	wd, err := os.Getwd()

	if err != nil {
		return "", fmt.Errorf("cannot get working directory: %w", err)
	}

	if path, ok := Scan(wd); ok {
		return path, nil
	}

	return "", fmt.Errorf(
		"no serenity config file found (looked for %v). Run `serenity init`",
		CANDIDATES,
	)
}

func Scan(start string) (string, bool) {
	dir := start

	for {
		for _, name := range CANDIDATES {
			path := filepath.Join(dir, name)

			if ok, _ := Exists(path); ok {
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

func getFromEnv() (string, error) {
	env := os.Getenv("SERENITY_CONFIG_PATH")

	if env == "" {
		return "", nil
	}

	if ok, err := Exists(env); err != nil {
		return "", err
	} else if !ok {
		return "", fmt.Errorf("SERENITY_CONFIG_PATH points to a non-existent file: %s", env)
	}

	return env, nil

}

func marshalConfigByExt(config *rules.LinterOptions, path string) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(path))

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
