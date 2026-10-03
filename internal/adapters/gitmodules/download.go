package gitmodules

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	disk "github.com/easyp-tech/easyp/internal/fs/fs"
)

func v1ModuleCachePath(cacheRoot string, entry v1.LockedModule) string {
	return filepath.Join(cacheRoot, "modules", v1CacheSourceKey(entry.Source), entry.Commit)
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

func fetchPinnedV1Module(ctx context.Context, entry v1.LockedModule, cacheRoot, installed string) error {
	if err := os.MkdirAll(cacheRoot, 0o755); err != nil {
		return err
	}
	checkout, err := clonePinnedV1GitModule(ctx, entry, cacheRoot)
	if err != nil {
		return fmt.Errorf("clonePinnedV1GitModule: %w", err)
	}
	defer func() { _ = os.RemoveAll(checkout) }()
	commit, err := gitV1(ctx, checkout, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	if strings.TrimSpace(commit) != entry.Commit {
		return fmt.Errorf("%s: checked out commit does not match lock", entry.Source)
	}
	module, err := readCachedV1Module(checkout, entry.Source)
	if err != nil {
		return fmt.Errorf("readCachedV1Module: %w", err)
	}
	files, err := trackedV1Files(ctx, checkout)
	if err != nil {
		return fmt.Errorf("trackedV1Files: %w", err)
	}
	files = selectV1ProtoFiles(files, module.ProtoFilters)
	actual, err := hashV1Files(checkout, files)
	if err != nil {
		return fmt.Errorf("hashV1Files: %w", err)
	}
	if actual != entry.Hash {
		return fmt.Errorf("%s@%s hash mismatch: got %s, want %s; downloaded contents do not match the pinned hash; investigate repository integrity and keep protobuf.lock unchanged", entry.Source, entry.Commit, actual, entry.Hash)
	}
	if err := moduleconfig.ValidateLegacyMajor(checkout, entry.Source, entry.Version); err != nil {
		return fmt.Errorf("ValidateLegacyMajor: %w", err)
	}
	parent := filepath.Dir(installed)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(parent, "install-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(stage) }()
	if len(module.ProtoFilters) > 0 {
		// A valid Buf root may become empty after includes/excludes are applied.
		for _, root := range module.Roots {
			info, err := os.Lstat(filepath.Join(checkout, root))
			if err != nil {
				return fmt.Errorf("Lstat: %w", err)
			}
			if !info.IsDir() {
				return fmt.Errorf("module %s has invalid root %q: not a directory", module.Name, root)
			}
			if err := os.MkdirAll(filepath.Join(stage, root), 0o755); err != nil {
				return fmt.Errorf("MkdirAll: %w", err)
			}
		}
	}
	for _, name := range files {
		path := filepath.FromSlash(name)
		if err := disk.CopyRegularFile(filepath.Join(checkout, path), filepath.Join(stage, path)); err != nil {
			return fmt.Errorf("CopyRegularFile: %w", err)
		}
	}
	if err := os.Rename(stage, installed); err != nil {
		return fmt.Errorf("install %s@%s: %w", entry.Source, entry.Commit, err)
	}
	return nil
}
