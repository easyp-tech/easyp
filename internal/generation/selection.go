package generation

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/modules"
)

// v1ModuleSelection distinguishes workspace directories from module identities.
// Absolute module identities can refer to a local Git repository.
type v1ModuleSelection struct {
	directory string
	source    string
}

func selectV1Modules(repoRoot, configDir string, names []string) ([]v1ModuleSelection, error) {
	if len(names) == 0 {
		dir, err := generatorV1ModuleDir(repoRoot, configDir)
		if err != nil {
			return nil, fmt.Errorf("generatorV1ModuleDir: %w", err)
		}
		return []v1ModuleSelection{{directory: dir}}, nil
	}
	result := make([]v1ModuleSelection, 0, len(names))
	for _, name := range names {
		if filepath.IsAbs(name) {
			result = append(result, v1ModuleSelection{source: name})
			continue
		}
		path := filepath.Clean(filepath.Join(repoRoot, name))
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return nil, fmt.Errorf("Rel: %w", err)
		}
		if !filepath.IsLocal(rel) {
			return nil, fmt.Errorf("module path %q leaves repository", name)
		}
		_, err = os.Stat(filepath.Join(path, v1.ModuleFile))
		if errors.Is(err, os.ErrNotExist) {
			result = append(result, v1ModuleSelection{source: name})
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("Stat: %w", err)
		}
		result = append(result, v1ModuleSelection{directory: path})
	}
	return result, nil
}

func generatorV1ModuleDir(repoRoot, configDir string) (string, error) {
	for _, dir := range []string{configDir, repoRoot} {
		_, err := os.Stat(filepath.Join(dir, v1.ModuleFile))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("Stat: %w", err)
		}
		return dir, nil
	}
	return "", fmt.Errorf("no protobuf.mod for generator in %s", configDir)
}

type v1GenerationModule struct {
	directory    string
	module       v1.Module
	dependencies modules.SourceRoots
}

func generateSelectedV1Module(ctx context.Context, log logger.Logger, cache modules.Cache, request Request, configPath, repoRoot string, selection v1ModuleSelection, gen v1.Generate) error {
	selected, err := resolveV1GenerationModule(ctx, cache, repoRoot, filepath.Dir(configPath), selection)
	if err != nil {
		return fmt.Errorf("resolveV1GenerationModule: %w", err)
	}
	return generateV1ModuleWithRoots(ctx, log, request, configPath, selected.directory, gen, selected.module, selected.dependencies)
}

func resolveV1GenerationModule(ctx context.Context, cache modules.Cache, repoRoot, configDir string, selection v1ModuleSelection) (v1GenerationModule, error) {
	if selection.directory != "" {
		return readV1GenerationModule(ctx, cache, selection.directory)
	}
	localDir, err := findV1LocalModuleByName(repoRoot, selection.source)
	if err != nil {
		return v1GenerationModule{}, fmt.Errorf("findV1LocalModuleByName: %w", err)
	}
	if localDir != "" {
		return readV1GenerationModule(ctx, cache, localDir)
	}
	consumerDir, err := generatorV1ModuleDir(repoRoot, configDir)
	if err != nil {
		return v1GenerationModule{}, fmt.Errorf("generatorV1ModuleDir: %w", err)
	}
	_, consumer, err := modules.ReadManifest(consumerDir)
	if err != nil {
		return v1GenerationModule{}, fmt.Errorf("ReadManifest: %w", err)
	}
	if consumer.Name == selection.source {
		return readV1GenerationModule(ctx, cache, consumerDir)
	}

	var moduleDir string
	var module v1.Module
	var dependencyRoots modules.SourceRoots
	for _, replacement := range consumer.Replaces {
		if replacement.Module != selection.source {
			continue
		}
		dependencyRoots, err = modules.EnsureSources(ctx, consumerDir, consumer, cache)
		if err != nil {
			return v1GenerationModule{}, fmt.Errorf("EnsureSources: %w", err)
		}
		moduleDir = filepath.Clean(filepath.Join(consumerDir, replacement.Target))
		module, err = moduleconfig.ReadGitDependency(moduleDir, selection.source)
		if err != nil {
			return v1GenerationModule{}, fmt.Errorf("ReadGitDependency: %w", err)
		}
		break
	}
	if moduleDir == "" {
		dependencyRoots, err = modules.EnsureLockedSources(ctx, consumerDir, consumer, cache)
		if err != nil {
			return v1GenerationModule{}, fmt.Errorf("EnsureLockedSources: %w", err)
		}
		lock, err := modules.ReadLock(filepath.Join(consumerDir, v1.LockFile))
		if err != nil {
			return v1GenerationModule{}, fmt.Errorf("ReadLock: %w", err)
		}
		for _, entry := range lock.Modules {
			if entry.Source == selection.source {
				moduleDir, module, err = cache.Cached(entry)
				if err != nil {
					return v1GenerationModule{}, fmt.Errorf("Cached: %w", err)
				}
				break
			}
		}
		if moduleDir == "" {
			return v1GenerationModule{}, fmt.Errorf("module %q is not selected in %s", selection.source, filepath.Join(consumerDir, v1.LockFile))
		}
		localRoots, err := modules.LocalSources(consumerDir, consumer)
		if err != nil {
			return v1GenerationModule{}, fmt.Errorf("LocalSources: %w", err)
		}
		dependencyRoots = append(dependencyRoots, localRoots...)
	}

	otherRoots := make(modules.SourceRoots, 0, len(dependencyRoots))
	for _, root := range dependencyRoots {
		if root.Module != selection.source {
			otherRoots = append(otherRoots, root)
		}
	}
	return v1GenerationModule{directory: moduleDir, module: module, dependencies: otherRoots}, nil
}

func readV1GenerationModule(ctx context.Context, cache modules.Cache, directory string) (v1GenerationModule, error) {
	_, module, err := modules.ReadManifest(directory)
	if err != nil {
		return v1GenerationModule{}, fmt.Errorf("ReadManifest: %w", err)
	}
	dependencies, err := modules.EnsureSources(ctx, directory, module, cache)
	if err != nil {
		return v1GenerationModule{}, fmt.Errorf("EnsureSources: %w", err)
	}
	return v1GenerationModule{directory: directory, module: module, dependencies: dependencies}, nil
}

func findV1LocalModuleByName(repoRoot, name string) (string, error) {
	var found string
	err := filepath.WalkDir(repoRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != repoRoot && (strings.HasPrefix(entry.Name(), ".") || entry.Name() == modules.VendorDir) {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Name() != v1.ModuleFile {
			return nil
		}
		manifest, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("ReadFile: %w", err)
		}
		if !v1.IsModuleManifest(manifest) {
			return nil
		}
		module, parseErr := v1.ParseModule(bytes.NewReader(manifest))
		if parseErr != nil {
			return fmt.Errorf("%s: %w", path, parseErr)
		}
		if module.Name == name {
			if found != "" {
				return fmt.Errorf("module identity %q is declared in both %s and %s", name, found, filepath.Dir(path))
			}
			found = filepath.Dir(path)
		}
		return nil
	})
	return found, err
}
