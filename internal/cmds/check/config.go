package check

import (
	"github.com/serenitysz/serenity/internal/config"
	"github.com/serenitysz/serenity/internal/rules"
)

func loadConfig(path string) (*rules.LinterOptions, error) {
	if path == "" {
		p, err := config.SearchConfigPath()

		if err != nil {
			return nil, err
		}

		path = p
	}

	cfg := config.GenDefaultConfig(new(bool))

	exists, err := config.Exists(path)

	if err != nil {
		return nil, err
	}

	if exists {
		loaded, err := config.Read(path)

		if err != nil {
			return nil, err
		}

		cfg = loaded
	}

	config.ApplyRecommended(cfg)

	return cfg, nil
}
