package modules

import (
	"fmt"
	"os"
	"path/filepath"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// SourceRoot associates a physical import root with its module identity.
type SourceRoot struct {
	Path   string
	Module string
}

// SourceRoots preserves import precedence and module ownership.
type SourceRoots []SourceRoot

// Paths returns a new slice in the same order as the source roots.
func (roots SourceRoots) Paths() []string {
	paths := make([]string, 0, len(roots))
	for _, root := range roots {
		paths = append(paths, root.Path)
	}
	return paths
}

// FileModules maps import paths to their owning modules for managed selectors.
// Callers check import collisions before using this mapping.
func (roots SourceRoots) FileModules() (map[string]string, error) {
	modules := make(map[string]string)
	for _, root := range roots {
		err := WalkProtoFiles(root.Path, func(path string) error {
			relative, err := filepath.Rel(root.Path, path)
			if err != nil {
				return err
			}
			modules[filepath.ToSlash(relative)] = root.Module
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return modules, nil
}

// ModuleSources validates module roots and resolves them relative to its directory.
func ModuleSources(directory string, module v1.Module) (SourceRoots, error) {
	roots := make(SourceRoots, 0, len(module.Roots))
	for _, root := range module.Roots {
		path := filepath.Join(directory, root)
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("Stat: %w", err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("module %s has invalid root %q: not a directory", module.Name, root)
		}
		roots = append(roots, SourceRoot{Path: path, Module: module.Name})
	}
	return roots, nil
}
