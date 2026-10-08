package modules

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

// TidyResult reports import edits that have been validated and committed.
type TidyResult struct {
	Imports []ImportRewrite
}

// ImportRewrite identifies one applied consumer import mapping and its pins.
// File is relative to the owning module and names the physical write target.
type ImportRewrite struct {
	File       string
	From       string
	To         string
	Module     string
	OldVersion string
	Version    string
	OldCommit  string
	Commit     string
	OldRoots   []string
	Roots      []string
}

// TidyWithReport resolves the manifest's requirements, checks and applies unique
// pinned import mappings, and returns their report only after successful commit.
// Local replacements only validate their ephemeral graph and produce no edits.
func TidyWithReport(ctx context.Context, root string, repository Repository) (_ TidyResult, resultErr error) {
	tx, err := newResolvedFilesTransaction(root)
	if err != nil {
		return TidyResult{}, fmt.Errorf("newResolvedFilesTransaction: %w", err)
	}
	defer func() { resultErr = errors.Join(resultErr, tx.close()) }()
	// All source comparisons use the absolute lexical spelling that was
	// opened by the transaction. Keep aliases for bounded identity rechecks.
	root = tx.requestedRoot
	boundaries, err := tidyCacheBoundaries(repository, v1.Lock{})
	if err != nil {
		return TidyResult{}, fmt.Errorf("tidyCacheBoundaries: %w", err)
	}
	err = checkTidyCacheOwnership(tx, boundaries)
	if err != nil {
		return TidyResult{}, fmt.Errorf("checkTidyCacheOwnership: %w", err)
	}
	original, module, err := ReadManifest(root)
	if err != nil {
		return TidyResult{}, fmt.Errorf("ReadManifest: %w", err)
	}
	if len(module.Replaces) > 0 {
		err = validateLocalOverlay(ctx, root, module, repository, false)
		return TidyResult{}, err
	}
	manifest, err := tx.capture(v1.ModuleFile, nil)
	if err != nil {
		return TidyResult{}, fmt.Errorf("capture: %w", err)
	}
	if !bytes.Equal(manifest.data, original) {
		return TidyResult{}, fmt.Errorf("protobuf.mod changed while reading: %w", sourceview.ErrChanged)
	}
	locked, err := tx.capture(v1.LockFile, nil)
	if err != nil {
		return TidyResult{}, fmt.Errorf("capture: %w", err)
	}
	var existing v1.Lock
	if locked.exists {
		existing, err = v1.ParseLock(bytes.NewReader(locked.data))
	} else {
		// Preserve the existing legacy-lock diagnostic on a missing native lock.
		_, err = ReadLock(filepath.Join(root, v1.LockFile))
		if errors.Is(err, os.ErrNotExist) {
			err = nil
		}
	}
	if err != nil {
		return TidyResult{}, fmt.Errorf("ParseLock: %w", err)
	}
	boundaries, err = tidyCacheBoundaries(repository, existing)
	if err != nil {
		return TidyResult{}, fmt.Errorf("tidyCacheBoundaries: %w", err)
	}
	err = checkTidyCacheOwnership(tx, boundaries)
	if err != nil {
		return TidyResult{}, fmt.Errorf("checkTidyCacheOwnership: %w", err)
	}
	own, err := ModuleSources(root, module)
	if err != nil {
		return TidyResult{}, fmt.Errorf("ModuleSources: %w", err)
	}
	own = excludeTidyCacheSources(own, boundaries)
	var files []string
	_, completeCacheOwnership := repository.(sourceCacheDirectories)
	if completeCacheOwnership {
		// Complete cache ownership permits an early consumer checkpoint. A
		// generic repository must first identify the new snapshot directories.
		files, err = captureTidySources(tx, own)
		if err != nil {
			return TidyResult{}, fmt.Errorf("captureTidySources: %w", err)
		}
	}
	resolved, err := resolveV1Graph(ctx, module, existing, repository, graphResolveRequest{
		preserveHeads: true,
		transitions:   namespaceTransitionsTidyRepairs,
	})
	if err != nil {
		return TidyResult{}, fmt.Errorf("resolveV1Graph: %w", err)
	}
	lock := resolved.lockFile()
	view := newTidySourceView(tx, own)
	if err := repository.Install(ctx, lock); err != nil {
		return TidyResult{}, fmt.Errorf("Install: %w", err)
	}
	view.retainPinnedSources(resolved)
	dependencies, err := view.cachedSources(ctx, lock, repository)
	if err != nil {
		return TidyResult{}, fmt.Errorf("cachedSources: %w", err)
	}
	nextBoundaries, err := tidyCacheBoundaries(repository, lock)
	if err != nil {
		return TidyResult{}, fmt.Errorf("tidyCacheBoundaries: %w", err)
	}
	boundaries = append(boundaries, nextBoundaries...)
	err = checkTidyCacheOwnership(tx, boundaries)
	if err != nil {
		return TidyResult{}, fmt.Errorf("checkTidyCacheOwnership: %w", err)
	}
	own = excludeTidyCacheSources(own, boundaries)
	if !completeCacheOwnership {
		files, err = captureTidySources(tx, own)
		if err != nil {
			return TidyResult{}, fmt.Errorf("captureTidySources: %w", err)
		}
	}
	needed, err := capturedTidyImports(tx, files)
	if err != nil {
		return TidyResult{}, fmt.Errorf("capturedTidyImports: %w", err)
	}
	planner := newTidyImportPlanner(resolved, existing, repository)
	bindings, err := planner.tidyImportBindings(ctx, needed, view)
	if err != nil {
		return TidyResult{}, fmt.Errorf("tidyImportBindings: %w", err)
	}
	// Proof of an old pin can acquire a previously missing snapshot. Those
	// newly known directories are excluded from subsequent consumer rechecks.
	oldBoundaries, err := tidyCacheBoundaries(repository, existing)
	if err != nil {
		return TidyResult{}, fmt.Errorf("tidyCacheBoundaries: %w", err)
	}
	boundaries = append(boundaries, oldBoundaries...)
	own = excludeTidyCacheSources(own, boundaries)
	roots := append(slices.Clone(own), dependencies...)
	if err := CheckSourceCollisions(roots); err != nil {
		return TidyResult{}, fmt.Errorf("CheckSourceCollisions: %w", err)
	}
	view.roots = roots
	verifyCache := tidyInputVerifier(ctx, lock, repository)
	tx.validateInputs = func() error {
		err := checkTidySourceSelection(tx, own, files)
		if err != nil {
			return fmt.Errorf("checkTidySourceSelection: %w", err)
		}
		err = view.verifyOldNamespaces(ctx, repository)
		if err != nil {
			return fmt.Errorf("verifyOldNamespaces: %w", err)
		}
		return verifyCache()
	}
	report, err := planTidyImportRewrites(tx, files, bindings, view)
	if err != nil {
		return TidyResult{}, fmt.Errorf("planTidyImportRewrites: %w", err)
	}
	owners, err := validateTidyImports(ctx, files, report, view)
	if err != nil {
		return TidyResult{}, fmt.Errorf("validateTidyImports: %w", err)
	}
	updated, err := augmentV1ManifestRequirementsWithOwners(original, module, lock, repository, owners)
	if err != nil {
		return TidyResult{}, fmt.Errorf("augmentV1ManifestRequirementsWithOwners: %w", err)
	}
	err = ctx.Err()
	if err != nil {
		return TidyResult{}, fmt.Errorf("Err: %w", err)
	}
	err = writeV1ResolvedFilesTransaction(tx, updated, lock)
	if err != nil {
		return TidyResult{}, fmt.Errorf("writeV1ResolvedFilesTransaction: %w", err)
	}
	return report, nil
}

func captureTidySources(tx *resolvedFilesTransaction, roots SourceRoots) ([]string, error) {
	files, err := tidySourceFiles(tx, roots)
	if err != nil {
		return nil, fmt.Errorf("tidySourceFiles: %w", err)
	}
	for _, name := range files {
		_, err := tx.capture(name, roots)
		if err != nil {
			return nil, fmt.Errorf("capture: %w", err)
		}
	}
	return files, nil
}

func checkTidySourceSelection(tx *resolvedFilesTransaction, roots SourceRoots, expected []string) error {
	current, err := tidySourceFiles(tx, roots)
	if err != nil {
		return fmt.Errorf("tidySourceFiles: %w", err)
	}
	if !slices.Equal(current, expected) {
		return fmt.Errorf("owned consumer source selection changed since planning in %q: %w", tx.requestedRoot, sourceview.ErrChanged)
	}
	return nil
}

func tidySourceFiles(tx *resolvedFilesTransaction, roots SourceRoots) ([]string, error) {
	seen := make(map[string]bool)
	var files []string
	for _, root := range roots {
		err := roots.WalkSelected(root, roots.FileAllowed(), func(path string) error {
			name, err := filepath.Rel(tx.requestedRoot, path)
			if err != nil {
				return fmt.Errorf("Rel: %w", err)
			}
			if seen[name] {
				return nil
			}
			seen[name] = true
			files = append(files, name)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("WalkSelected: %w", err)
		}
	}
	slices.Sort(files)
	return files, nil
}

func capturedTidyImports(tx *resolvedFilesTransaction, files []string) (map[string]v1UnresolvedImport, error) {
	needed := make(map[string]v1UnresolvedImport)
	for _, file := range files {
		imports, err := ParseProtoImports(file, tx.expected[file].data)
		if err != nil {
			return nil, fmt.Errorf("ParseProtoImports: %w", err)
		}
		for _, name := range imports {
			if _, exists := needed[name]; !exists {
				needed[name] = v1UnresolvedImport{owner: file, path: name}
			}
		}
	}
	return needed, nil
}
