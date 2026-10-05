package modules

import (
	"fmt"
	"path/filepath"
)

// CheckSourceCollisions respects each module's source selection when comparing import names.
func CheckSourceCollisions(roots SourceRoots) error {
	seen := make(map[string]string)
	allowed := roots.FileAllowed()
	for _, root := range roots {
		err := root.Walk(func(path string) error {
			if allowed != nil && !allowed(path) {
				return nil
			}
			importPath, err := filepath.Rel(root.Path, path)
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
