package lockfile

import (
	"bufio"
	"log"
	"os"
	"strings"
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
	fp    *os.File
	cache map[string]fileInfo
}

func New() *LockFile {
	fp, err := os.OpenFile(lockFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, lockFilePerms)
	if err != nil {
		// TODO: return err?
		log.Fatal(err)
	}

	cache := make(map[string]fileInfo)

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

	lockFile := &LockFile{
		fp:    fp,
		cache: cache,
	}
	return lockFile
}
