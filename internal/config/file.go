package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/almeidazs/gowther/internal/rules"
)

func GetConfigFilePath() (path string, err error) {
	path, err = os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error to get user working directory: %w", err)
	}

	return filepath.Join(path, "gowther.json"), nil
}

func CreateConfigFile(cfg *rules.Config, path string) error {
	_, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating config file: %w", err)
	}

	jsonData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %v", err)
	}

	err = os.WriteFile("gowther.json", jsonData, 0o644)
	if err != nil {
		return fmt.Errorf("error writing gowther.json: %v", err)
	}

	return nil
}

func CheckHasConfigFile(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, fmt.Errorf("error to get file info: %w", err)
	}

	return true, nil
}

func ReadConfigFile(path string) (*rules.Config, error) {
	var cfg *rules.Config
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file '%s' not found", path)
		}
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	if err := json.Unmarshal(fileBytes, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file json: %w", err)
	}

	return cfg, nil
}
