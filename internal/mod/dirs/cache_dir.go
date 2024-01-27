package dirs

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/mod/dependency"
)

// CacheDir create and return path to cache dir.
// cache dir contains repo cache for repository with proto files.
// cmd/go/internal/modfetch/codehost/codehost.go: 228 - create workdir
func (d *Dirs) CacheDir(dep dependency.Dependency) (string, error) {
	key := dep.Name + ":" + dep.Version
	cacheDir := filepath.Join(d.cacheRootDir, fmt.Sprintf("%x", sha256.Sum256([]byte(key))))

	if err := os.MkdirAll(cacheDir, cacheDirPerm); err != nil {
		return "", fmt.Errorf("os.MkdirAll: %w", err)
	}

	return cacheDir, nil
}
