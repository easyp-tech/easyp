package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/mod/models"
)

// CacheDownload create and return path to downloaded cache.
// Like $GOPATH/pkg/mod/cache/download
func (s *Storage) CreateCacheDownloadDir(module models.Module) (string, error) {
	cacheDownloadPath := filepath.Join(s.rootDir, cacheDir, cacheDownloadDir, module.Name)

	if err := os.MkdirAll(cacheDownloadPath, dirPerm); err != nil {
		return "", fmt.Errorf("os.MkdirAll: %w", err)
	}

	return cacheDownloadPath, nil
}
