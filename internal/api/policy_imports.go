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

func ensureV1PolicyImportRoots(ctx context.Context, cache modules.Cache, moduleDir string) ([]string, error) {
	_, module, err := modules.ReadManifest(moduleDir)
	if err != nil {
		return nil, fmt.Errorf("ReadManifest: %w", err)
	}
	roots, err := modules.ModuleSources(moduleDir, module)
	if err != nil {
		return nil, fmt.Errorf("ModuleSources: %w", err)
	}
	dependencies, err := modules.EnsureSources(ctx, moduleDir, module, cache)
	if err != nil {
		return nil, fmt.Errorf("EnsureSources: %w", err)
	}
	if err := modules.CheckImportCollisions(moduleDir, module.Roots, dependencies.Paths()); err != nil {
		return nil, fmt.Errorf("CheckImportCollisions: %w", err)
	}
	return append(roots.Paths(), dependencies.Paths()...), nil
}
