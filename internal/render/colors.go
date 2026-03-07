package render

import (
	"fmt"
	"os"
)

var IS_COLOR_ENABLED = os.Getenv("NO_COLOR") == ""
var forceNoColor bool

const (
	Reset = "\033[0m"
	Bold  = "\033[1m"

	Purple = "\033[38;5;141m"
	Gray   = "\033[90m"
	Blue   = "\033[34m"
	Green  = "\033[32m"
	Red    = "\033[31m"
	Yellow = "\033[33m"
)

func SetNoColor(noColor bool) {
	forceNoColor = noColor
}

func NoColor(noColor bool) bool {
	return noColor || forceNoColor || !IS_COLOR_ENABLED
}

func Paint(content, code string, noColor bool) string {
	if NoColor(noColor) || code == "" {
		return content
	}

	return code + content + Reset
}

func Tag(label, code string, noColor bool) string {
	return Paint(fmt.Sprintf("%-5s", label), code, noColor)
}
