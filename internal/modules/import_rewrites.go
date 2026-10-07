package modules

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

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
	files, err := captureTidySources(tx, own)
	if err != nil {
		return TidyResult{}, fmt.Errorf("captureTidySources: %w", err)
	}
	lock, source, err := resolveV1Graph(ctx, module, existing, repository, true, nil, true)
	if err != nil {
		return TidyResult{}, fmt.Errorf("resolveV1Graph: %w", err)
	}
	bindings, err := source.tidyImportBindings(ctx, existing, lock)
	if err != nil {
		return TidyResult{}, fmt.Errorf("tidyImportBindings: %w", err)
	}
	if err := repository.Install(ctx, lock); err != nil {
		return TidyResult{}, fmt.Errorf("Install: %w", err)
	}
	dependencies, err := CachedSources(lock, repository)
	if err != nil {
		return TidyResult{}, fmt.Errorf("CachedSources: %w", err)
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
	files = filterTidyCapturedSources(tx, files, boundaries)
	roots := append(slices.Clone(own), dependencies...)
	if err := CheckSourceCollisions(roots); err != nil {
		return TidyResult{}, fmt.Errorf("CheckSourceCollisions: %w", err)
	}
	view := newTidySourceView(tx, roots)
	err = view.observeMetadata(source, lock)
	if err != nil {
		return TidyResult{}, fmt.Errorf("observeMetadata: %w", err)
	}
	verifyCache := tidyInputVerifier(ctx, lock, repository)
	tx.validateInputs = func() error {
		err := checkTidySourceSelection(tx, own, files)
		if err != nil {
			return fmt.Errorf("checkTidySourceSelection: %w", err)
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

type tidyImportBinding struct {
	before Fetched
	after  Fetched
	file   RootProtoFile
	names  map[string]RootProtoFile
}

func (source *importRootSource) tidyImportBindings(ctx context.Context, before, after v1.Lock) (map[string]tidyImportBinding, error) {
	bindings := make(map[string]tidyImportBinding)
	selector, supported := source.Source.(rootSelectionSource)
	if !supported {
		return bindings, nil
	}
	current := make(map[string]v1.LockedModule, len(after.Modules))
	for _, entry := range after.Modules {
		current[entry.Source] = entry
	}
	for _, old := range before.Modules {
		entry, retained := current[old.Source]
		if retained && strings.EqualFold(old.Commit, entry.Commit) && slices.Equal(old.Roots, entry.Roots) {
			continue
		}
		previous, err := source.lockedRootScope(ctx, selector, old)
		if err != nil {
			return nil, fmt.Errorf("lockedRootScope: %w", err)
		}
		if previous.Inspection == nil {
			previous, err = fetchLockedRootScope(ctx, selector, old, previous.Module.Roots)
			if err != nil {
				return nil, fmt.Errorf("fetchLockedRootScope: %w", err)
			}
		}
		previous.Lock.Version = old.Version
		next := source.fetched[[2]string{entry.Source, strings.ToLower(entry.Commit)}]
		if previous.Inspection == nil || previous.Inspection.Provisional || (retained && (next.Inspection == nil || next.Inspection.Provisional || next.Lock.Hash == "" || next.Lock.Hash != entry.Hash)) {
			return nil, fmt.Errorf("module %s: cannot verify old %s commit %s roots %v and new %s commit %s roots %v from complete pinned inspections; restore the old pin or verify explicit producer roots before running easyp mod tidy", old.Source, old.Version, old.Commit, previous.Module.Roots, entry.Version, entry.Commit, next.Module.Roots)
		}
		names := make(map[string]RootProtoFile)
		if retained {
			names, err = tidyRootNamespace(next)
			if err != nil {
				return nil, fmt.Errorf("tidyRootNamespace: %w", err)
			}
		}
		previousNames, err := tidyRootNamespace(previous)
		if err != nil {
			return nil, fmt.Errorf("tidyRootNamespace: %w", err)
		}
		for name, file := range previousNames {
			if prior, exists := bindings[name]; exists && prior.before.Lock.Source != old.Source {
				return nil, fmt.Errorf("old locked namespace has duplicate import %q from %s and %s", name, prior.before.Lock.Source, old.Source)
			}
			bindings[name] = tidyImportBinding{before: previous, after: next, file: file, names: names}
		}
	}
	return bindings, nil
}

func tidyRootNamespace(fetched Fetched) (map[string]RootProtoFile, error) {
	names := make(map[string]RootProtoFile)
	prefix := fetched.Lock.Source + "@" + fetched.Lock.Commit + ":"
	for _, file := range fetched.Inspection.Files {
		for _, root := range fetched.Module.Roots {
			name, within := importRootRelative(root, file.Path)
			if !within {
				continue
			}
			if !strings.HasPrefix(file.Identity, prefix) || !ValidProtoImportPath(strings.TrimPrefix(file.Identity, prefix)) {
				return nil, fmt.Errorf("module %s has an unverified pinned source identity for %s at commit %s", fetched.Lock.Source, file.Path, fetched.Lock.Commit)
			}
			if previous, exists := names[name]; exists && previous.Path != file.Path {
				return nil, fmt.Errorf("module %s has duplicate import path %q: %s and %s", fetched.Lock.Source, name, previous.Path, file.Path)
			}
			names[name] = file
		}
	}
	return names, nil
}

func (binding tidyImportBinding) replacement(path, name string) (string, error) {
	physical := rootGitSourcePath(binding.file, binding.before.Lock)
	var names []string
	for current, file := range binding.names {
		if physical == rootGitSourcePath(file, binding.after.Lock) {
			names = append(names, current)
		}
	}
	if current, retained := binding.names[name]; retained && (len(names) == 0 || rootGitSourcePath(current, binding.after.Lock) == physical) {
		// A moved source retaining its public name remains ordinary producer
		// evolution. A surviving source under a new name has priority instead.
		return name, nil
	}
	if len(names) == 1 {
		return names[0], nil
	}
	slices.Sort(names)
	reason := "the previous source is no longer exported by the same producer"
	if len(names) > 1 {
		reason = fmt.Sprintf("the previous physical source has ambiguous current names %v", names)
	}
	return "", fmt.Errorf("%s imports %q from module %s: old %s commit %s roots %v source %s; new %s commit %s roots %v: %s; restore the previous revision or adjust the consumer import after checking producer roots and generated SDK paths, then run easyp mod tidy", path, name, binding.before.Lock.Source, binding.before.Lock.Version, binding.before.Lock.Commit, binding.before.Module.Roots, binding.file.Path, binding.after.Lock.Version, binding.after.Lock.Commit, binding.after.Module.Roots, reason)
}
