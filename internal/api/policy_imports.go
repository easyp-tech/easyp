package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/easyp-tech/easyp/internal/workspace"
	"os"
	"path/filepath"

	gitadapter "github.com/easyp-tech/easyp/internal/adapters/go_git"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core/path_helpers"
	"github.com/easyp-tech/easyp/internal/migration"
	"github.com/easyp-tech/easyp/internal/modules"
)

// findV1PolicyModuleDir finds the nearest module containing the scanned files.
// A policy can also be used without a module manifest.
func findV1PolicyModuleDir(projectRoot, scanDir string) (string, error) {
	for _, dir := range ancestorDirs(scanDir, projectRoot) {
		_, err := os.Stat(filepath.Join(dir, v1.ModuleFile))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("Stat: %w", err)
		}
		return dir, nil
	}
	return "", nil
}

type policyReplacementSources struct {
	ctx                         context.Context
	projectRoot, repositoryRoot string
	cache                       func() (modules.Cache, error)
	frozen                      bool
	manifests                   map[string]v1.Module
	targets                     map[string][]string
}

func newPolicyReplacementSources(ctx context.Context, projectRoot, repositoryRoot string, cache func() (modules.Cache, error), frozen bool) *policyReplacementSources {
	return &policyReplacementSources{
		ctx: ctx, projectRoot: projectRoot, repositoryRoot: repositoryRoot, cache: cache, frozen: frozen,
		manifests: make(map[string]v1.Module), targets: make(map[string][]string),
	}
}

func (s *policyReplacementSources) isUnselectedReplacementSource(scanPath, sourcePath string) (bool, error) {
	directories := ancestorDirs(filepath.Dir(sourcePath), s.projectRoot)
	// Read consuming manifests before replacement metadata, which may use a
	// legacy format or declare additional nested modules.
	for i := len(directories) - 1; i >= 0; i-- {
		ancestor := directories[i]
		targets, err := s.replacementTargets(ancestor, scanPath, sourcePath)
		if err != nil {
			return false, fmt.Errorf("replacementTargets: %w", err)
		}
		for _, target := range targets {
			containsSource, err := policyPathContains(target, sourcePath)
			if err != nil {
				return false, fmt.Errorf("policyPathContains: %w", err)
			}
			if !containsSource {
				continue
			}
			selected, err := policyPathContains(target, scanPath)
			if err != nil {
				return false, fmt.Errorf("policyPathContains: %w", err)
			}
			if !selected {
				return true, nil
			}
		}
	}
	return false, nil
}

// Resolve each consuming graph once per discovery pass. Only reached main-module
// replacements own import-only trees; unused directives never open their targets.
func (s *policyReplacementSources) replacementTargets(directory, scanPath, sourcePath string) ([]string, error) {
	if targets, known := s.targets[directory]; known {
		return targets, nil
	}
	module, err := s.replacementManifest(directory)
	if err != nil {
		return nil, fmt.Errorf("replacementManifest: %w", err)
	}
	if len(module.Replaces) == 0 {
		s.targets[directory] = nil
		return nil, nil
	}
	// An explicitly selected replacement or independent module does not use an
	// unrelated ancestor's graph. Declarations only identify possible ownership;
	// the effective graph below decides whether a candidate is actually reached.
	candidate := false
	for _, replacement := range module.Replaces {
		target := modules.ResolveReplacementPath(directory, replacement.Target)
		if filepath.IsAbs(replacement.Target) && s.projectRoot != s.repositoryRoot {
			relative, err := baselineRepositoryRelative(s.repositoryRoot, target)
			if err != nil {
				return nil, fmt.Errorf("baselineRepositoryRelative: %w", err)
			}
			if !filepath.IsLocal(relative) {
				continue
			}
			target = filepath.Join(s.projectRoot, relative)
		}
		containsSource, err := policyPathContains(target, sourcePath)
		if err != nil {
			return nil, fmt.Errorf("policyPathContains: %w", err)
		}
		if !containsSource {
			continue
		}
		selected, err := policyPathContains(target, scanPath)
		if err != nil {
			return nil, fmt.Errorf("policyPathContains: %w", err)
		}
		if !selected {
			candidate = true
			break
		}
	}
	if !candidate {
		return nil, nil
	}
	if s.frozen {
		return nil, fmt.Errorf("frozen mode rejects replace directives in protobuf.mod; remove replace before using --frozen")
	}
	var cache modules.Cache
	if s.cache != nil {
		cache, err = s.cache()
		if err != nil {
			return nil, fmt.Errorf("cache: %w", err)
		}
	}
	var localPath func(string) (string, error)
	if s.projectRoot != s.repositoryRoot {
		snapshot := &gitadapter.Snapshot{Root: s.projectRoot, RepositoryRoot: s.repositoryRoot}
		localPath = func(target string) (string, error) {
			return snapshotReplacementPath(snapshot, directory, target)
		}
	}
	graph, err := modules.EnsureEffectiveGraph(s.ctx, directory, module, cache, localPath)
	if err != nil {
		return nil, fmt.Errorf("EnsureEffectiveGraph: %w", err)
	}
	var targets []string
	for _, replacement := range module.Replaces {
		if dependency, reached := graph.Modules[replacement.Module]; reached {
			targets = append(targets, dependency.Directory)
		}
	}
	s.targets[directory] = targets
	return targets, nil
}

func (s *policyReplacementSources) replacementManifest(directory string) (v1.Module, error) {
	if module, known := s.manifests[directory]; known {
		return module, nil
	}
	raw, err := workspace.ReadFileAt(filepath.Join(directory, v1.ModuleFile))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return v1.Module{}, fmt.Errorf("ReadFile: %w", err)
	}
	var module v1.Module
	if s.projectRoot != s.repositoryRoot && !v1.IsModuleManifest(raw) {
		module, _, err = migration.ReadLegacyBaselineModule(directory)
		if err != nil {
			return v1.Module{}, fmt.Errorf("ReadLegacyBaselineModule: %w", err)
		}
	}
	if v1.IsModuleManifest(raw) {
		module, err = v1.ParseModule(bytes.NewReader(raw))
		if err != nil {
			return v1.Module{}, fmt.Errorf("ParseModule: %w", err)
		}
	}
	s.manifests[directory] = module
	return module, nil
}

func policyPathContains(directory, path string) (bool, error) {
	directoryAbs, err := filepath.Abs(directory)
	if err != nil {
		return false, fmt.Errorf("Abs: %w", err)
	}
	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return false, fmt.Errorf("Abs: %w", err)
	}
	info, err := os.Stat(directoryAbs)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("Stat: %w", err)
	}
	if !info.IsDir() {
		return false, nil
	}
	directoryAbs, err = filepath.EvalSymlinks(directoryAbs)
	if err != nil {
		return false, fmt.Errorf("EvalSymlinks: %w", err)
	}
	pathAbs, err = canonicalPolicyPath(pathAbs)
	if err != nil {
		return false, fmt.Errorf("canonicalPolicyPath: %w", err)
	}
	return path_helpers.IsTargetPath(directoryAbs, pathAbs), nil
}

func ensureV1PolicyImportRoots(ctx context.Context, cache modules.Cache, moduleDir string) (modules.SourceRoots, error) {
	return policyImportRoots(ctx, cache, moduleDir, false)
}

func policyImportRoots(ctx context.Context, cache modules.Cache, moduleDir string, frozen bool) (modules.SourceRoots, error) {
	var dependencies modules.SourceRoots
	var err error
	if frozen {
		dependencies, err = modules.EnsureFrozenSources(ctx, moduleDir, cache)
		if err != nil {
			return nil, fmt.Errorf("EnsureFrozenSources: %w", err)
		}
	}
	_, module, err := modules.ReadManifest(moduleDir)
	if err != nil {
		return nil, fmt.Errorf("ReadManifest: %w", err)
	}
	roots, err := modules.ModuleSources(moduleDir, module)
	if err != nil {
		return nil, fmt.Errorf("ModuleSources: %w", err)
	}
	if !frozen {
		dependencies, err = modules.EnsureSources(ctx, moduleDir, module, cache)
		if err != nil {
			return nil, fmt.Errorf("EnsureSources: %w", err)
		}
	}
	allSources := append(roots, dependencies...)
	if err := modules.CheckSourceCollisions(allSources); err != nil {
		return nil, fmt.Errorf("CheckSourceCollisions: %w", err)
	}
	return allSources, nil
}
