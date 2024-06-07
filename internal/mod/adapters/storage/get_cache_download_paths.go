package storage

import (
	"path/filepath"
	"strings"

	"github.com/easyp-tech/easyp/internal/mod/models"
)

// GetDownloadArchivePath returns full path to download archive (include extension)
func (s *Storage) GetCacheDownloadPaths(module models.Module, revision models.Revision) models.CacheDownloadPaths {
	cacheDownloadDir := filepath.Join(s.rootDir, cacheDir, cacheDownloadDir, module.Name)

	version := strings.Replace(revision.Version, "/", "-", -1)

	archiveFile := filepath.Join(cacheDownloadDir, version) + ".zip"
	archiveHashFile := filepath.Join(cacheDownloadDir, version) + ".ziphash"
	moduleInfoFile := filepath.Join(cacheDownloadDir, version) + ".info"

	return models.CacheDownloadPaths{
		CacheDownloadDir: cacheDownloadDir,
		ArchiveFile:      archiveFile,
		ArchiveHashFile:  archiveHashFile,
		ModuleInfoFile:   moduleInfoFile,
	}
}
