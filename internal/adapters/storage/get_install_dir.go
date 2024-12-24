package storage

import (
	"path"

	"github.com/easyp-tech/easyp/internal/core/adapters"
)

// getInstallDir returns dir to install package
// rootDir + installedDir + module full remote path + module's version
// eg: ~/.EASYP/mod/github.com/google/googleapis/v1.2.3
func (s *Storage) GetInstallDir(moduleName string, version string) string {
	version = adapters.SanitizePath(version)
	return path.Join(s.rootDir, installedDir, moduleName, version)
}
