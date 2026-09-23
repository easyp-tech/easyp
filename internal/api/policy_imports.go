package api

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
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

func resolveV1PolicyImportRoots(ctx context.Context, log logger.Logger, moduleDir string) ([]string, error) {
	_, module, err := readV1Manifest(moduleDir)
	if err != nil {
		return nil, fmt.Errorf("readV1Manifest: %w", err)
	}
	roots, err := moduleV1SourceRoots(moduleDir, module)
	if err != nil {
		return nil, fmt.Errorf("moduleV1SourceRoots: %w", err)
	}
	dependencies, err := resolveV1DependencySources(ctx, log, moduleDir, module)
	if err != nil {
		return nil, fmt.Errorf("resolveV1DependencySources: %w", err)
	}
	if err := checkV1ImportPathCollisions(moduleDir, module.Roots, dependencies.paths()); err != nil {
		return nil, fmt.Errorf("checkV1ImportPathCollisions: %w", err)
	}
	return append(roots.paths(), dependencies.paths()...), nil
}
