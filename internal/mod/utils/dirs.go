package utils

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/mod/dependency"
)

const (
	cacheDirPerm = 0755
)

// CreateCacheDir create and return path to cache dir.
// cache dir contains repo cache for repository with proto files.
// cmd/go/internal/modfetch/codehost/codehost.go: 228 - create workdir
func CreateCacheDir(dep dependency.Dependency) (string, error) {
	tmpDir := os.TempDir()
	key := dep.Name + ":" + dep.Version
	cacheDir := filepath.Join(tmpDir, fmt.Sprintf("%x", sha256.Sum256([]byte(key))))

	if err := os.MkdirAll(cacheDir, cacheDirPerm); err != nil {
		return "", fmt.Errorf("os.MkdirAll: %w", err)
	}

	return cacheDir, nil
}
