package modules

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// EnsureLockedSources loads import roots from verified cached modules
// selected by the module's lockfile.
func EnsureLockedSources(ctx context.Context, moduleDir string, module v1.Module, repository Cache) (SourceRoots, error) {
	remoteRequires := RemoteRequirements(module)
	if len(remoteRequires) == 0 {
		return nil, nil
	}
	lock, err := ReadLock(filepath.Join(moduleDir, v1.LockFile))
	if err != nil {
		return nil, fmt.Errorf("read protobuf.lock for module %s; run easyp mod tidy: %w", module.Name, err)
	}
	if err := ValidateRequirements(remoteRequires, lock); err != nil {
		return nil, err
	}
	replaced := replacedModules(module)
	remoteLock := v1.Lock{Version: lock.Version}
	for _, entry := range lock.Modules {
		if !replaced[entry.Source] {
			remoteLock.Modules = append(remoteLock.Modules, entry)
		}
	}
	if err := repository.Install(ctx, remoteLock); err != nil {
		return nil, err
	}
	return cachedSources(lock, replaced, repository)
}

// RemoteRequirements excludes requirements supplied by local replacements.
func RemoteRequirements(module v1.Module) []v1.Requirement {
	return unreplacedRequirements(module.Requires, replacedModules(module))
}

func replacedModules(module v1.Module) map[string]bool {
	replaced := make(map[string]bool, len(module.Replaces))
	for _, replacement := range module.Replaces {
		replaced[replacement.Module] = true
	}
	return replaced
}

func unreplacedRequirements(requirements []v1.Requirement, replaced map[string]bool) []v1.Requirement {
	var remoteRequires []v1.Requirement
	for _, requirement := range requirements {
		if !replaced[requirement.Module] {
			remoteRequires = append(remoteRequires, requirement)
		}
	}
	return remoteRequires
}

// ReadLock parses a lockfile without installing or changing dependencies.
func ReadLock(path string) (v1.Lock, error) {
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

func cachedRoots(lock v1.Lock, repository Cache) ([]string, error) {
	roots, err := CachedSources(lock, repository)
	if err != nil {
		return nil, err
	}
	return roots.Paths(), nil
}

// CachedSources reads installed module roots and validates their transitive requirements.
// It does not install or verify cached contents; call Install first.
func CachedSources(lock v1.Lock, repository Cache) (SourceRoots, error) {
	return cachedSources(lock, nil, repository)
}

func cachedSources(lock v1.Lock, replaced map[string]bool, repository Cache) (SourceRoots, error) {
	var roots SourceRoots
	for _, entry := range lock.Modules {
		if replaced[entry.Source] {
			continue
		}
		installDir, dependency, err := repository.Cached(entry)
		if err != nil {
			return nil, fmt.Errorf("cached %s: %w", entry.Source, err)
		}
		if err := ValidateRequirements(unreplacedRequirements(dependency.Requires, replaced), lock); err != nil {
			return nil, fmt.Errorf("dependency %s: %w", dependency.Name, err)
		}
		dependencyRoots, err := ModuleSources(installDir, dependency)
		if err != nil {
			return nil, fmt.Errorf("ModuleSources: %w", err)
		}
		roots = append(roots, dependencyRoots...)
	}
	return roots, nil
}

// ValidateRequirements checks that the lock satisfies every requested version or commit.
func ValidateRequirements(requirements []v1.Requirement, lock v1.Lock) error {
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
