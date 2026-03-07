//go:build !linux

package linter

import "os"

type fileStamp struct {
	ChangeTimeUnixNano int64
	Device             uint64
	Inode              uint64
	FastPathSupported  bool
}

func readFileStamp(info os.FileInfo) fileStamp {
	return fileStamp{}
}
