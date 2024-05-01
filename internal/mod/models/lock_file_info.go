package models

import (
	"errors"
)

// LockFileInfo contains information about module from lock file
type LockFileInfo struct {
	Name    string
	Version string
	Hash    ModuleHash
}

var (
	ErrModuleNotFoundInLockFile = errors.New("module not found in lock file")
)
