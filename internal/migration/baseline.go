package migration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
	"github.com/easyp-tech/easyp/internal/workspace"
)

// ReadLegacyBaselineModule reads only source/dependency metadata for an old Git
// baseline. Generation plugins and lint policies are not migrated or executed.
// The boolean distinguishes a legacy project from a native or unconfigured tree.
func ReadLegacyBaselineModule(directory string) (v1.Module, bool, error) {
	module, _, found, err := readLegacyBaselineModule(directory)
	return module, found, err
}

func readLegacyBaselineModule(directory string) (v1.Module, map[string]gitModuleSelection, bool, error) {
	manifest, err := workspace.ReadFileAt(filepath.Join(directory, v1.ModuleFile))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return v1.Module{}, nil, false, fmt.Errorf("ReadFileAt: %w", err)
	}
	hasManifest := err == nil
	if hasManifest && v1.IsModuleManifest(manifest) {
		return v1.Module{}, nil, false, nil
	}
	raw, err := workspace.ReadFileAt(filepath.Join(directory, v1.PolicyFile))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return v1.Module{}, nil, false, fmt.Errorf("ReadFileAt: %w", err)
	}
	if errors.Is(err, os.ErrNotExist) && !hasManifest {
		return v1.Module{}, nil, false, nil
	}
	var cfg legacyConfig
	if err == nil {
		var document yaml.Node
		if err := yaml.Unmarshal(raw, &document); err != nil {
			return v1.Module{}, nil, true, fmt.Errorf("Unmarshal: %w", err)
		}
		if len(document.Content) == 1 && document.Content[0].Kind == yaml.MappingNode && nativePolicy(document.Content[0]) {
			return v1.Module{}, nil, false, nil
		}
		cfg, _, err = parseLegacy(raw)
		if err != nil {
			return v1.Module{}, nil, true, fmt.Errorf("parseLegacy: %w", err)
		}
	}
	// Plugin pinning and generation target coverage do not affect compilation of
	// the baseline. In particular, old unpinned plugins must not block breaking.
	cfg.Generate.Plugins = nil
	roots, inputs, err := localRoots(cfg)
	if err != nil {
		return v1.Module{}, nil, true, fmt.Errorf("localRoots: %w", err)
	}
	if len(roots) == 0 {
		roots = []string{"."}
	}
	deps := requirements{}
	for _, dependency := range cfg.Deps {
		if err := deps.add(dependency, false); err != nil {
			return v1.Module{}, nil, true, fmt.Errorf("add: %w", err)
		}
	}
	const identity = "legacy.invalid/baseline"
	_, selections, err := planGitSelections(cfg, identity, len(inputs) > 0, nil, nil, &deps)
	if err != nil {
		return v1.Module{}, nil, true, fmt.Errorf("planGitSelections: %w", err)
	}
	if hasManifest {
		replacements, err := parseManifest(manifest, &deps)
		if err != nil {
			return v1.Module{}, nil, true, fmt.Errorf("parseManifest: %w", err)
		}
		if len(replacements) > 0 {
			return v1.Module{}, nil, true, fmt.Errorf("legacy baseline has local replacements whose historical dependency contents cannot be verified")
		}
	}
	return v1.Module{Name: identity, Roots: roots, Requires: deps.items}, selections, true, nil
}

// LegacyBaselineSources reconstructs a v0 baseline exclusively from its own
// easyp.lock. It never resolves HEAD or writes migrated files into the project.
// Cache installation is allowed only after the historical hashes are verified.
func LegacyBaselineSources(ctx context.Context, directory string, cache modules.Cache) (modules.SourceRoots, bool, error) {
	module, selections, found, err := readLegacyBaselineModule(directory)
	if err != nil || !found {
		return nil, found, err
	}
	own, err := modules.ModuleSources(directory, module)
	if err != nil {
		return nil, true, fmt.Errorf("ModuleSources: %w", err)
	}
	raw, err := workspace.ReadFileAt(filepath.Join(directory, "easyp.lock"))
	if errors.Is(err, os.ErrNotExist) {
		if len(module.Requires) > 0 {
			return nil, true, fmt.Errorf("legacy baseline requires its own easyp.lock; historical dependencies cannot be resolved from HEAD or the current project")
		}
		return own, true, nil
	}
	if err != nil {
		return nil, true, fmt.Errorf("ReadFileAt: %w", err)
	}
	pins, err := parseLegacyLock(raw)
	if err != nil {
		return nil, true, fmt.Errorf("parseLegacyLock: %w", err)
	}
	if len(pins) == 0 && len(module.Requires) == 0 {
		return own, true, nil
	}
	repository, ok := cache.(Repository)
	if !ok {
		return nil, true, fmt.Errorf("legacy baseline requires a repository that can verify historical dependency hashes")
	}
	lock, _, err := migrateSelectionLock(ctx, module, pins, true, repository, selections)
	if err != nil {
		return nil, true, fmt.Errorf("migrateSelectionLock: %w", err)
	}
	if err := cache.Install(ctx, lock); err != nil {
		return nil, true, fmt.Errorf("Install: %w", err)
	}
	dependencies, err := modules.CachedSources(lock, cache)
	if err != nil {
		return nil, true, fmt.Errorf("CachedSources: %w", err)
	}
	roots := append(own, dependencies...)
	if err := modules.CheckSourceCollisions(roots); err != nil {
		return nil, true, fmt.Errorf("CheckSourceCollisions: %w", err)
	}
	return roots, true, nil
}
