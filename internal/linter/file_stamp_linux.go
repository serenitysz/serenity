//go:build linux

package linter

import (
	"os"
	"syscall"
)

type fileStamp struct {
	ChangeTimeUnixNano int64
	Device             uint64
	Inode              uint64
	FastPathSupported  bool
}

func readFileStamp(info os.FileInfo) fileStamp {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return fileStamp{}
	}

	return fileStamp{
		ChangeTimeUnixNano: stat.Ctim.Sec*1e9 + stat.Ctim.Nsec,
		Device:             uint64(stat.Dev),
		Inode:              uint64(stat.Ino),
		FastPathSupported:  true,
	}
}
