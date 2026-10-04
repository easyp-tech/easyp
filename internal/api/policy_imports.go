package api

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
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

func isReplacementTargetModule(projectRoot, moduleDir string) (bool, error) {
	if moduleDir == "" {
		return false, nil
	}
	for _, ancestor := range ancestorDirs(filepath.Dir(moduleDir), projectRoot) {
		manifest := filepath.Join(ancestor, v1.ModuleFile)
		_, err := os.Stat(manifest)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return false, fmt.Errorf("Stat: %w", err)
		}
		_, module, err := modules.ReadManifest(ancestor)
		if err != nil {
			return false, fmt.Errorf("ReadManifest: %w", err)
		}
		for _, replacement := range module.Replaces {
			target := modules.ResolveReplacementPath(ancestor, replacement.Target)
			match, err := samePolicyPath(target, moduleDir)
			if err != nil {
				return false, err
			}
			if match {
				return true, nil
			}
		}
	}
	return false, nil
}

func samePolicyPath(left, right string) (bool, error) {
	leftAbs, err := filepath.Abs(left)
	if err != nil {
		return false, fmt.Errorf("Abs: %w", err)
	}
	rightAbs, err := filepath.Abs(right)
	if err != nil {
		return false, fmt.Errorf("Abs: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(leftAbs); err == nil {
		leftAbs = resolved
	}
	if resolved, err := filepath.EvalSymlinks(rightAbs); err == nil {
		rightAbs = resolved
	}
	return filepath.Clean(leftAbs) == filepath.Clean(rightAbs), nil
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
