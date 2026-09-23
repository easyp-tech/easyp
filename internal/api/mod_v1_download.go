package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/sumdb/dirhash"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func v1ModuleCachePath(cacheRoot string, entry v1.LockedModule) string {
	return filepath.Join(cacheRoot, "modules", v1CacheSourceKey(entry.Source), entry.Commit)
}

// downloadV1Lock installs each locked Git commit, verifying its content hash
// before accepting a downloaded or already cached module.
func downloadV1Lock(ctx context.Context, lock v1.Lock, cacheRoot string) error {
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
	defer os.RemoveAll(checkout)
	commit, err := gitV1(ctx, checkout, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	if strings.TrimSpace(commit) != entry.Commit {
		return fmt.Errorf("%s: checked out commit does not match lock", entry.Source)
	}
	filesRaw, err := gitV1(ctx, checkout, "ls-files", "-z")
	if err != nil {
		return err
	}
	files := strings.FieldsFunc(filesRaw, func(r rune) bool { return r == 0 })
	actual, err := dirhash.Hash1(files, func(name string) (io.ReadCloser, error) {
		return os.Open(filepath.Join(checkout, filepath.FromSlash(name)))
	})
	if err != nil {
		return err
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
	defer os.RemoveAll(stage)
	for _, name := range files {
		clean := filepath.Clean(filepath.FromSlash(name))
		if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return fmt.Errorf("invalid tracked file path %q", name)
		}
		sourcePath := filepath.Join(checkout, clean)
		info, err := os.Lstat(sourcePath)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported non-regular file %q", name)
		}
		destination := filepath.Join(stage, clean)
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return err
		}
		in, err := os.Open(sourcePath)
		if err != nil {
			return err
		}
		out, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
		if err != nil {
			_ = in.Close()
			return err
		}
		_, copyErr := io.Copy(out, in)
		closeErr := out.Close()
		_ = in.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	if err := os.Rename(stage, installed); err != nil {
		return fmt.Errorf("install %s@%s: %w", entry.Source, entry.Commit, err)
	}
	return nil
}
