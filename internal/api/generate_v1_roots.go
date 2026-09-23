package api

import (
	"fmt"
	"os"
	"path/filepath"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

type v1SourceRoot struct {
	path   string
	module string
}

type v1SourceRoots []v1SourceRoot

func (roots v1SourceRoots) paths() []string {
	paths := make([]string, 0, len(roots))
	for _, root := range roots {
		paths = append(paths, root.path)
	}
	return paths
}

func (roots v1SourceRoots) fileModules() (map[string]string, error) {
	modules := make(map[string]string)
	for _, root := range roots {
		err := walkV1ProtoFiles(root.path, func(path string) error {
			relative, err := filepath.Rel(root.path, path)
			if err != nil {
				return err
			}
			modules[filepath.ToSlash(relative)] = root.module
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return modules, nil
}

func moduleV1SourceRoots(directory string, module v1.Module) (v1SourceRoots, error) {
	roots := make(v1SourceRoots, 0, len(module.Roots))
	for _, root := range module.Roots {
		path := filepath.Join(directory, root)
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("Stat: %w", err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("module %s has invalid root %q: not a directory", module.Name, root)
		}
		roots = append(roots, v1SourceRoot{path: path, module: module.Name})
	}
	return roots, nil
}
