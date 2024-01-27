package mod

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

const (
	cacheDirPerm = 0755
)

// CreateCacheDir create and return path to cache dir.
// cache dir contains repo cache for repository with proto files.
// cmd/go/internal/modfetch/codehost/codehost.go: 228 - create workdir
func CreateCacheDir(name string) (string, error) {
	tmpDir := os.TempDir()
	cacheDir := filepath.Join(tmpDir, fmt.Sprintf("%x", sha256.Sum256([]byte(name))))

	if err := os.MkdirAll(cacheDir, cacheDirPerm); err != nil {
		return "", fmt.Errorf("os.MkdirAll: %w", err)
	}

	return cacheDir, nil
}
