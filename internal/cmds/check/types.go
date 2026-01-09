package check

type CheckOptions struct {
	Write       bool
	Unsafe      bool
	MaxFileSize int64
	ConfigPath  string
}
