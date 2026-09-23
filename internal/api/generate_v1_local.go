package api

import (
	"fmt"
	"os"
	"path/filepath"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func localV1DependencyRoots(moduleDir string, module v1.Module, visiting map[string]bool) ([]string, error) {
	roots, err := localV1DependencySources(moduleDir, module, visiting)
	if err != nil {
		return nil, err
	}
	return roots.paths(), nil
}

func localV1DependencySources(moduleDir string, module v1.Module, visiting map[string]bool) (v1SourceRoots, error) {
	if visiting[moduleDir] {
		return nil, fmt.Errorf("local dependency cycle at %s", moduleDir)
	}
	visiting[moduleDir] = true
	defer delete(visiting, moduleDir)

	replaces := make(map[string]string, len(module.Replaces))
	for _, replacement := range module.Replaces {
		replaces[replacement.Module] = replacement.Target
	}
	var roots v1SourceRoots
	for _, requirement := range module.Requires {
		target, ok := replaces[requirement.Module]
		if !ok {
			continue // resolved from this module's protobuf.lock
		}
		depDir := filepath.Clean(filepath.Join(moduleDir, target))
		dep, err := moduleconfig.ReadGitDependency(depDir, requirement.Module)
		if err != nil {
			return nil, fmt.Errorf("local replacement %s: %w", target, err)
		}
		for _, root := range dep.Roots {
			path := filepath.Join(depDir, root)
			info, err := os.Stat(path)
			if err != nil || !info.IsDir() {
				return nil, fmt.Errorf("dependency %s has invalid root %q", dep.Name, root)
			}
			roots = append(roots, v1SourceRoot{path: path, module: dep.Name})
		}
		transitive, err := localV1DependencySources(depDir, dep, visiting)
		if err != nil {
			return nil, err
		}
		roots = append(roots, transitive...)
	}
	return roots, nil
}
