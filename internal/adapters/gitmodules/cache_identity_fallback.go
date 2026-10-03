//go:build !darwin && !linux

package gitmodules

import "os"

func cacheFileIdentity(os.FileInfo) (cacheIdentity, bool) {
	return cacheIdentity{}, false
}
