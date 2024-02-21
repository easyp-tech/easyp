package storage

import (
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/mod/models"
)

// GetDownloadArchivePath returns full path to download archive (include extension)
func (s *Storage) GetDownloadArchivePath(cacheDownloadPath string, revision models.Revision) string {
	return filepath.Join(cacheDownloadPath, revision.Version) + ".zip"
}
