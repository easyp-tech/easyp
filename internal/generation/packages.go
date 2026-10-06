package generation

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
	"github.com/easyp-tech/easyp/internal/protosource"
)

type sourceSelectorMatches struct {
	packages       []string
	paths          []string
	packageMatches map[string]bool
	pathMatches    map[string]bool
}

func newSourceSelectorMatches(packages, paths []string) *sourceSelectorMatches {
	return &sourceSelectorMatches{packages: packages, paths: paths, packageMatches: make(map[string]bool), pathMatches: make(map[string]bool)}
}

func (s *sourceSelectorMatches) matchesPath(name string) bool {
	return len(s.paths) == 0 || slices.ContainsFunc(s.paths, func(selector string) bool { return v1.PathSelectorMatches(selector, name) })
}

func (s *sourceSelectorMatches) matchesPackage(name string) bool {
	return len(s.packages) == 0 || slices.Contains(s.packages, name)
}

func (s *sourceSelectorMatches) recordMatches(packageName, path string) {
	if len(s.packages) > 0 {
		s.packageMatches[packageName] = true
	}
	for _, selector := range s.paths {
		if v1.PathSelectorMatches(selector, path) {
			s.pathMatches[selector] = true
		}
	}
}

func (s *sourceSelectorMatches) validateMatches(section string) error {
	if unknown := unmatchedSourceSelectors(s.packages, s.packageMatches); len(unknown) > 0 {
		return fmt.Errorf("%s.packages did not match any selected module source files: %s", section, strings.Join(unknown, ", "))
	}
	if unknown := unmatchedSourceSelectors(s.paths, s.pathMatches); len(unknown) > 0 {
		return fmt.Errorf("%s.paths did not match any selected module source files: %s", section, strings.Join(unknown, ", "))
	}
	return nil
}

// selectedSourceFiles intersects literal module-relative paths and exact package
// names without compiling unrelated files. Package declarations are read only
// when needed and after the path filter; required imports are compiled later.
func selectedSourceFiles(ctx context.Context, selected v1GenerationModule, project, module *sourceSelectorMatches) ([]string, error) {
	roots, err := modules.ModuleSources(selected.directory, selected.module)
	if err != nil {
		return nil, fmt.Errorf("ModuleSources: %w", err)
	}
	allRoots := append(append(modules.SourceRoots(nil), roots...), selected.dependencies...)
	seen := make(map[string]bool)
	var files []string
	for _, root := range roots {
		err := allRoots.WalkSelected(root, func(path string) bool {
			relative, err := filepath.Rel(selected.directory, path)
			return err == nil && project.matchesPath(filepath.ToSlash(relative)) && module.matchesPath(filepath.ToSlash(relative))
		}, func(path string) error {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("Err: %w", err)
			}
			modulePath, err := filepath.Rel(selected.directory, path)
			if err != nil {
				return fmt.Errorf("Rel: %w", err)
			}
			modulePath = filepath.ToSlash(modulePath)
			if !project.matchesPath(modulePath) || !module.matchesPath(modulePath) {
				return nil
			}
			packageName := ""
			if len(project.packages) > 0 || len(module.packages) > 0 {
				raw, err := allRoots.ReadSourceFile(path)
				if err != nil {
					return fmt.Errorf("ReadFile: %w", err)
				}
				packageName = protosource.Package(raw)
				if !project.matchesPackage(packageName) || !module.matchesPackage(packageName) {
					return nil
				}
			}
			project.recordMatches(packageName, modulePath)
			module.recordMatches(packageName, modulePath)
			relative, err := filepath.Rel(root.Path, path)
			if err != nil {
				return fmt.Errorf("Rel: %w", err)
			}
			relative = filepath.ToSlash(relative)
			if !seen[relative] {
				seen[relative] = true
				files = append(files, relative)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("Walk: %w", err)
		}
	}
	slices.Sort(files)
	return files, nil
}
