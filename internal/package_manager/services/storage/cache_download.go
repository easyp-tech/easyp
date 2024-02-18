package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/package_manager/models"
)

// CacheDownload create and return path to downloaded cache.
// Like $GOPATH/pkg/mod/cache/download
func (s *Storage) CacheDownload(module models.Module) (string, error) {
	cacheDownloadPath := filepath.Join(s.rootDir, cacheDownload, module.Name)

	if err := os.MkdirAll(cacheDownloadPath, cacheDirPerm); err != nil {
		return "", fmt.Errorf("os.MkdirAll: %w", err)
	}

	return cacheDownloadPath, nil
}
