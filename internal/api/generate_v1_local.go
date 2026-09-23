package api

import (
	"fmt"
	"path/filepath"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

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
			continue
		}
		depDir := filepath.Clean(filepath.Join(moduleDir, target))
		dep, err := moduleconfig.ReadGitDependency(depDir, requirement.Module)
		if err != nil {
			return nil, fmt.Errorf("local replacement %s: %w", target, err)
		}
		dependencyRoots, err := moduleV1SourceRoots(depDir, dep)
		if err != nil {
			return nil, fmt.Errorf("moduleV1SourceRoots: %w", err)
		}
		roots = append(roots, dependencyRoots...)
		transitive, err := localV1DependencySources(depDir, dep, visiting)
		if err != nil {
			return nil, err
		}
		roots = append(roots, transitive...)
	}
	return roots, nil
}
