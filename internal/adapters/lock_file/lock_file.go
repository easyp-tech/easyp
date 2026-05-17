package lockfile

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/easyp-tech/easyp/internal/core"
)

const (
	lockFileName = "easyp.lock"
)

type fileInfo struct {
	version string
	hash    string
}

type LockFile struct {
	dirWalker core.DirWalker
	cache     map[string]fileInfo
}

func New(dirWalker core.DirWalker) (*LockFile, error) {
	cache := make(map[string]fileInfo)

	fp, err := dirWalker.Open(lockFileName)
	if err == nil {
		defer func() { _ = fp.Close() }()

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

		if err := fscanner.Err(); err != nil {
			return nil, fmt.Errorf("scan %s: %w", lockFileName, err)
		}
	}

	lockFile := &LockFile{
		dirWalker: dirWalker,
		cache:     cache,
	}
	return lockFile, nil
}
