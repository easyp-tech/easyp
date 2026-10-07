package modules

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// sourceCacheDirectories is an optional ownership port for repository storage,
// including snapshots that are not referenced by the current lock.
type sourceCacheDirectories interface {
	SourceCacheDirectories() []string
}

type tidyCacheBoundary struct {
	logical  string
	physical string
}

func tidyCacheBoundaries(repository Repository, lock v1.Lock) ([]tidyCacheBoundary, error) {
	var directories []string
	if cache, supported := repository.(sourceCacheDirectories); supported {
		directories = cache.SourceCacheDirectories()
	} else {
		for _, entry := range lock.Modules {
			directory, _, err := repository.Cached(entry)
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("Cached: %w", err)
			}
			directories = append(directories, directory)
		}
	}
	return newTidyCacheBoundaries(directories)
}

func newTidyCacheBoundaries(directories []string) ([]tidyCacheBoundary, error) {
	var boundaries []tidyCacheBoundary
	for _, directory := range directories {
		logical, err := filepath.Abs(directory)
		if err != nil {
			return nil, fmt.Errorf("Abs: %w", err)
		}
		// Resolve an existing ancestor when a first tidy has not created the
		// cache yet. Storage aliases still establish physical ownership.
		for ancestor := logical; ; ancestor = filepath.Dir(ancestor) {
			physical, err := filepath.EvalSymlinks(ancestor)
			if err == nil {
				relative, err := filepath.Rel(ancestor, logical)
				if err != nil {
					return nil, fmt.Errorf("Rel: %w", err)
				}
				boundaries = append(boundaries, tidyCacheBoundary{logical: logical, physical: filepath.Join(physical, relative)})
				break
			}
			if !errors.Is(err, os.ErrNotExist) || ancestor == filepath.Dir(ancestor) {
				return nil, fmt.Errorf("EvalSymlinks: %w", err)
			}
		}
	}
	return boundaries, nil
}

func tidyCacheContains(path string, boundaries []tidyCacheBoundary) bool {
	physical := physicalSourcePath(path)
	for _, boundary := range boundaries {
		if sourcePathWithin(path, boundary.logical) || sourcePathWithin(physical, boundary.physical) {
			return true
		}
	}
	return false
}

func excludeTidyCacheSources(roots SourceRoots, boundaries []tidyCacheBoundary) SourceRoots {
	roots = slices.Clone(roots)
	boundaries = slices.Clone(boundaries)
	for index := range roots {
		fileAllowed, directoryAllowed := roots[index].fileAllowed, roots[index].directoryAllowed
		roots[index].fileAllowed = func(path string) bool {
			return !tidyCacheContains(path, boundaries) && (fileAllowed == nil || fileAllowed(path))
		}
		roots[index].directoryAllowed = func(path string) bool {
			return !tidyCacheContains(path, boundaries) && (directoryAllowed == nil || directoryAllowed(path))
		}
	}
	return roots
}

func checkTidyCacheOwnership(tx *resolvedFilesTransaction, boundaries []tidyCacheBoundary) error {
	if tidyCacheContains(tx.requestedRoot, boundaries) || tidyCacheContains(tx.canonicalRoot, boundaries) {
		return fmt.Errorf("consumer module %q is inside repository-owned cache storage; run easyp mod tidy from the consumer module outside the cache", tx.requestedRoot)
	}
	for _, name := range []string{v1.ModuleFile, v1.LockFile} {
		state, captured := tx.expected[name]
		if !captured {
			continue
		}
		if tidyCacheContains(filepath.Join(tx.canonicalRoot, filepath.FromSlash(state.resolution.Path)), boundaries) {
			return fmt.Errorf("consumer metadata %q resolves into repository-owned cache storage; restore its owning module path before running easyp mod tidy", name)
		}
	}
	return nil
}

func filterTidyCapturedSources(tx *resolvedFilesTransaction, files []string, boundaries []tidyCacheBoundary) []string {
	var owned []string
	for _, name := range files {
		if tidyCacheContains(filepath.Join(tx.requestedRoot, name), boundaries) || tidyCacheContains(filepath.Join(tx.canonicalRoot, filepath.FromSlash(tx.expected[name].resolution.Path)), boundaries) {
			delete(tx.expected, name)
			continue
		}
		owned = append(owned, name)
	}
	return owned
}
