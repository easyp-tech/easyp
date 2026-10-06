package modules

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

type localReplacements struct {
	directory string
	targets   map[string]string
	modules   map[string]EffectiveModule
	paths     map[string]string
	localPath func(string) (string, error)
}

func newLocalReplacements(directory string, module v1.Module, localPath func(string) (string, error)) *localReplacements {
	targets := make(map[string]string, len(module.Replaces))
	for _, replacement := range module.Replaces {
		targets[replacement.Module] = replacement.Target
	}
	return &localReplacements{directory: directory, targets: targets, modules: make(map[string]EffectiveModule), paths: make(map[string]string), localPath: localPath}
}

func (l *localReplacements) lookupRequirement(name, version string) (v1.Module, bool, error) {
	if err := v1.ValidateModuleVersion(name, version); err != nil {
		return v1.Module{}, false, fmt.Errorf("ValidateModuleVersion: %w", err)
	}
	// A local replacement has no published version. Go allows a native local
	// module to replace a legacy +incompatible requirement; provenance checks
	// belong to published Fetch/Install/Cached and the overlay never writes lock.
	return l.lookup(name)
}

func (l *localReplacements) lookup(name string) (v1.Module, bool, error) {
	target, ok := l.targets[name]
	if !ok {
		return v1.Module{}, false, nil
	}
	if entry, ok := l.modules[name]; ok {
		return entry.Module, true, nil
	}
	directory := ResolveReplacementPath(l.directory, target)
	if l.localPath != nil {
		var err error
		directory, err = l.localPath(target)
		if err != nil {
			return v1.Module{}, true, fmt.Errorf("local replacement %s => %s: %w", name, target, err)
		}
	}
	entry, err := readLocalReplacement(directory, name)
	if err != nil {
		return v1.Module{}, true, fmt.Errorf("local replacement %s => %s: %w", name, target, err)
	}
	canonical, err := filepath.EvalSymlinks(directory)
	if err != nil {
		return v1.Module{}, true, fmt.Errorf("EvalSymlinks: %w", err)
	}
	if other, ok := l.paths[canonical]; ok && other != name {
		return v1.Module{}, true, fmt.Errorf("local replacement %s => %s: directory already supplies module %s", name, target, other)
	}
	l.paths[canonical] = name
	l.modules[name] = entry
	return entry.Module, true, nil
}

func readLocalReplacement(directory, name string) (EffectiveModule, error) {
	info, err := os.Stat(directory)
	if err != nil {
		return EffectiveModule{}, fmt.Errorf("Stat: %w", err)
	}
	if !info.IsDir() {
		return EffectiveModule{}, fmt.Errorf("%s is not a directory", directory)
	}
	raw, err := sourceview.ReadLocal(context.Background(), directory, v1.ModuleFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return EffectiveModule{}, fmt.Errorf("ReadFile: %w", err)
	}
	var module v1.Module
	if v1.IsModuleManifest(raw) {
		module, err = v1.ParseModule(bytes.NewReader(raw))
		if err != nil {
			return EffectiveModule{}, fmt.Errorf("ParseModule: %w", err)
		}
		if module.Name != name {
			return EffectiveModule{}, fmt.Errorf("%s declares module %s, want %s", directory, module.Name, name)
		}
	} else {
		// Compatibility reads only roots/requirements, never dependency plugins.
		module, err = moduleconfig.ReadLocalDependency(directory, name)
		if err != nil {
			return EffectiveModule{}, fmt.Errorf("ReadGitDependency: %w", err)
		}
	}
	if _, err := ModuleSources(directory, module); err != nil {
		return EffectiveModule{}, fmt.Errorf("ModuleSources: %w", err)
	}
	return EffectiveModule{Directory: directory, Module: module}, nil
}

func localDependencySources(moduleDir string, module v1.Module) (SourceRoots, error) {
	locals := newLocalReplacements(moduleDir, module, nil)
	visited := map[string]bool{module.Name: true}
	queue := slices.Clone(module.Requires)
	var roots SourceRoots
	for len(queue) > 0 {
		requirement := queue[0]
		queue = queue[1:]
		if visited[requirement.Module] {
			continue
		}
		visited[requirement.Module] = true
		dep, local, err := locals.lookupRequirement(requirement.Module, requirement.Version)
		if err != nil {
			return nil, err
		}
		if !local {
			continue
		}
		source, err := ModuleSources(locals.modules[requirement.Module].Directory, dep)
		if err != nil {
			return nil, fmt.Errorf("ModuleSources: %w", err)
		}
		roots = append(roots, source...)
		queue = append(queue, dep.Requires...)
	}
	return roots, nil
}

// ResolveReplacementPath resolves relative replacements from their owning manifest.
// Absolute replacement paths are used directly.
func ResolveReplacementPath(moduleDir, target string) string {
	if filepath.IsAbs(target) {
		return filepath.Clean(target)
	}
	return filepath.Clean(filepath.Join(moduleDir, target))
}
