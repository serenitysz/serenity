package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/serenitysz/serenity/internal/rules"
)

func GetConfigFilePath() (string, error) {
	if path := os.Getenv("SERENITY_CONFIG_PATH"); path != "" {
		return path, nil
	}

	wd, err := os.Getwd()

	if err != nil {
		return "", fmt.Errorf("cannot get working directory: %w", err)
	}

	return filepath.Join(wd, "serenity.json"), nil
}

func CreateConfigFile(config *rules.LinterOptions, path string) error {
	data, err := json.Marshal(config)

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
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

func ReadConfig(path string) (*rules.LinterOptions, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var config rules.LinterOptions

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &config, nil
}
