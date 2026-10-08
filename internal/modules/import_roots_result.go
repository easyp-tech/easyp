package modules

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

type rootRevisionKey struct {
	source string
	commit string
}

func rootRevision(source, commit string) rootRevisionKey {
	return rootRevisionKey{source: source, commit: strings.ToLower(commit)}
}

// graphResolveResult owns only final selected revisions and previously checked
// exact root scopes. Each source view is independent of the retained evidence.
type graphResolveResult struct {
	lock           v1.Lock
	revisions      map[rootRevisionKey]Fetched
	previousScopes map[rootRevisionKey]Fetched
}

func (source *importRootSource) checkedResult(lock v1.Lock) (graphResolveResult, error) {
	err := lock.Validate()
	if err != nil {
		return graphResolveResult{}, fmt.Errorf("Validate: %w", err)
	}
	result := graphResolveResult{
		lock:           cloneRootLock(lock),
		revisions:      make(map[rootRevisionKey]Fetched, len(lock.Modules)),
		previousScopes: make(map[rootRevisionKey]Fetched, len(source.verifiedScopes)),
	}
	for _, entry := range lock.Modules {
		key := rootRevision(entry.Source, entry.Commit)
		fetched, exists := source.fetched[key]
		if !exists || fetched.Module.Name != entry.Source || fetched.Lock.Source != entry.Source || !strings.EqualFold(fetched.Lock.Commit, entry.Commit) || fetched.Lock.Hash != entry.Hash {
			return graphResolveResult{}, fmt.Errorf("module %s: source evidence differs from selected commit %s hash %s", entry.Source, entry.Commit, entry.Hash)
		}
		if fetched.Inspection != nil && fetched.Inspection.Provisional {
			return graphResolveResult{}, fmt.Errorf("module %s: selected source inspection remains provisional", entry.Source)
		}
		// MVS can reuse one exact revision under a tag or commit requirement.
		// Retain the final selected pin without resolving its version again.
		fetched.Lock = entry
		result.revisions[key] = cloneRootFetched(fetched)
	}
	for _, previous := range source.verifiedScopes {
		key := rootRevision(previous.Lock.Source, previous.Lock.Commit)
		result.previousScopes[key] = cloneRootFetched(previous)
	}
	return result, nil
}

func (result graphResolveResult) lockFile() v1.Lock {
	return cloneRootLock(result.lock)
}

func (result graphResolveResult) currentRevision(entry v1.LockedModule) (Fetched, bool) {
	fetched, exists := result.revisions[rootRevision(entry.Source, entry.Commit)]
	if !exists || fetched.Lock.Hash != entry.Hash {
		return Fetched{}, false
	}
	return cloneRootFetched(fetched), true
}

func (result graphResolveResult) previousRootScope(entry v1.LockedModule) (Fetched, bool) {
	fetched, exists := result.previousScopes[rootRevision(entry.Source, entry.Commit)]
	if !exists || fetched.Lock.Hash != entry.Hash {
		return Fetched{}, false
	}
	return cloneRootFetched(fetched), true
}

func cloneRootLock(lock v1.Lock) v1.Lock {
	lock.Modules = slices.Clone(lock.Modules)
	for offset, entry := range lock.Modules {
		lock.Modules[offset] = cloneRootPin(entry)
	}
	return lock
}

func cloneRootPin(entry v1.LockedModule) v1.LockedModule {
	entry.Roots = slices.Clone(entry.Roots)
	entry.BSR = slices.Clone(entry.BSR)
	return entry
}

func cloneRootFetched(fetched Fetched) Fetched {
	fetched.Lock = cloneRootPin(fetched.Lock)
	fetched.Module.Roots = slices.Clone(fetched.Module.Roots)
	fetched.Module.Requires = slices.Clone(fetched.Module.Requires)
	fetched.Module.Replaces = slices.Clone(fetched.Module.Replaces)
	fetched.Module.BSRDependencies = slices.Clone(fetched.Module.BSRDependencies)
	fetched.Module.ProtoFilters = slices.Clone(fetched.Module.ProtoFilters)
	for offset, filter := range fetched.Module.ProtoFilters {
		filter.Includes = slices.Clone(filter.Includes)
		filter.Excludes = slices.Clone(filter.Excludes)
		fetched.Module.ProtoFilters[offset] = filter
	}
	fetched.Inspection = cloneRootInspection(fetched.Inspection)
	return fetched
}

func cloneRootInspection(inspection *RootInspection) *RootInspection {
	if inspection == nil {
		return nil
	}
	owned := *inspection
	owned.Files = slices.Clone(inspection.Files)
	for offset, file := range owned.Files {
		file.Content = bytes.Clone(file.Content)
		owned.Files[offset] = file
	}
	owned.Problems = slices.Clone(inspection.Problems)
	owned.LegacyFiles = maps.Clone(inspection.LegacyFiles)
	return &owned
}
