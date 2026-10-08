package modules

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

// tidyInputs owns consumer metadata, source selection and capture timing.
// The root keeps the absolute lexical spelling opened by the transaction.
type tidyInputs struct {
	observations           *resolvedFileObservations
	root                   string
	original               []byte
	module                 v1.Module
	previous               v1.Lock
	roots                  SourceRoots
	files                  []string
	boundaries             []tidyCacheBoundary
	completeCacheOwnership bool
}

func captureTidyInputs(observations *resolvedFileObservations, repository Repository) (tidyInputs, error) {
	inputs := tidyInputs{observations: observations, root: observations.requestedRoot}
	boundaries, err := tidyCacheBoundaries(repository, v1.Lock{})
	if err != nil {
		return tidyInputs{}, fmt.Errorf("tidyCacheBoundaries: %w", err)
	}
	err = checkTidyCacheOwnership(observations, boundaries)
	if err != nil {
		return tidyInputs{}, fmt.Errorf("checkTidyCacheOwnership: %w", err)
	}
	inputs.original, inputs.module, err = ReadManifest(inputs.root)
	if err != nil {
		return tidyInputs{}, fmt.Errorf("ReadManifest: %w", err)
	}
	if len(inputs.module.Replaces) > 0 {
		return inputs, nil
	}
	manifest, err := observations.capture(v1.ModuleFile, nil)
	if err != nil {
		return tidyInputs{}, fmt.Errorf("capture: %w", err)
	}
	if !bytes.Equal(manifest.data, inputs.original) {
		return tidyInputs{}, fmt.Errorf("protobuf.mod changed while reading: %w", sourceview.ErrChanged)
	}
	locked, err := observations.capture(v1.LockFile, nil)
	if err != nil {
		return tidyInputs{}, fmt.Errorf("capture: %w", err)
	}
	if locked.exists {
		inputs.previous, err = v1.ParseLock(bytes.NewReader(locked.data))
	} else {
		// Preserve the existing legacy-lock diagnostic on a missing native lock.
		_, err = ReadLock(filepath.Join(inputs.root, v1.LockFile))
		if errors.Is(err, os.ErrNotExist) {
			err = nil
		}
	}
	if err != nil {
		return tidyInputs{}, fmt.Errorf("ParseLock: %w", err)
	}
	inputs.boundaries, err = tidyCacheBoundaries(repository, inputs.previous)
	if err != nil {
		return tidyInputs{}, fmt.Errorf("tidyCacheBoundaries: %w", err)
	}
	err = checkTidyCacheOwnership(observations, inputs.boundaries)
	if err != nil {
		return tidyInputs{}, fmt.Errorf("checkTidyCacheOwnership: %w", err)
	}
	inputs.roots, err = ModuleSources(inputs.root, inputs.module)
	if err != nil {
		return tidyInputs{}, fmt.Errorf("ModuleSources: %w", err)
	}
	inputs.roots = excludeTidyCacheSources(inputs.roots, inputs.boundaries)
	_, inputs.completeCacheOwnership = repository.(sourceCacheDirectories)
	return inputs, nil
}

func (inputs *tidyInputs) captureConsumerSources() error {
	files, err := tidySourceFiles(inputs.observations, inputs.roots)
	if err != nil {
		return fmt.Errorf("tidySourceFiles: %w", err)
	}
	for _, name := range files {
		_, err := inputs.observations.capture(name, inputs.roots)
		if err != nil {
			return fmt.Errorf("capture: %w", err)
		}
	}
	inputs.files = files
	return nil
}

func (inputs *tidyInputs) captureResolvedSources(repository Repository, lock v1.Lock) error {
	nextBoundaries, err := tidyCacheBoundaries(repository, lock)
	if err != nil {
		return fmt.Errorf("tidyCacheBoundaries: %w", err)
	}
	inputs.boundaries = append(inputs.boundaries, nextBoundaries...)
	err = checkTidyCacheOwnership(inputs.observations, inputs.boundaries)
	if err != nil {
		return fmt.Errorf("checkTidyCacheOwnership: %w", err)
	}
	inputs.roots = excludeTidyCacheSources(inputs.roots, inputs.boundaries)
	if inputs.completeCacheOwnership {
		return nil
	}
	// Generic repositories cannot capture consumers until the newly resolved
	// snapshot directories have been excluded from the owning roots.
	err = inputs.captureConsumerSources()
	if err != nil {
		return fmt.Errorf("captureConsumerSources: %w", err)
	}
	return nil
}

func (inputs *tidyInputs) excludePreviousPinSources(repository Repository) error {
	// Old-pin proof may acquire a previously missing snapshot. Preserve its
	// boundary before any subsequent consumer selection recheck.
	boundaries, err := tidyCacheBoundaries(repository, inputs.previous)
	if err != nil {
		return fmt.Errorf("tidyCacheBoundaries: %w", err)
	}
	inputs.boundaries = append(inputs.boundaries, boundaries...)
	inputs.roots = excludeTidyCacheSources(inputs.roots, inputs.boundaries)
	return nil
}

func checkTidySourceSelection(tx *resolvedFileObservations, roots SourceRoots, expected []string) error {
	current, err := tidySourceFiles(tx, roots)
	if err != nil {
		return fmt.Errorf("tidySourceFiles: %w", err)
	}
	if !slices.Equal(current, expected) {
		return fmt.Errorf("owned consumer source selection changed since planning in %q: %w", tx.requestedRoot, sourceview.ErrChanged)
	}
	return nil
}

func tidySourceFiles(tx *resolvedFileObservations, roots SourceRoots) ([]string, error) {
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

func (inputs tidyInputs) capturedImports() (map[string]v1UnresolvedImport, error) {
	needed := make(map[string]v1UnresolvedImport)
	for _, file := range inputs.files {
		observed, captured := inputs.observations.capturedFile(file)
		if !captured {
			return nil, fmt.Errorf("consumer source %q was not captured during planning", file)
		}
		imports, err := ParseProtoImports(file, observed.data)
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
