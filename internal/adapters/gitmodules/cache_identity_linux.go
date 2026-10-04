//go:build linux

package gitmodules

import (
	"os"
	"syscall"
)

func cacheFileIdentity(info os.FileInfo) (cacheIdentity, bool) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return cacheIdentity{}, false
	}
	return cacheIdentity{
		Device:            uint64(stat.Dev),
		Inode:             uint64(stat.Ino),
		Links:             uint64(stat.Nlink),
		ChangeSeconds:     int64(stat.Ctim.Sec),
		ChangeNanoseconds: int64(stat.Ctim.Nsec),
	}, true
}
