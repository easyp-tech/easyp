package lockfile

import (
	"bufio"
	"strings"

	"github.com/easyp-tech/easyp/internal/core"
)

const (
	lockFileName  = "easyp.lock"
	lockFilePerms = 0644
)

type fileInfo struct {
	version string
	hash    string
}

type LockFile struct {
	dirWalker core.DirWalker
	cache     map[string]fileInfo
}

func New(dirWalker core.DirWalker) *LockFile {
	cache := make(map[string]fileInfo)

	fp, err := dirWalker.Open(lockFileName)
	if err == nil {
		fscanner := bufio.NewScanner(fp)

		for fscanner.Scan() {
			parts := strings.Fields(fscanner.Text())
			if len(parts) != 3 {
				continue
			}

			fileInfo := fileInfo{
				version: parts[1],
				hash:    parts[2],
			}
			cache[parts[0]] = fileInfo
		}
	}

	lockFile := &LockFile{
		dirWalker: dirWalker,
		cache:     cache,
	}
	return lockFile
}
