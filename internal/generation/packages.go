package generation

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
	"github.com/easyp-tech/easyp/internal/protosource"
)

// selectedSourceFiles intersects literal import-relative paths and exact package
// names without compiling unrelated files. Package declarations are read only
// when needed and after the path filter; required imports are compiled later.
func selectedSourceFiles(ctx context.Context, selected v1GenerationModule, packages, paths []string) ([]string, map[string]bool, map[string]bool, error) {
	roots, err := modules.ModuleSources(selected.directory, selected.module)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("ModuleSources: %w", err)
	}
	allRoots := append(append(modules.SourceRoots(nil), roots...), selected.dependencies...)
	requested := make(map[string]bool, len(packages))
	for _, name := range packages {
		requested[name] = true
	}
	packageMatches := make(map[string]bool)
	pathMatches := make(map[string]bool)
	seen := make(map[string]bool)
	var files []string
	for _, root := range roots {
		err := allRoots.WalkSelected(root, func(path string) bool {
			relative, err := filepath.Rel(root.Path, path)
			return err == nil && (len(paths) == 0 || slices.ContainsFunc(paths, func(selector string) bool { return v1.PathSelectorMatches(selector, filepath.ToSlash(relative)) }))
		}, func(path string) error {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("Err: %w", err)
			}
			relative, err := filepath.Rel(root.Path, path)
			if err != nil {
				return fmt.Errorf("Rel: %w", err)
			}
			relative = filepath.ToSlash(relative)
			if len(paths) > 0 && !slices.ContainsFunc(paths, func(selector string) bool {
				return v1.PathSelectorMatches(selector, relative)
			}) {
				return nil
			}
			if len(packages) > 0 {
				raw, err := allRoots.ReadSourceFile(path)
				if err != nil {
					return fmt.Errorf("ReadFile: %w", err)
				}
				name := protosource.Package(raw)
				if !requested[name] {
					return nil
				}
				packageMatches[name] = true
			}
			for _, selector := range paths {
				if v1.PathSelectorMatches(selector, relative) {
					pathMatches[selector] = true
				}
			}
			if !seen[relative] {
				seen[relative] = true
				files = append(files, relative)
			}
			return nil
		})
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Walk: %w", err)
		}
	}
	slices.Sort(files)
	return files, packageMatches, pathMatches, nil
}
