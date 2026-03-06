package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/serenitysz/serenity/internal/rules"
)

func TestSearchConfigPathUsesEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "serenity.json")

	if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("SERENITY_CONFIG_PATH", path)

	got, err := SearchConfigPath()
	if err != nil {
		t.Fatal(err)
	}

	if got != path {
		t.Fatalf("expected env path %q, got %q", path, got)
	}
}

func TestSearchConfigPathReturnsEmptyWhenMissing(t *testing.T) {
	t.Setenv("SERENITY_CONFIG_PATH", "")
	t.Chdir(t.TempDir())

	got, err := SearchConfigPath()
	if err != nil {
		t.Fatal(err)
	}

	if got != "" {
		t.Fatalf("expected empty path, got %q", got)
	}
}

func TestApplyRecommendedPersistsMutations(t *testing.T) {
	useRecommended := true
	cfg := &rules.LinterOptions{
		Linter: rules.LinterRules{
			Use: true,
			Rules: rules.LinterRulesGroup{
				UseRecommended: &useRecommended,
			},
		},
	}

	ApplyRecommended(cfg)

	if cfg.Assistance == nil || !cfg.Assistance.Use {
		t.Fatal("expected assistance options to be persisted")
	}

	if cfg.Linter.Rules.Imports == nil || !cfg.Linter.Rules.Imports.Use {
		t.Fatal("expected imports recommendations to be enabled")
	}

	if cfg.Linter.Rules.BestPractices == nil || !cfg.Linter.Rules.BestPractices.Use {
		t.Fatal("expected best practices recommendations to be enabled")
	}

	if cfg.Linter.Rules.Complexity == nil || !cfg.Linter.Rules.Complexity.Use {
		t.Fatal("expected complexity recommendations to be enabled")
	}
}
