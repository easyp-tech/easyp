package modules

import (
	"fmt"
	"path/filepath"
)

// CheckSourceCollisions respects each module's source selection when comparing import names.
func CheckSourceCollisions(roots SourceRoots) error {
	_, err := roots.indexFiles(true)
	return err
}

func (roots SourceRoots) indexFiles(checkCollisions bool) (map[string]string, error) {
	seen := make(map[string]string)
	owners := make(map[string]string)
	allowed := roots.FileAllowed()
	for _, root := range roots {
		err := root.walk(allowed, true, roots.ownsSelectedPhysical, func(path string) error {
			importPath, err := filepath.Rel(root.Path, path)
			if err != nil {
				return err
			}
			importPath = filepath.ToSlash(importPath)
			if previous, ok := seen[importPath]; checkCollisions && ok && previous != path {
				return fmt.Errorf("duplicate import path %q: %s and %s", importPath, previous, path)
			}
			seen[importPath] = path
			owners[importPath] = root.Module
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return owners, nil
}

// IndexFileModules checks import collisions and records module ownership in a
// single source traversal. The returned map can be shared by generation stages.
func (roots SourceRoots) IndexFileModules() (map[string]string, error) {
	return roots.indexFiles(true)
}
