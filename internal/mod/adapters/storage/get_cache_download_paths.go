package storage

import (
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/mod/models"
)

// GetDownloadArchivePath returns full path to download archive (include extension)
func (s *Storage) GetCacheDownloadPaths(module models.Module, revision models.Revision) models.CacheDownloadPaths {
	cacheDownloadDir := filepath.Join(s.rootDir, cacheDir, cacheDownloadDir, module.Name)
	archiveFile := filepath.Join(cacheDownloadDir, revision.Version) + ".zip"
	archiveHashFile := filepath.Join(cacheDownloadDir, revision.Version) + ".ziphash"
	moduleInfoFile := filepath.Join(cacheDownloadDir, revision.Version) + ".info"

	return models.CacheDownloadPaths{
		CacheDownloadDir: cacheDownloadDir,
		ArchiveFile:      archiveFile,
		ArchiveHashFile:  archiveHashFile,
		ModuleInfoFile:   moduleInfoFile,
	}
}
