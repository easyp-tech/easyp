package gitmodules

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/sumdb/dirhash"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	disk "github.com/easyp-tech/easyp/internal/fs/fs"
)

func v1ModuleCachePath(cacheRoot string, entry v1.LockedModule) string {
	return filepath.Join(cacheRoot, "modules", v1CacheSourceKey(entry.Source), entry.Commit)
}

// Install installs each locked Git commit, verifying its content hash
// before accepting a downloaded or already cached module.
func (c *Cache) Install(ctx context.Context, lock v1.Lock) error {
	cacheRoot := c.root
	if err := lock.Validate(); err != nil {
		return fmt.Errorf("Validate: %w", err)
	}
	for _, entry := range lock.Modules {
		installed := v1ModuleCachePath(cacheRoot, entry)
		info, err := os.Lstat(installed)
		if errors.Is(err, os.ErrNotExist) {
			if err := fetchPinnedV1Module(ctx, entry, cacheRoot, installed); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("%s: cached path is not a directory", entry.Source)
		}
		actual, err := dirhash.HashDir(installed, "", dirhash.Hash1)
		if err != nil {
			return fmt.Errorf("verify cached %s: %w", entry.Source, err)
		}
		if actual != entry.Hash {
			return fmt.Errorf("cached %s@%s hash mismatch: got %s, want %s", entry.Source, entry.Commit, actual, entry.Hash)
		}
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
	files, err := trackedV1Files(ctx, checkout)
	if err != nil {
		return fmt.Errorf("trackedV1Files: %w", err)
	}
	actual, err := hashV1Files(checkout, files)
	if err != nil {
		return fmt.Errorf("hashV1Files: %w", err)
	}
	if actual != entry.Hash {
		return fmt.Errorf("%s@%s hash mismatch: got %s, want %s", entry.Source, entry.Commit, actual, entry.Hash)
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
