// Package gitmodules retrieves and verifies Git module contents in the v1 cache.
package gitmodules

import (
	"context"
	"fmt"
	"path/filepath"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// Cache owns Git checkouts, installed contents, and their layout below EASYPPATH.
type Cache struct{ root string }

// New places the v1 Git cache below the supplied EasyP storage directory.
func New(storageDir string) *Cache {
	return &Cache{root: filepath.Join(storageDir, "v1", "git")}
}

// Cached reads installed metadata without downloading or changing files.
// Call Install first to verify contents against the lock.
func (c *Cache) Cached(entry v1.LockedModule) (string, v1.Module, error) {
	directory := v1ModuleCachePath(c.root, entry)
	module, err := moduleconfig.ReadGitDependency(directory, entry.Source)
	if err != nil {
		return "", v1.Module{}, fmt.Errorf("ReadGitDependency: %w", err)
	}
	return directory, module, nil
}

// Versions lists semantic versions available for a module, including nested modules.
func (c *Cache) Versions(ctx context.Context, source string) ([]string, error) {
	return listV1ModuleTags(ctx, source)
}

// ValidateSource checks whether a module identity can identify a Git repository.
func ValidateSource(source string) error {
	_, err := v1GitModuleCandidates(source)
	return err
}
