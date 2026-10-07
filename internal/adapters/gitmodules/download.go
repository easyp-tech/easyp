package gitmodules

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func v1ModuleCachePath(cacheRoot string, entry v1.LockedModule) string {
	identity := entry.Commit + ":" + entry.Hash
	if roots := v1RootSelectionKey(entry.Roots); roots != "" {
		identity += ":roots:" + roots
	}
	return filepath.Join(cacheRoot, "modules", v1CacheSourceKey(entry.Source), v1CacheSourceKey(identity))
}

// Install installs each locked Git commit, verifying its content hash
// before accepting a downloaded or already cached module.
func (c *Cache) Install(ctx context.Context, lock v1.Lock) error {
	return c.install(ctx, lock, true)
}

// VerifyCached verifies already installed entries without downloading or writing.
// Missing content must be acquired by an explicit mod download first.
func (c *Cache) VerifyCached(ctx context.Context, lock v1.Lock) error {
	return c.install(ctx, lock, false)
}

func (c *Cache) install(ctx context.Context, lock v1.Lock, acquire bool) error {
	if err := lock.Validate(); err != nil {
		return fmt.Errorf("Validate: %w", err)
	}
	for _, entry := range lock.Modules {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("Err: %w", err)
		}
		if err := c.installEntry(ctx, entry, acquire); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cache) installEntry(ctx context.Context, entry v1.LockedModule, acquire bool) error {
	installed := v1ModuleCachePath(c.root, entry)
	info, err := os.Lstat(installed)
	if errors.Is(err, os.ErrNotExist) {
		if !acquire {
			return fmt.Errorf("policy module %s@%s is not cached; run easyp mod download before validate-config: %w", entry.Source, entry.Commit, err)
		}
		if err := os.MkdirAll(filepath.Dir(installed), 0o755); err != nil {
			return fmt.Errorf("MkdirAll: %w", err)
		}
		unlock, lockErr := lockObjectRepository(ctx, installed+".install.lock")
		if lockErr != nil {
			return fmt.Errorf("lockObjectRepository: %w", lockErr)
		}
		defer unlock()

		// Another process may have completed the same atomic install while this
		// process waited for the per-snapshot lock. Reuse and verify that result
		// rather than racing a second rename into the same cache directory.
		info, err = os.Lstat(installed)
		if errors.Is(err, os.ErrNotExist) {
			if err := fetchPinnedV1Module(ctx, entry, c.root, installed); err != nil {
				return err
			}
			info, err = os.Lstat(installed)
		}
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s: cached path %q is not a directory; inspect this path and keep protobuf.lock unchanged", entry.Source, installed)
	}
	if err := verifyInstalledV1Module(installed, entry); err != nil {
		return fmt.Errorf("verify cached %s at %q: %w", entry.Source, installed, err)
	}
	if err := moduleconfig.ValidateLegacyMajor(installed, entry.Source, entry.Version); err != nil {
		return fmt.Errorf("ValidateLegacyMajor: %w", err)
	}
	return nil
}
