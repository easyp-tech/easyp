package storage

import (
	"github.com/easyp-tech/easyp/internal/core/models"
)

const (
	// root cache dir
	cacheDir = "cache"
	// dir for downloaded (check sum, archive)
	cacheDownloadDir = "download"
	// dir for installed packages
	installedDir = "mod"
)

type (
	// LockFile should implement adapter for lock file workflow
	LockFile interface {
		Read(moduleName string) (models.LockFileInfo, error)
	}

	// Storage implements workflows with directories
	Storage struct {
		rootDir  string
		lockFile LockFile
	}
)

func New(rootDir string, lockFile LockFile) *Storage {
	return &Storage{
		rootDir:  rootDir,
		lockFile: lockFile,
	}
}

const (
	dirPerm      = 0755
	infoFilePerm = 0644
)
