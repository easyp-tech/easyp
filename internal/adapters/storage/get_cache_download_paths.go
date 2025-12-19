package storage

import (
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/core/models"
)

// GetDownloadArchivePath returns full path to download archive (include extension)
func (s *Storage) GetCacheDownloadPaths(moduleName, version string) models.CacheDownloadPaths {
	cacheDownloadDir := filepath.Join(s.rootDir, cacheDir, cacheDownloadDir, moduleName)

	fileName := sanitizePath(version)

	archiveFile := filepath.Join(cacheDownloadDir, fileName) + ".zip"
	moduleInfoFile := filepath.Join(cacheDownloadDir, fileName) + ".info"

	return models.CacheDownloadPaths{
		CacheDownloadDir: cacheDownloadDir,
		ArchiveFile:      archiveFile,
		ModuleInfoFile:   moduleInfoFile,
	}
}
