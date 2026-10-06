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

	"golang.org/x/mod/sumdb/dirhash"
)

// A historical lock can describe a whole installed tree or released v0's proto
// archive. Each proof must match independently at the same pinned checkout.
func verifyMigrationLegacyHash(ctx context.Context, checkout v1ModuleCheckout, source, expectedHash string, tracked migrationFiles) error {
	var treeHash string
	var treeErr error
	if !tracked.omittedAuxiliarySymlinks {
		treeHash, treeErr = hashMigrationLegacyFiles(checkout.dir, tracked.regularFiles)
		if treeErr != nil {
			treeErr = fmt.Errorf("hashMigrationLegacyFiles: %w", treeErr)
		}
		// Initial resolution still validates the legacy whole-tree layout.
		if expectedHash == "" {
			return treeErr
		}
		if treeErr == nil && treeHash == expectedHash {
			return nil
		}
	}

	// Dropping links cannot prove a whole-tree hash. Archive verification also
	// checks proto path/content equivalence when there is no historical pin.
	// A pinned archive may independently succeed despite non-proto collisions
	// in the alternate whole-tree representation.
	archiveHash, err := hashMigrationProtoArchive(ctx, checkout.dir, tracked.regularFiles)
	if err != nil {
		return errors.Join(treeErr, fmt.Errorf("hashMigrationProtoArchive: %w", err))
	}
	if expectedHash == "" || archiveHash == expectedHash {
		return nil
	}

	candidates := "archive " + archiveHash
	if !tracked.omittedAuxiliarySymlinks {
		candidates += " or whole-tree " + treeHash
	}
	mismatch := fmt.Errorf("legacy hash mismatch for %s at %s: got %s, want %s", source, checkout.commit, candidates, expectedHash)
	return errors.Join(treeErr, mismatch)
}

// hashMigrationLegacyFiles calculates the whole-tree legacy hash without
// creating an installed tree. All tracked files and their renamed directory
// nodes participate. Released v0 installers instead used the filtered Git
// archive reproduced by hashMigrationProtoArchive.
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
