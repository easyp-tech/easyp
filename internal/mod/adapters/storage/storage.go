package storage

import (
	"path"
)

const (
	// root cache dir
	cacheDir = "cache"
	// dir for downloaded (check sum, archive)
	cacheDownloadDir = "download"
	// dir for installed packages
	installedDir = "mod"
)

// Storage implements workflows with directories
type Storage struct {
	rootDir string
}

func New(rootDir string) *Storage {
	return &Storage{
		rootDir: rootDir,
	}
}

const (
	dirPerm = 0755
)

// getInstallDir returns dir to install package
// rootDir + installedDir + module full remote path + module's version
// eg: ~/.EASYP/mod/github.com/google/googleapis/v1.2.3
func (s *Storage) getInstallDir(moduleName string, version string) string {
	return path.Join(s.rootDir, installedDir, moduleName, version)
}
