package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/pelletier/go-toml/v2"
	"github.com/serenitysz/serenity/internal/exception"
	"github.com/serenitysz/serenity/internal/rules"
)

func Read(path string) (*rules.LinterOptions, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, exception.InternalError("%v", err)
	}

	ext := strings.ToLower(filepath.Ext(path))

	if ext == "" {
		return nil, exception.InternalError("config file has no extension: %s", path)
	}

	var cfg rules.LinterOptions

	if err := unmarshalByExt(ext, data, &cfg); err != nil {
		return nil, exception.InternalError("failed to parse config %q: %w", path, err)
	}

	return &cfg, nil
}

func unmarshalByExt(ext string, data []byte, out any) error {
	switch ext {
	case ".json":
		return json.Unmarshal(data, out)

	case ".toml":
		return toml.Unmarshal(data, out)

	case ".yml", ".yaml":
		return yaml.Unmarshal(data, out)

	default:
		return exception.InternalError(
			"unsupported config format %q (supported: JSON, TOML, YAML, YML)",
			ext,
		)
	}
}
