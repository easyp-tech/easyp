package modules

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	"golang.org/x/mod/semver"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// EnsureFrozenSources verifies a native manifest and its complete locked graph
// without resolving revisions or writing project files. Replacements in the main
// manifest are forbidden; replacements in dependencies are ignored. The cache
// may install missing contents, but only at the exact identities in the lock.
func EnsureFrozenSources(ctx context.Context, root string, repository Cache) (SourceRoots, error) {
	sources, err := ensureFrozenSources(ctx, root, repository)
	if err != nil {
		return nil, fmt.Errorf("frozen graph %s: %w", root, err)
	}
	return sources, nil
}

func ensureFrozenSources(ctx context.Context, root string, repository Cache) (SourceRoots, error) {
	_, module, err := ReadManifest(root)
	if err != nil {
		return nil, fmt.Errorf("ReadManifest: %w", err)
	}
	if len(module.Replaces) > 0 {
		return nil, fmt.Errorf("frozen mode rejects replace directives in protobuf.mod; remove replace before using --frozen")
	}
	lock, err := ReadLock(filepath.Join(root, v1.LockFile))
	if err != nil {
		return nil, fmt.Errorf("ReadLock: %w", err)
	}
	err = validateFrozenRequirements(module.Requires, lock)
	if err != nil {
		return nil, fmt.Errorf("validateFrozenRequirements: %w", err)
	}
	if len(module.Requires) == 0 {
		if len(lock.Modules) > 0 {
			return nil, fmt.Errorf("protobuf.lock contains unreachable locked module %s; run easyp mod tidy", lock.Modules[0].Source)
		}
		return nil, nil
	}
	if repository == nil {
		return nil, fmt.Errorf("frozen mode requires a cache for locked dependencies")
	}
	return frozenCachedSources(ctx, module, lock, repository)
}

func frozenCachedSources(ctx context.Context, module v1.Module, lock v1.Lock, repository Cache) (SourceRoots, error) {
	locked := make(map[string]v1.LockedModule, len(lock.Modules))
	for _, entry := range lock.Modules {
		locked[entry.Source] = entry
	}
	visited := make(map[string]bool, len(lock.Modules))
	sources := make(map[string]SourceRoots, len(lock.Modules))
	queue := slices.Clone(module.Requires)
	for len(queue) > 0 {
		err := ctx.Err()
		if err != nil {
			return nil, fmt.Errorf("Err: %w", err)
		}
		requirement := queue[0]
		queue = queue[1:]
		if visited[requirement.Module] {
			continue
		}
		visited[requirement.Module] = true
		entry := locked[requirement.Module]
		// An unrelated stale lock entry is not permission to acquire its repository.
		// Install only reached exact pins while walking their verified metadata.
		if err := repository.Install(ctx, v1.Lock{Version: lock.Version, Modules: []v1.LockedModule{entry}}); err != nil {
			return nil, fmt.Errorf("Install: %w", err)
		}
		directory, dependency, err := repository.Cached(entry)
		if err != nil {
			return nil, fmt.Errorf("Cached: %w", err)
		}
		err = validateFrozenRequirements(dependency.Requires, lock)
		if err != nil {
			return nil, fmt.Errorf("dependency %s: validateFrozenRequirements: %w", dependency.Name, err)
		}
		dependencyRoots, err := ModuleSources(directory, dependency)
		if err != nil {
			return nil, fmt.Errorf("ModuleSources: %w", err)
		}
		sources[requirement.Module] = dependencyRoots
		queue = append(queue, dependency.Requires...)
	}
	var roots SourceRoots
	for _, entry := range lock.Modules {
		if !visited[entry.Source] {
			return nil, fmt.Errorf("protobuf.lock contains unreachable locked module %s; run easyp mod tidy", entry.Source)
		}
		roots = append(roots, sources[entry.Source]...)
	}
	return roots, nil
}

func validateFrozenRequirements(requirements []v1.Requirement, lock v1.Lock) error {
	for _, requirement := range requirements {
		version := requirement.Version
		if version != "" && !v1.IsCommitRef(version) && !semver.IsValid(version) {
			return fmt.Errorf("require %s: expected semantic version or full Git commit, got %q", requirement.Module, version)
		}
	}
	err := ValidateRequirements(requirements, lock)
	if err != nil {
		return fmt.Errorf("ValidateRequirements: %w", err)
	}
	return nil
}
