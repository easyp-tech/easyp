package gitmodules

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// retainFetchedSnapshot publishes the bytes just hashed by Fetch. Install can
// reuse this complete snapshot instead of materializing the same commit again.
// Provisional root inspections never reach this path.
func (c *Cache) retainFetchedSnapshot(ctx context.Context, checkout v1ModuleCheckout, entry v1.LockedModule, files []string) error {
	installed := v1ModuleCachePath(c.root, entry)
	snapshot := c.snapshotState(installed)
	if err := snapshot.lock(ctx); err != nil {
		return fmt.Errorf("lock: %w", err)
	}
	defer snapshot.unlock()
	if err := os.MkdirAll(filepath.Dir(installed), 0o755); err != nil {
		return fmt.Errorf("MkdirAll: %w", err)
	}
	unlock, err := lockObjectRepository(ctx, installed+".install.lock")
	if err != nil {
		return fmt.Errorf("lockObjectRepository: %w", err)
	}
	defer unlock()
	info, err := os.Lstat(installed)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("cached path %q is not a directory", installed)
		}
		if !snapshot.verified || !os.SameFile(snapshot.identity, info) {
			snapshot.verified = false
			files, err = verifiedSnapshotFiles(installed, entry, true)
			if err != nil {
				return fmt.Errorf("verifiedSnapshotFiles: %w", err)
			}
		}
	} else if errors.Is(err, os.ErrNotExist) {
		if err := retainFetchedSource(ctx, c.root, checkout.dir, entry); err != nil {
			return fmt.Errorf("retainFetchedSource: %w", err)
		}
		if err := os.Rename(checkout.snapshot, installed); err != nil {
			return fmt.Errorf("Rename: %w", err)
		}
		return snapshot.accept(installed, files)
	} else {
		return fmt.Errorf("Lstat: %w", err)
	}
	snapshot.files, snapshot.identity, snapshot.verified = files, info, true
	return nil
}

// retainFetchedSource copies already fetched objects locally. HEAD/tag Fetch
// must not download the revision a second time just to preserve policy objects.
func retainFetchedSource(ctx context.Context, root, checkout string, entry v1.LockedModule) error {
	absoluteCheckout, err := filepath.Abs(checkout)
	if err != nil {
		return fmt.Errorf("Abs: %w", err)
	}
	checkout = absoluteCheckout
	raw, err := gitV1(ctx, checkout, "config", "--get", "remote.origin.url")
	if err != nil {
		return fmt.Errorf("gitV1: %w", err)
	}
	origin := strings.TrimSpace(raw)
	objectsRoot, err := filepath.Abs(filepath.Join(root, "objects"))
	if err != nil {
		return fmt.Errorf("Abs: %w", err)
	}
	relative, relErr := filepath.Rel(objectsRoot, origin)
	if relErr == nil && filepath.IsLocal(relative) && !strings.ContainsAny(relative, "/\\") {
		return writeSourceBindingRepository(root, origin, entry)
	}
	if err := os.MkdirAll(objectsRoot, 0o700); err != nil {
		return fmt.Errorf("MkdirAll: %w", err)
	}
	repository := filepath.Join(objectsRoot, v1CacheSourceKey(origin))
	unlock, err := lockObjectRepository(ctx, repository+".lock")
	if err != nil {
		return fmt.Errorf("lockObjectRepository: %w", err)
	}
	defer unlock()
	if _, err := os.Stat(filepath.Join(repository, "HEAD")); errors.Is(err, os.ErrNotExist) {
		if _, err := gitV1(ctx, "", "clone", "--quiet", "--bare", "--no-hardlinks", "--", checkout, repository); err != nil {
			return fmt.Errorf("gitV1: %w", err)
		}
		if _, err := gitV1(ctx, repository, "config", "gc.auto", "0"); err != nil {
			return fmt.Errorf("gitV1: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("Stat: %w", err)
	} else {
		if _, err := gitV1(ctx, repository, "fetch", "--quiet", "--no-tags", "--no-write-fetch-head", "--depth=1", "--", checkout, entry.Commit); err != nil {
			return fmt.Errorf("gitV1: %w", err)
		}
	}
	if _, err := gitV1(ctx, repository, "update-ref", "refs/easyp/"+strings.ToLower(entry.Commit), entry.Commit); err != nil {
		return fmt.Errorf("gitV1: %w", err)
	}
	return writeSourceBindingRepository(root, repository, entry)
}
