package gitmodules

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"slices"
	"strings"

	"golang.org/x/mod/sumdb/dirhash"

	"github.com/easyp-tech/easyp/internal/sourceview"
)

// migrationLegacyLayout separates the historical installed namespace from the
// consumer's root/subdirectory. Every mapping binds equivalent pinned bytes.
type migrationLegacyLayout struct {
	files map[string]string
}

func proveMigrationLegacyLayout(ctx context.Context, checkout v1ModuleCheckout, source, expectedHash string, tracked migrationFiles, native, selection bool) (migrationLegacyLayout, error) {
	nativeInitial := native && expectedHash == ""
	if nativeInitial && !selection {
		layout := migrationLegacyLayout{files: make(map[string]string)}
		for _, file := range checkout.inspection.Files {
			layout.files[file.Path] = file.Path
		}
		return layout, nil
	}
	readRoots := readMigrationLegacyRoots
	if nativeInitial {
		readRoots = readMigrationArchiveRoots
	}
	roots, err := readRoots(checkout.snapshot, tracked.trackedFiles)
	if err != nil {
		return migrationLegacyLayout{}, fmt.Errorf("readRoots: %w", err)
	}
	var treeHash string
	var treeErr error
	view, err := sourceV1SnapshotView(ctx, checkout.dir, checkout.commit)
	if err != nil {
		return migrationLegacyLayout{}, fmt.Errorf("sourceV1SnapshotView: %w", err)
	}
	if !tracked.hasSymlinks && !nativeInitial {
		treeHash, treeErr = hashMigrationRenamedFiles(tracked.regularFiles, roots, func(name string) (io.ReadCloser, error) { return view.Open(ctx, name) })
		if treeErr == nil && expectedHash != "" && treeHash == expectedHash {
			return migrationRegularLayout(tracked.regularFiles, roots), nil
		}
		if expectedHash == "" && treeErr != nil {
			return migrationLegacyLayout{}, fmt.Errorf("hashMigrationRenamedFiles: %w", treeErr)
		}
	}
	nodes, err := readMigrationProtoArchive(ctx, checkout.dir, checkout.commit, tracked.trackedFiles)
	if err != nil {
		return migrationLegacyLayout{}, fmt.Errorf("readMigrationProtoArchive: %w", err)
	}
	if err := validateMigrationArchiveCoverage(nodes, tracked.regularFiles); err != nil {
		return migrationLegacyLayout{}, fmt.Errorf("validateMigrationArchiveCoverage: %w", err)
	}
	var hashes []string
	var failures []error
	for _, rewrite := range []bool{false, true} {
		installed, names, err := installMigrationArchive(nodes, roots, rewrite)
		if err != nil {
			failures = append(failures, fmt.Errorf("installMigrationArchive: %w", err))
			continue
		}
		legacyView := sourceview.New(installed)
		layout, err := migrationArchiveLayout(ctx, legacyView, view, nodes, roots)
		if err != nil {
			failures = append(failures, fmt.Errorf("migrationArchiveLayout: %w", err))
			continue
		}
		hash, err := dirhash.Hash1(names, func(name string) (io.ReadCloser, error) { return legacyView.Open(ctx, name) })
		if err != nil {
			failures = append(failures, fmt.Errorf("Hash1: %w", err))
			continue
		}
		if expectedHash == "" || hash == expectedHash {
			return layout, nil
		}
		if !slices.Contains(hashes, hash) {
			hashes = append(hashes, hash)
		}
	}
	if len(hashes) == 0 {
		return migrationLegacyLayout{}, errors.Join(failures...)
	}
	candidates := "archive " + strings.Join(hashes, " or archive ")
	if !tracked.hasSymlinks {
		candidates += " or whole-tree " + treeHash
	}
	return migrationLegacyLayout{}, errors.Join(treeErr, fmt.Errorf("legacy hash mismatch for %s at %s: got %s, want %s", source, checkout.commit, candidates, expectedHash))
}

func validateMigrationArchiveCoverage(nodes []migrationArchiveNode, files []string) error {
	archived := make(map[string]bool, len(nodes))
	for _, node := range nodes {
		if !node.mode.IsDir() {
			archived[node.name] = true
		}
	}
	for _, file := range files {
		if path.Ext(file) == ".proto" && !archived[file] {
			return fmt.Errorf("legacy archive source selection omits %q; manual migration is required", file)
		}
	}
	return nil
}

func migrationRegularLayout(files, roots []string) migrationLegacyLayout {
	layout := migrationLegacyLayout{files: make(map[string]string)}
	for _, file := range files {
		if path.Ext(file) != ".proto" {
			continue
		}
		name := renameMigrationLegacyFile(file, roots)
		layout.files[name] = file
	}
	return layout
}

func migrationArchiveLayout(ctx context.Context, legacy, current *sourceview.View, nodes []migrationArchiveNode, roots []string) (migrationLegacyLayout, error) {
	layout := migrationLegacyLayout{files: make(map[string]string)}
	for _, node := range nodes {
		if node.mode.IsDir() || path.Ext(node.name) != ".proto" {
			continue
		}
		name := renameMigrationLegacyFile(node.name, roots)
		old, err := readMigrationView(ctx, legacy, name)
		if err != nil {
			return migrationLegacyLayout{}, fmt.Errorf("readMigrationView: %w", err)
		}
		data, err := readMigrationView(ctx, current, node.name)
		if err != nil {
			return migrationLegacyLayout{}, fmt.Errorf("readMigrationView: %w", err)
		}
		if !bytes.Equal(old, data) {
			return migrationLegacyLayout{}, fmt.Errorf("legacy archive source selection changes %q contents; manual migration is required", node.name)
		}
		layout.files[name] = node.name
	}
	return layout, nil
}

func readMigrationView(ctx context.Context, view *sourceview.View, name string) ([]byte, error) {
	file, err := view.Open(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("Open: %s: %w", name, err)
	}
	data, readErr := io.ReadAll(file)
	closeErr := file.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		return nil, fmt.Errorf("ReadAll: %w", err)
	}
	return data, nil
}
