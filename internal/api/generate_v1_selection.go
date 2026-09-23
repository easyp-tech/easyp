package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
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
	dependencies v1SourceRoots
}

func generateSelectedV1Module(ctx *cli.Context, log logger.Logger, configPath, repoRoot string, selection v1ModuleSelection, gen v1.Generate) error {
	selected, err := resolveV1GenerationModule(ctx.Context, log, repoRoot, filepath.Dir(configPath), selection)
	if err != nil {
		return fmt.Errorf("resolveV1GenerationModule: %w", err)
	}
	return generateV1ModuleWithRoots(ctx, log, configPath, selected.directory, gen, selected.module, selected.dependencies)
}

func resolveV1GenerationModule(ctx context.Context, log logger.Logger, repoRoot, configDir string, selection v1ModuleSelection) (v1GenerationModule, error) {
	if selection.directory != "" {
		return readV1GenerationModule(ctx, log, selection.directory)
	}
	localDir, err := findV1LocalModuleByName(repoRoot, selection.source)
	if err != nil {
		return v1GenerationModule{}, fmt.Errorf("findV1LocalModuleByName: %w", err)
	}
	if localDir != "" {
		return readV1GenerationModule(ctx, log, localDir)
	}
	consumerDir, err := generatorV1ModuleDir(repoRoot, configDir)
	if err != nil {
		return v1GenerationModule{}, fmt.Errorf("generatorV1ModuleDir: %w", err)
	}
	_, consumer, err := readV1Manifest(consumerDir)
	if err != nil {
		return v1GenerationModule{}, fmt.Errorf("readV1Manifest: %w", err)
	}
	if consumer.Name == selection.source {
		return readV1GenerationModule(ctx, log, consumerDir)
	}

	var moduleDir string
	var dependencyRoots v1SourceRoots
	for _, replacement := range consumer.Replaces {
		if replacement.Module != selection.source {
			continue
		}
		dependencyRoots, err = resolveV1DependencySources(ctx, log, consumerDir, consumer)
		if err != nil {
			return v1GenerationModule{}, fmt.Errorf("resolveV1DependencySources: %w", err)
		}
		moduleDir = filepath.Clean(filepath.Join(consumerDir, replacement.Target))
		break
	}
	if moduleDir == "" {
		gitCacheRoot, err := gitCachePath(log)
		if err != nil {
			return v1GenerationModule{}, fmt.Errorf("gitCachePath: %w", err)
		}
		dependencyRoots, err = lockedV1DependencySources(ctx, consumerDir, consumer, gitCacheRoot)
		if err != nil {
			return v1GenerationModule{}, fmt.Errorf("lockedV1DependencySources: %w", err)
		}
		lock, err := readV1Lock(filepath.Join(consumerDir, v1.LockFile))
		if err != nil {
			return v1GenerationModule{}, fmt.Errorf("readV1Lock: %w", err)
		}
		for _, entry := range lock.Modules {
			if entry.Source == selection.source {
				moduleDir = v1ModuleCachePath(gitCacheRoot, entry)
				break
			}
		}
		if moduleDir == "" {
			return v1GenerationModule{}, fmt.Errorf("module %q is not selected in %s", selection.source, filepath.Join(consumerDir, v1.LockFile))
		}
		localRoots, err := localV1DependencySources(consumerDir, consumer, map[string]bool{})
		if err != nil {
			return v1GenerationModule{}, fmt.Errorf("localV1DependencySources: %w", err)
		}
		dependencyRoots = append(dependencyRoots, localRoots...)
	}

	module, err := moduleconfig.ReadGitDependency(moduleDir, selection.source)
	if err != nil {
		return v1GenerationModule{}, fmt.Errorf("ReadGitDependency: %w", err)
	}
	otherRoots := make(v1SourceRoots, 0, len(dependencyRoots))
	for _, root := range dependencyRoots {
		if root.module != selection.source {
			otherRoots = append(otherRoots, root)
		}
	}
	return v1GenerationModule{directory: moduleDir, module: module, dependencies: otherRoots}, nil
}

func readV1GenerationModule(ctx context.Context, log logger.Logger, directory string) (v1GenerationModule, error) {
	_, module, err := readV1Manifest(directory)
	if err != nil {
		return v1GenerationModule{}, fmt.Errorf("readV1Manifest: %w", err)
	}
	dependencies, err := resolveV1DependencySources(ctx, log, directory, module)
	if err != nil {
		return v1GenerationModule{}, fmt.Errorf("resolveV1DependencySources: %w", err)
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
			if path != repoRoot && (strings.HasPrefix(entry.Name(), ".") || entry.Name() == defaultVendorDir) {
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
