package linter

import (
	"go/parser"
	"runtime"

	"github.com/serenitysz/serenity/internal/rules"
)

type Linter struct {
	Write       bool
	Unsafe      bool
	MaxIssues   int // 0 means unlimited
	MaxFileSize int64
	Config      *rules.LinterOptions
	Workers     int
	ParseMode   parser.Mode
	ActiveRules *ActiveRules
	Cache       *cacheStore
}

func New(write, unsafe bool, config *rules.LinterOptions, maxIssues int, maxFileSize int64) *Linter {
	workers := runtime.GOMAXPROCS(0)
	activeRules := BuildActiveRules(config)
	autofix := activeRules.HasAutofixRules && (write || config.ShouldAutofix())
	effectiveMaxIssues := maxIssues
	parseMode := parser.ParseComments | parser.SkipObjectResolution

	if activeRules.NeedsConstAnalysis {
		parseMode = parser.ParseComments
	}

	if autofix {
		effectiveMaxIssues = 0
	}

	if perf := config.Performance; perf != nil && perf.Use && perf.Threads != nil && *perf.Threads > 0 {
		workers = *perf.Threads
	}

	// Limited runs should be deterministic: stop after the first issues in
	// traversal order instead of whichever packages finish first.
	if effectiveMaxIssues > 0 {
		workers = 1
	}

	return &Linter{
		Write:       write,
		Unsafe:      unsafe,
		Config:      config,
		MaxIssues:   effectiveMaxIssues,
		MaxFileSize: maxFileSize,
		Workers:     workers,
		ParseMode:   parseMode,
		ActiveRules: activeRules,
		Cache:       newCacheStore(config, autofix, unsafe),
	}
}
