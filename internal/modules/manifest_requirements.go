package modules

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// augmentV1ManifestRequirements records selected transitive modules. A module
// imported by a root .proto file is direct even if it arrived transitively.
func augmentV1ManifestRequirements(original []byte, root string, module v1.Module, lock v1.Lock, repository Cache) ([]byte, error) {
	existing := make(map[string]bool, len(module.Requires))
	for _, requirement := range module.Requires {
		existing[requirement.Module] = true
	}
	imports, err := v1RootImports(root, module.Roots)
	if err != nil {
		return nil, fmt.Errorf("v1RootImports: %w", err)
	}
	var additions []v1ManifestRequirement
	for _, entry := range lock.Modules {
		if existing[entry.Source] {
			continue
		}
		installDir, dependency, err := repository.Cached(entry)
		if err != nil {
			return nil, fmt.Errorf("Cached: %w", err)
		}
		direct := false
		for _, depRoot := range dependency.Roots {
			base := filepath.Join(installDir, depRoot)
			for importPath := range imports {
				if !filepath.IsLocal(importPath) {
					continue
				}
				info, err := os.Stat(filepath.Join(base, filepath.FromSlash(importPath)))
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				if err != nil {
					return nil, fmt.Errorf("Stat: %w", err)
				}
				if info.Mode().IsRegular() {
					direct = true
					break
				}
			}
			if direct {
				break
			}
		}
		additions = append(additions, v1ManifestRequirement{Requirement: v1.Requirement{Module: entry.Source, Version: entry.Version}, indirect: !direct})
	}
	updated := appendV1Requirements(original, additions)
	if _, err := v1.ParseModule(bytes.NewReader(updated)); err != nil {
		return nil, fmt.Errorf("ParseModule: %w", err)
	}
	return updated, nil
}

func v1RootImports(moduleDir string, roots []string) (map[string]bool, error) {
	imports := make(map[string]bool)
	for _, root := range roots {
		sourceRoot := filepath.Join(moduleDir, root)
		err := WalkProtoFiles(sourceRoot, func(path string) error {
			paths, err := ReadProtoImports(path)
			if err != nil {
				return fmt.Errorf("ReadProtoImports: %w", err)
			}
			for _, importPath := range paths {
				imports[importPath] = true
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("WalkProtoFiles: %w", err)
		}
	}
	return imports, nil
}
