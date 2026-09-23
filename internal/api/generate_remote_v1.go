package api

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// lockedV1DependencySources loads import roots from verified cached modules
// selected by the module's lockfile.
func lockedV1DependencySources(ctx context.Context, moduleDir string, module v1.Module, cacheRoot string) (v1SourceRoots, error) {
	remoteRequires := remoteV1Requirements(module)
	if len(remoteRequires) == 0 {
		return nil, nil
	}
	lock, err := readV1Lock(filepath.Join(moduleDir, v1.LockFile))
	if err != nil {
		return nil, fmt.Errorf("read protobuf.lock for module %s; run easyp mod tidy: %w", module.Name, err)
	}
	if err := validateV1Requirements(remoteRequires, lock); err != nil {
		return nil, err
	}
	if err := downloadV1Lock(ctx, lock, cacheRoot); err != nil {
		return nil, err
	}
	return sourcesFromV1Lock(lock, cacheRoot)
}

func remoteV1Requirements(module v1.Module) []v1.Requirement {
	replaced := make(map[string]bool, len(module.Replaces))
	for _, replacement := range module.Replaces {
		replaced[replacement.Module] = true
	}
	var remoteRequires []v1.Requirement
	for _, requirement := range module.Requires {
		if !replaced[requirement.Module] {
			remoteRequires = append(remoteRequires, requirement)
		}
	}
	return remoteRequires
}

func readV1Lock(path string) (v1.Lock, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return v1.Lock{}, fmt.Errorf("ReadFile: %w", err)
	}
	lock, err := v1.ParseLock(bytes.NewReader(raw))
	if err != nil {
		return v1.Lock{}, fmt.Errorf("ParseLock: %w", err)
	}
	return lock, nil
}

func rootsFromV1Lock(lock v1.Lock, cacheRoot string) ([]string, error) {
	roots, err := sourcesFromV1Lock(lock, cacheRoot)
	if err != nil {
		return nil, err
	}
	return roots.paths(), nil
}

func sourcesFromV1Lock(lock v1.Lock, cacheRoot string) (v1SourceRoots, error) {
	var roots v1SourceRoots
	for _, entry := range lock.Modules {
		installDir := v1ModuleCachePath(cacheRoot, entry)
		dependency, err := moduleconfig.ReadGitDependency(installDir, entry.Source)
		if err != nil {
			return nil, fmt.Errorf("cached %s: %w", entry.Source, err)
		}
		if err := validateV1Requirements(dependency.Requires, lock); err != nil {
			return nil, fmt.Errorf("dependency %s: %w", dependency.Name, err)
		}
		dependencyRoots, err := moduleV1SourceRoots(installDir, dependency)
		if err != nil {
			return nil, fmt.Errorf("moduleV1SourceRoots: %w", err)
		}
		roots = append(roots, dependencyRoots...)
	}
	return roots, nil
}

func validateV1Requirements(requirements []v1.Requirement, lock v1.Lock) error {
	locked := make(map[string]v1.LockedModule, len(lock.Modules))
	for _, entry := range lock.Modules {
		if _, duplicate := locked[entry.Source]; duplicate {
			return fmt.Errorf("duplicate locked module %s", entry.Source)
		}
		locked[entry.Source] = entry
	}
	for _, requirement := range requirements {
		entry, ok := locked[requirement.Module]
		if !ok || !v1RequirementSatisfied(requirement.Version, entry) {
			return fmt.Errorf("protobuf.lock does not satisfy %s %s; run easyp mod tidy", requirement.Module, requirement.Version)
		}
	}
	return nil
}

func v1RequirementSatisfied(required string, entry v1.LockedModule) bool {
	if required == "" {
		return true
	}
	if v1.IsCommitRef(required) {
		return strings.EqualFold(entry.Commit, required)
	}
	return semver.IsValid(entry.Version) && semver.Compare(entry.Version, required) >= 0
}
