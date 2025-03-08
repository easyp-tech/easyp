package storage

import (
	"path/filepath"

	"go.redsock.ru/protopack/internal/core/adapters"
	"go.redsock.ru/protopack/internal/core/models"
)

// GetDownloadArchivePath returns full path to download archive (include extension)
func (s *Storage) GetCacheDownloadPaths(module models.Module, revision models.Revision) models.CacheDownloadPaths {
	cacheDownloadDir := filepath.Join(s.rootDir, cacheDir, cacheDownloadDir, module.Name)

	fileName := adapters.SanitizePath(revision.Version)

	archiveFile := filepath.Join(cacheDownloadDir, fileName) + ".zip"
	archiveHashFile := filepath.Join(cacheDownloadDir, fileName) + ".ziphash"
	moduleInfoFile := filepath.Join(cacheDownloadDir, fileName) + ".info"

	return models.CacheDownloadPaths{
		CacheDownloadDir: cacheDownloadDir,
		ArchiveFile:      archiveFile,
		ArchiveHashFile:  archiveHashFile,
		ModuleInfoFile:   moduleInfoFile,
	}
}
