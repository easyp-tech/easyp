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

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/modules"
	"github.com/easyp-tech/easyp/internal/sourceview"
	"github.com/easyp-tech/easyp/internal/workspace"
)

// v1ModuleSelection distinguishes workspace directories from module identities.
// Absolute module identities can refer to a local Git repository.
type v1ModuleSelection struct {
	directory string
	source    string
	index     int
	packages  []string
	paths     []string
}

func selectV1Modules(repoRoot, configDir string, entries []v1.GenerateModule) ([]v1ModuleSelection, error) {
	if len(entries) == 0 {
		dir, err := generatorV1ModuleDir(repoRoot, configDir)
		if err != nil {
			return nil, fmt.Errorf("generatorV1ModuleDir: %w", err)
		}
		return []v1ModuleSelection{{directory: dir}}, nil
	}
	result := make([]v1ModuleSelection, 0, len(entries))
	for i, entry := range entries {
		name := entry.Module
		selection := v1ModuleSelection{source: name, index: i, packages: entry.Packages, paths: entry.Paths}
		if filepath.IsAbs(name) {
			result = append(result, selection)
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
		_, err = sourceview.ResolveLocal(context.Background(), repoRoot, filepath.Join(rel, v1.ModuleFile))
		if errors.Is(err, os.ErrNotExist) {
			if name == "." || name == ".." || strings.HasPrefix(name, "./") || strings.HasPrefix(name, "../") {
				return nil, fmt.Errorf("generate.modules entry %q is a local path, but %s has no %s; point to a workspace module directory containing %s, use its module identity, or omit generate.modules to select the module containing this generator", name, path, v1.ModuleFile, v1.ModuleFile)
			}
			result = append(result, selection)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("Stat: %w", err)
		}
		selection.directory, selection.source = path, ""
		result = append(result, selection)
	}
	return result, nil
}

func generatorV1ModuleDir(repoRoot, configDir string) (string, error) {
	// Explicit projects outside the workspace keep their own local context.
	relative, err := filepath.Rel(repoRoot, configDir)
	if err != nil {
		return "", err
	}
	if !filepath.IsLocal(relative) {
		return configDir, nil
	}
	path, err := workspace.FindUp(configDir, repoRoot, v1.ModuleFile)
	if err != nil {
		return "", err
	}
	if path != "" {
		return filepath.Dir(path), nil
	}
	return configDir, nil
}

type v1GenerationModule struct {
	resolutionDir string
	directory     string
	module        v1.Module
	dependencies  modules.SourceRoots
}

func generateSelectedV1Module(ctx context.Context, log logger.Logger, cache modules.Cache, request Request, configPath, repoRoot string, selection v1ModuleSelection, gen v1.Generate) error {
	selected, err := resolveV1GenerationModule(ctx, cache, repoRoot, filepath.Dir(configPath), selection, request.Frozen)
	if err != nil {
		return fmt.Errorf("resolveV1GenerationModule: %w", err)
	}
	return generateV1ModuleWithRoots(ctx, log, request, configPath, selected.directory, gen, selected.module, selected.dependencies)
}

func resolveV1GenerationModule(ctx context.Context, cache modules.Cache, repoRoot, configDir string, selection v1ModuleSelection, frozen bool) (v1GenerationModule, error) {
	if frozen {
		return resolveFrozenGenerationModule(ctx, cache, repoRoot, configDir, selection)
	}
	if selection.directory != "" {
		return readV1GenerationModule(ctx, cache, selection.directory)
	}
	localDir, err := findV1LocalModuleByName(repoRoot, selection.source)
	if err != nil {
		return v1GenerationModule{}, fmt.Errorf("findV1LocalModuleByName: %w", err)
	}
	consumerDir, err := generatorV1ModuleDir(repoRoot, configDir)
	if err != nil {
		return v1GenerationModule{}, fmt.Errorf("generatorV1ModuleDir: %w", err)
	}
	_, consumer, err := modules.ReadManifest(consumerDir)
	if err != nil {
		if localDir != "" {
			return readV1GenerationModule(ctx, cache, localDir)
		}
		return v1GenerationModule{}, fmt.Errorf("ReadManifest: %w", err)
	}
	if consumer.Name == selection.source {
		return readV1GenerationModule(ctx, cache, consumerDir)
	}

	if len(consumer.Replaces) > 0 {
		graph, err := modules.EnsureEffectiveGraph(ctx, consumerDir, consumer, cache, nil)
		if err != nil {
			return v1GenerationModule{}, fmt.Errorf("EnsureEffectiveGraph: %w", err)
		}
		if selected, ok := graph.Modules[selection.source]; ok {
			var others modules.SourceRoots
			for _, root := range graph.Sources {
				if root.Module != selection.source {
					others = append(others, root)
				}
			}
			return v1GenerationModule{directory: selected.Directory, resolutionDir: consumerDir, module: selected.Module, dependencies: others}, nil
		}
		for _, replacement := range consumer.Replaces {
			if replacement.Module == selection.source {
				return v1GenerationModule{}, fmt.Errorf("module %q is not selected in the effective graph", selection.source)
			}
		}
		if localDir == "" {
			return v1GenerationModule{}, fmt.Errorf("module %q is not selected in the effective graph", selection.source)
		}
	}
	if localDir != "" {
		return readV1GenerationModule(ctx, cache, localDir)
	}

	var moduleDir string
	var module v1.Module
	var dependencyRoots modules.SourceRoots
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

	otherRoots := make(modules.SourceRoots, 0, len(dependencyRoots))
	for _, root := range dependencyRoots {
		if root.Module != selection.source {
			otherRoots = append(otherRoots, root)
		}
	}
	return v1GenerationModule{directory: moduleDir, resolutionDir: consumerDir, module: module, dependencies: otherRoots}, nil
}

func readV1GenerationModule(ctx context.Context, cache modules.Cache, directory string) (v1GenerationModule, error) {
	module, err := modules.ReadModuleOrDefault(directory)
	if err != nil {
		return v1GenerationModule{}, fmt.Errorf("ReadModuleOrDefault: %w", err)
	}
	dependencies, err := modules.EnsureSources(ctx, directory, module, cache)
	if err != nil {
		return v1GenerationModule{}, fmt.Errorf("EnsureSources: %w", err)
	}
	return v1GenerationModule{directory: directory, resolutionDir: directory, module: module, dependencies: dependencies}, nil
}

func findV1LocalModuleByName(repoRoot, name string) (string, error) {
	var found string
	err := workspace.Walk(repoRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// Module lookup reads declarations; source aliases are validated after selection.
			if filepath.Base(path) != v1.ModuleFile {
				return nil
			}
			return walkErr
		}
		if entry.IsDir() {
			if workspace.SkipDirectory(repoRoot, path) {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Name() != v1.ModuleFile {
			return nil
		}
		manifest, err := workspace.ReadFile(repoRoot, path)
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
