package internal

import "go/token"

type Linter struct {
	write      bool
	unsafe     bool
	fset       *token.FileSet
	fixes      []Fix
	violations []Violation
}

type Fix struct {
	File    string
	Content []byte
}

type Violation struct {
	File    string
	Line    int
	Column  int
	Message string
	Level   string // "error", "warning", "unsafe"
}
