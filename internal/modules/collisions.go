package modules

import (
	"fmt"
	"path/filepath"
)

// CheckImportCollisions rejects different files with the same name from
// the compiler's point of view. Every root contributes paths relative to that
// root, regardless of which protobuf module owns the root.
func CheckImportCollisions(moduleDir string, moduleRoots, dependencyRoots []string) error {
	roots := make([]string, 0, len(moduleRoots)+len(dependencyRoots))
	for _, root := range moduleRoots {
		roots = append(roots, filepath.Join(moduleDir, root))
	}
	roots = append(roots, dependencyRoots...)
	seen := make(map[string]string)
	for _, root := range roots {
		err := WalkProtoFiles(root, func(path string) error {
			importPath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			importPath = filepath.ToSlash(importPath)
			if previous, ok := seen[importPath]; ok && previous != path {
				return fmt.Errorf("duplicate import path %q: %s and %s", importPath, previous, path)
			}
			seen[importPath] = path
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
