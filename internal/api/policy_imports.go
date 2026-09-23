package api

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/logger"
)

// findV1PolicyModuleDir finds the nearest module containing the scanned files.
// A policy can also be used without a module manifest.
func findV1PolicyModuleDir(projectRoot, scanDir string) (string, error) {
	for dir := scanDir; ; dir = filepath.Dir(dir) {
		_, err := os.Stat(filepath.Join(dir, "protobuf.mod"))
		switch {
		case errors.Is(err, os.ErrNotExist):
		case err != nil:
			return "", fmt.Errorf("stat protobuf.mod in %s: %w", dir, err)
		default:
			return dir, nil
		}
		if dir == projectRoot || dir == filepath.Dir(dir) {
			return "", nil
		}
	}
}

func resolveV1PolicyImportRoots(ctx context.Context, log logger.Logger, moduleDir string) ([]string, error) {
	_, module, err := readV1Manifest(moduleDir)
	if err != nil {
		return nil, err
	}
	dependencies, err := resolveV1DependencySources(ctx, log, moduleDir, module)
	if err != nil {
		return nil, err
	}
	if err := checkV1ImportPathCollisions(moduleDir, module.Roots, dependencies.paths()); err != nil {
		return nil, fmt.Errorf("module %s: %w", module.Name, err)
	}
	roots := make([]string, 0, len(module.Roots)+len(dependencies))
	for _, root := range module.Roots {
		roots = append(roots, filepath.Join(moduleDir, root))
	}
	roots = append(roots, dependencies.paths()...)
	return roots, nil
}
