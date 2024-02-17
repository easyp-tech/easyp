package dirs

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

// CacheDir create and return path to cache dir.
// cache dir contains repository cache for repository with proto files.
// cmd/go/internal/modfetch/codehost/codehost.go: 228 - create workdir
func (d *Dirs) CacheDir(name string) (string, error) {
	cacheDir := filepath.Join(d.cacheRootDir, fmt.Sprintf("%x", sha256.Sum256([]byte(name))))

	if err := os.MkdirAll(cacheDir, cacheDirPerm); err != nil {
		return "", fmt.Errorf("os.MkdirAll: %w", err)
	}

	return cacheDir, nil
}
