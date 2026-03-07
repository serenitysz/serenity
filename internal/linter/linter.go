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
	parseMode := parser.ParseComments | parser.SkipObjectResolution

	if activeRules.NeedsConstAnalysis {
		parseMode = parser.ParseComments
	}

	if perf := config.Performance; perf != nil && perf.Use && perf.Threads != nil && *perf.Threads > 0 {
		workers = *perf.Threads
	}

	return &Linter{
		Write:       write,
		Unsafe:      unsafe,
		Config:      config,
		MaxIssues:   maxIssues,
		MaxFileSize: maxFileSize,
		Workers:     workers,
		ParseMode:   parseMode,
		ActiveRules: activeRules,
		Cache:       newCacheStore(config, write, unsafe),
	}
}
