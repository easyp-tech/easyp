package gitmodules

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/mod/semver"
	"golang.org/x/mod/sumdb/dirhash"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

// FetchMigration verifies a legacy installed-tree hash before calculating the
// v1 tracked-checkout hash. Legacy repositories must preserve their proto source
// selection and import names. An empty version requires an empty legacyHash and
// permits initial resolution; native v1 repositories need no legacy comparison.
func (c *Cache) FetchMigration(ctx context.Context, source, version, legacyHash string) (fetched modules.Fetched, err error) {
	if (version != "" || legacyHash != "") && !v1.IsCommitRef(version) && !semver.IsValid(version) {
		return modules.Fetched{}, fmt.Errorf("migration version %q must be a full Git commit or SemVer tag", version)
	}
	checkout, err := checkoutV1Module(ctx, source, version, c.root)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("checkoutV1Module: %w", err)
	}
	defer func() {
		removeErr := os.RemoveAll(checkout.dir)
		if removeErr != nil {
			fetched = modules.Fetched{}
			err = errors.Join(err, fmt.Errorf("RemoveAll: %w", removeErr))
		}
	}()
	if err := moduleconfig.ValidateLegacyMajor(checkout.dir, source, version); err != nil {
		return modules.Fetched{}, fmt.Errorf("ValidateLegacyMajor: %w", err)
	}
	files, err := migrationTrackedFiles(ctx, checkout.dir)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("migrationTrackedFiles: %w", err)
	}
	if err := validateMigrationLegacyReplacements(checkout.dir, files); err != nil {
		return modules.Fetched{}, fmt.Errorf("validateMigrationLegacyReplacements: %w", err)
	}
	native, err := hasNativeMigrationModule(checkout.dir, files, source)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("hasNativeMigrationModule: %w", err)
	}
	if !native || legacyHash != "" {
		oldHash, err := hashMigrationLegacyFiles(checkout.dir, files)
		if err != nil {
			return modules.Fetched{}, fmt.Errorf("hashMigrationLegacyFiles: %w", err)
		}
		if legacyHash != "" && oldHash != legacyHash {
			return modules.Fetched{}, fmt.Errorf("legacy hash mismatch for %s at %s: got %s, want %s", source, checkout.commit, oldHash, legacyHash)
		}
		if err := validateMigrationSelection(checkout.dir, files, checkout.module); err != nil {
			return modules.Fetched{}, fmt.Errorf("validateMigrationSelection: %w", err)
		}
	}
	files = selectV1ProtoFiles(files, checkout.module.ProtoFilters)
	hash, err := hashV1Files(checkout.dir, files)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("hashV1Files: %w", err)
	}
	if version == "" {
		version = checkout.commit
	}
	module, bindings, err := c.resolveBSRDependencies(ctx, checkout.module)
	if err != nil {
		return modules.Fetched{}, fmt.Errorf("resolveBSRDependencies: %w", err)
	}
	return modules.Fetched{Module: module, Lock: v1.LockedModule{
		Source: source, Version: version, Commit: checkout.commit, Hash: hash, BSR: bindings,
	}}, nil
}

func migrationTrackedFiles(ctx context.Context, checkout string) ([]string, error) {
	files, err := trackedV1Files(ctx, checkout)
	if err != nil {
		return nil, fmt.Errorf("trackedV1Files: %w", err)
	}
	// A Git symlink may be checked out as a regular file on systems that do
	// not support symlinks. Inspect Git modes as well as filesystem modes.
	raw, err := gitV1(ctx, checkout, "ls-files", "--stage", "-z")
	if err != nil {
		return nil, fmt.Errorf("gitV1: %w", err)
	}
	for _, entry := range strings.Split(strings.TrimSuffix(raw, "\x00"), "\x00") {
		if entry == "" {
			continue
		}
		if !strings.HasPrefix(entry, "100644 ") && !strings.HasPrefix(entry, "100755 ") {
			return nil, fmt.Errorf("unsupported non-regular Git mode in %q", entry)
		}
	}
	return files, nil
}

// hashMigrationLegacyFiles reproduces the old installer's path rewrite without
// creating an installed tree. All tracked files participate, including files
// outside import roots and non-proto files. Directory nodes also participate in
// collision detection because buildInstallTree created them before hashing.
func hashMigrationLegacyFiles(checkout string, files []string) (string, error) {
	directories := make(map[string]bool)
	for _, name := range files {
		if !filepath.IsLocal(filepath.FromSlash(name)) || path.Clean(name) != name || strings.Contains(name, "\\") {
			return "", fmt.Errorf("unsupported tracked path %q", name)
		}
		if _, err := regularV1File(filepath.Join(checkout, filepath.FromSlash(name))); err != nil {
			return "", fmt.Errorf("regularV1File: %w", err)
		}
		for directory := path.Dir(name); directory != "."; directory = path.Dir(directory) {
			directories[directory] = true
		}
	}
	roots, err := readMigrationLegacyRoots(checkout, files)
	if err != nil {
		return "", fmt.Errorf("readMigrationLegacyRoots: %w", err)
	}

	installedDirectories := make(map[string]bool)
	for directory := range directories {
		for destination := renameMigrationLegacyFile(directory, roots); destination != "."; destination = path.Dir(destination) {
			installedDirectories[destination] = true
		}
	}
	originals := make(map[string]string, len(files))
	names := make([]string, 0, len(files))
	for _, original := range files {
		name := renameMigrationLegacyFile(original, roots)
		if other, exists := originals[name]; exists {
			return "", fmt.Errorf("legacy file collision at %q between %q and %q", name, other, original)
		}
		originals[name] = original
		names = append(names, name)
		for directory := path.Dir(name); directory != "."; directory = path.Dir(directory) {
			installedDirectories[directory] = true
		}
	}
	slices.Sort(names)
	for _, name := range names {
		if installedDirectories[name] {
			return "", fmt.Errorf("legacy directory/file collision at %q from %q", name, originals[name])
		}
	}
	hash, err := dirhash.Hash1(names, func(name string) (io.ReadCloser, error) {
		file, err := os.Open(filepath.Join(checkout, filepath.FromSlash(originals[name])))
		if err != nil {
			return nil, fmt.Errorf("Open: %w", err)
		}
		return file, nil
	})
	if err != nil {
		return "", fmt.Errorf("Hash1: %w", err)
	}
	return hash, nil
}

// The old renamer selected the first matching prefix, without sorting,
// cleaning, or repeatedly stripping roots from the resulting path.
func renameMigrationLegacyFile(name string, roots []string) string {
	for _, root := range roots {
		prefix := root + "/"
		if strings.HasPrefix(name, prefix) {
			return strings.TrimPrefix(name, prefix)
		}
	}
	return name
}
