package factories

import (
	lockfilePkg "github.com/easyp-tech/easyp/internal/mod/adapters/lock_file"
)

var lockFile *lockfilePkg.LockFile

func NewLockFile() *lockfilePkg.LockFile {
	if lockFile != nil {
		return lockFile
	}

	lockFile = lockfilePkg.New()
	return lockFile
}
