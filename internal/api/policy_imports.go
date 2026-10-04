package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core/path_helpers"
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

func isUnselectedReplacementSource(projectRoot, repositoryRoot, scanPath, sourcePath string) (bool, error) {
	directories := ancestorDirs(filepath.Dir(sourcePath), projectRoot)
	// Read consuming manifests before replacement metadata, which may use a
	// legacy format or declare additional nested modules.
	for i := len(directories) - 1; i >= 0; i-- {
		ancestor := directories[i]
		manifest := filepath.Join(ancestor, v1.ModuleFile)
		raw, err := os.ReadFile(manifest)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return false, fmt.Errorf("ReadFile: %w", err)
		}
		if !v1.IsModuleManifest(raw) {
			continue
		}
		module, err := v1.ParseModule(bytes.NewReader(raw))
		if err != nil {
			return false, fmt.Errorf("ParseModule: %w", err)
		}
		for _, replacement := range module.Replaces {
			target := modules.ResolveReplacementPath(ancestor, replacement.Target)
			// Snapshot manifests retain absolute paths from their original checkout.
			if filepath.IsAbs(replacement.Target) && projectRoot != repositoryRoot {
				relative, err := baselineRepositoryRelative(repositoryRoot, target)
				if err != nil {
					return false, fmt.Errorf("baselineRepositoryRelative: %w", err)
				}
				if !filepath.IsLocal(relative) {
					continue
				}
				target = filepath.Join(projectRoot, relative)
			}
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
	// A deleted source can still be selected for baseline comparison. Resolve
	// its nearest existing ancestor so filesystem aliases retain that selection.
	suffix := ""
	for {
		resolved, err := filepath.EvalSymlinks(pathAbs)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) || filepath.Dir(pathAbs) == pathAbs {
				return false, fmt.Errorf("EvalSymlinks: %w", err)
			}
			suffix = filepath.Join(filepath.Base(pathAbs), suffix)
			pathAbs = filepath.Dir(pathAbs)
			continue
		}
		return path_helpers.IsTargetPath(directoryAbs, filepath.Join(resolved, suffix)), nil
	}
}

func ensureV1PolicyImportRoots(ctx context.Context, cache modules.Cache, moduleDir string) ([]string, error) {
	return policyImportRoots(ctx, cache, moduleDir, false)
}

func policyImportRoots(ctx context.Context, cache modules.Cache, moduleDir string, frozen bool) ([]string, error) {
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
	if err := modules.CheckImportCollisions(moduleDir, module.Roots, dependencies.Paths()); err != nil {
		return nil, fmt.Errorf("CheckImportCollisions: %w", err)
	}
	return append(roots.Paths(), dependencies.Paths()...), nil
}
