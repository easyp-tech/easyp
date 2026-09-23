package api

import (
	"fmt"
	"path/filepath"
)

// checkV1ImportPathCollisions rejects different files with the same name from
// the compiler's point of view. Every root contributes paths relative to that
// root, regardless of which protobuf module owns the root.
func checkV1ImportPathCollisions(moduleDir string, moduleRoots, dependencyRoots []string) error {
	roots := make([]string, 0, len(moduleRoots)+len(dependencyRoots))
	for _, root := range moduleRoots {
		roots = append(roots, filepath.Join(moduleDir, root))
	}
	roots = append(roots, dependencyRoots...)
	seen := make(map[string]string)
	for _, root := range roots {
		err := walkV1ProtoFiles(root, func(path string) error {
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
