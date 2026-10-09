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
func verifyMigrationLegacyHash(ctx context.Context, checkout v1ModuleCheckout, tracked migrationFiles, request migrationRequest) error {
	if request.legacyHash == "" {
		return validateMigrationInitialSelection(ctx, checkout, tracked)
	}
	var treeHash string
	var treeErr error
	if !tracked.hasSymlinks {
		// Raw committed bytes are only a whole-tree proof when their recorded
		// digest matches; no host checkout/filter representation is assumed.
		treeHash, treeErr = migrationPinnedTreeHash(ctx, checkout, tracked.regularFiles)
		if treeErr != nil {
			treeErr = fmt.Errorf("hashMigrationLegacyFiles: %w", treeErr)
		}
		if treeErr == nil && treeHash == request.legacyHash {
			return validateMigrationSelection(checkout.snapshot, tracked.regularFiles, checkout.module)
		}
	}

	// Dropping links cannot prove a whole-tree hash. Archive verification also
	// checks proto path/content equivalence when there is no historical pin.
	// A pinned archive may independently succeed despite non-proto collisions
	// in the alternate whole-tree representation.
	files, err := snapshotV1Files(checkout.snapshot)
	if err != nil {
		return errors.Join(treeErr, fmt.Errorf("snapshotV1Files: %w", err))
	}
	roots, err := readMigrationLegacyRoots(checkout.snapshot, files)
	if err != nil {
		return errors.Join(treeErr, fmt.Errorf("readMigrationLegacyRoots: %w", err))
	}
	selected, err := migrationSelectedFiles(checkout.snapshot, checkout.module)
	if err != nil {
		return errors.Join(treeErr, fmt.Errorf("migrationSelectedFiles: %w", err))
	}
	nodes, err := readMigrationProtoArchive(ctx, checkout.dir, checkout.commit, tracked.trackedFiles)
	if err != nil {
		return errors.Join(treeErr, fmt.Errorf("readMigrationProtoArchive: %w", err))
	}
	hashes, err := migrationArchiveHashes(ctx, nodes, roots, selected)
	if err != nil {
		return errors.Join(treeErr, fmt.Errorf("migrationArchiveHashes: %w", err))
	}
	if slices.Contains(hashes, request.legacyHash) {
		return nil
	}

	candidates := "archive " + strings.Join(hashes, " or archive ")
	if !tracked.hasSymlinks {
		candidates += " or whole-tree " + treeHash
	}
	mismatch := fmt.Errorf("legacy hash mismatch for %s at %s: got %s, want %s", request.source, checkout.commit, candidates, request.legacyHash)
	return errors.Join(treeErr, mismatch)
}

func validateMigrationInitialSelection(ctx context.Context, checkout v1ModuleCheckout, tracked migrationFiles) error {
	if migrationUsesLogicalAliases(checkout, tracked) {
		err := validateMigrationLogicalOwnership(ctx, checkout, tracked)
		if err != nil {
			return fmt.Errorf("validateMigrationLogicalOwnership: %w", err)
		}
	} else {
		err := validateMigrationSelection(checkout.snapshot, tracked.regularFiles, checkout.module)
		if err != nil {
			return fmt.Errorf("validateMigrationSelection: %w", err)
		}
	}
	roots, err := readMigrationLegacyRoots(checkout.snapshot, tracked.trackedFiles)
	if err != nil {
		return fmt.Errorf("readMigrationLegacyRoots: %w", err)
	}
	nodes, err := readMigrationSourceArchive(ctx, checkout.dir, checkout.commit, tracked.trackedFiles, roots)
	if err != nil {
		return fmt.Errorf("readMigrationProtoArchive: %w", err)
	}
	err = validateMigrationArchiveRegularSources(nodes, checkout.snapshot, tracked.regularFiles)
	if err != nil {
		return fmt.Errorf("validateMigrationArchiveRegularSources: %w", err)
	}
	return nil
}

// Whole-tree digests belong only to actual historical locks. Their bytes come
// from the pinned repository, independently of the smaller native snapshot.
func migrationPinnedTreeHash(ctx context.Context, checkout v1ModuleCheckout, files []string) (string, error) {
	roots, err := readMigrationLegacyRoots(checkout.snapshot, files)
	if err != nil {
		return "", fmt.Errorf("readMigrationLegacyRoots: %w", err)
	}
	view, err := sourceV1SnapshotView(ctx, checkout.dir, checkout.commit)
	if err != nil {
		return "", fmt.Errorf("sourceV1SnapshotView: %w", err)
	}
	return hashMigrationRenamedFiles(files, roots, func(name string) (io.ReadCloser, error) { return view.Open(ctx, name) })
}

// hashMigrationLegacyFiles calculates the whole-tree legacy hash without
// creating an installed tree. All tracked files and their renamed directory
// nodes participate. Released v0 installers instead used the filtered Git
// archive reproduced by hashMigrationProtoArchive.
func hashMigrationLegacyFiles(checkout string, files []string) (string, error) {
	for _, name := range files {
		if !filepath.IsLocal(filepath.FromSlash(name)) || path.Clean(name) != name || strings.Contains(name, "\\") {
			return "", fmt.Errorf("unsupported tracked path %q", name)
		}
		if _, err := regularV1File(filepath.Join(checkout, filepath.FromSlash(name))); err != nil {
			return "", fmt.Errorf("regularV1File: %w", err)
		}
	}
	roots, err := readMigrationLegacyRoots(checkout, files)
	if err != nil {
		return "", fmt.Errorf("readMigrationLegacyRoots: %w", err)
	}

	hash, err := hashMigrationRenamedFiles(files, roots, func(name string) (io.ReadCloser, error) {
		file, err := os.Open(filepath.Join(checkout, filepath.FromSlash(name)))
		if err != nil {
			return nil, fmt.Errorf("Open: %w", err)
		}
		return file, nil
	})
	if err != nil {
		return "", fmt.Errorf("hashMigrationRenamedFiles: %w", err)
	}
	return hash, nil
}

func hashMigrationRenamedFiles(files, roots []string, open func(string) (io.ReadCloser, error)) (string, error) {
	directories := make(map[string]bool)
	for _, name := range files {
		for directory := path.Dir(name); directory != "."; directory = path.Dir(directory) {
			directories[directory] = true
		}
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
		return open(originals[name])
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
