package modules

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// PolicyModule is a resolved module identity and the directory containing its
// actual manifest. For nested and /vN modules this differs from the Git cache
// repository root and from its proto import roots.
type PolicyModule struct {
	Name      string
	Directory string
	Replaced  bool
	Files     PolicyFiles
}

// PolicyFile is one explicitly selected configuration from a pinned source.
// Path preserves its logical module-relative spelling; Canonical identifies its
// immutable physical source for cache and cycle detection.
type PolicyFile struct {
	Path      string
	Canonical string
	Content   []byte
}

// PolicyFiles reads selected policy files without acquiring dependencies.
type PolicyFiles interface {
	Read(context.Context, string) (PolicyFile, error)
}

// PolicyFilesCache optionally exposes the immutable files behind a verified pin.
// The prefix is the manifest directory relative to the installed repository.
type PolicyFilesCache interface {
	PolicyFiles(v1.LockedModule, string) PolicyFiles
}

// PolicyGraph contains only module identities reached through the consumer's
// already declared and resolved dependency graph.
type PolicyGraph map[string]PolicyModule

// ResolvePolicyGraph returns the verified module directories available to an
// extends reference. It never resolves versions or changes protobuf.mod/lock.
// In frozen mode it rejects replacements and uses only the exact locked graph.
func ResolvePolicyGraph(ctx context.Context, moduleDir string, cache Cache, frozen bool) (PolicyGraph, error) {
	return ResolvePolicyGraphAt(ctx, moduleDir, cache, frozen, nil)
}

// ResolvePolicyGraphAt additionally maps main-module replacements to a baseline
// snapshot. No historical policy may read a present-day replacement directory.
func ResolvePolicyGraphAt(ctx context.Context, moduleDir string, cache Cache, frozen bool, localPath func(string) (string, error)) (PolicyGraph, error) {
	_, module, err := ReadManifest(moduleDir)
	if err != nil {
		return nil, fmt.Errorf("ReadManifest: %w", err)
	}
	if frozen && len(module.Replaces) > 0 {
		return nil, fmt.Errorf("frozen mode rejects replace directives in protobuf.mod; remove replace before using --frozen")
	}
	lock, err := ReadLock(filepath.Join(moduleDir, v1.LockFile))
	if err != nil && (frozen || !errors.Is(err, os.ErrNotExist)) {
		return nil, fmt.Errorf("ReadLock: %w", err)
	}
	if errors.Is(err, os.ErrNotExist) {
		lock = v1.Lock{Version: 1}
	}
	locals := newLocalReplacements(moduleDir, module, localPath)
	result := make(PolicyGraph)
	visited := make(map[string]bool)
	queue := slices.Clone(module.Requires)
	for len(queue) > 0 {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("Err: %w", err)
		}
		requirement := queue[0]
		queue = queue[1:]
		dependency, local, err := locals.lookupRequirement(requirement.Module, requirement.Version)
		if err != nil {
			return nil, fmt.Errorf("lookupRequirement: %w", err)
		}
		if !local {
			// Every incoming edge is checked, including stronger edges encountered
			// after the same module has already been loaded. Never resolve a version.
			if err := ValidateRequirements([]v1.Requirement{requirement}, lock); err != nil {
				return nil, fmt.Errorf("policy dependency: %w", err)
			}
		}
		if visited[requirement.Module] {
			continue
		}
		visited[requirement.Module] = true
		var directory string
		var pin v1.LockedModule
		if local {
			directory = locals.modules[requirement.Module].Directory
		} else {
			if cache == nil {
				return nil, fmt.Errorf("policy dependency %s requires cached locked contents; run easyp mod download", requirement.Module)
			}
			for _, entry := range lock.Modules {
				if entry.Source == requirement.Module {
					pin = entry
					break
				}
			}
			if err := cache.Install(ctx, v1.Lock{Version: 1, Modules: []v1.LockedModule{pin}}); err != nil {
				return nil, fmt.Errorf("Install: %w", err)
			}
			directory, dependency, err = cache.Cached(pin)
			if err != nil {
				return nil, fmt.Errorf("Cached: %w", err)
			}
		}
		if dependency.Name != requirement.Module {
			return nil, fmt.Errorf("policy dependency %s declares %s", requirement.Module, dependency.Name)
		}
		manifestDir, err := moduleconfig.DependencyManifestDirectory(directory, dependency.Name)
		if err != nil {
			return nil, fmt.Errorf("DependencyManifestDirectory: %w", err)
		}
		policyModule := PolicyModule{Name: requirement.Module, Directory: manifestDir, Replaced: local}
		if provider, ok := cache.(PolicyFilesCache); ok && !local {
			prefix, err := filepath.Rel(directory, manifestDir)
			if err != nil {
				return nil, fmt.Errorf("Rel: %w", err)
			}
			policyModule.Files = provider.PolicyFiles(pin, filepath.ToSlash(prefix))
		}
		result[requirement.Module] = policyModule
		// Only the main module supplies replacements. Dependency replaces are not
		// inherited, just as for the ordinary source graph.
		queue = append(queue, dependency.Requires...)
	}
	if frozen {
		for _, pin := range lock.Modules {
			if !visited[pin.Source] {
				return nil, fmt.Errorf("protobuf.lock contains unreachable locked module %s; run easyp mod tidy", pin.Source)
			}
		}
	}
	return result, nil
}

// SortedNames returns graph identities in stable lexical order.
func (g PolicyGraph) SortedNames() []string {
	names := make([]string, 0, len(g))
	for name := range g {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}
