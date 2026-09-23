package api

import "path/filepath"

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
