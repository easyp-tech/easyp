package generation

import (
	"context"
	"fmt"
	"path/filepath"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

func readFrozenGenerationModule(ctx context.Context, cache modules.Cache, directory string) (v1GenerationModule, error) {
	dependencies, err := modules.EnsureFrozenSources(ctx, directory, cache)
	if err != nil {
		return v1GenerationModule{}, fmt.Errorf("EnsureFrozenSources: %w", err)
	}
	_, module, err := modules.ReadManifest(directory)
	if err != nil {
		return v1GenerationModule{}, fmt.Errorf("ReadManifest: %w", err)
	}
	return v1GenerationModule{directory: directory, resolutionDir: directory, module: module, dependencies: dependencies}, nil
}

func resolveFrozenGenerationModule(ctx context.Context, cache modules.Cache, repoRoot, configDir string, selection v1ModuleSelection) (v1GenerationModule, error) {
	if selection.directory != "" {
		return readFrozenGenerationModule(ctx, cache, selection.directory)
	}
	consumerDir, err := generatorV1ModuleDir(repoRoot, configDir)
	if err != nil {
		return v1GenerationModule{}, fmt.Errorf("generatorV1ModuleDir: %w", err)
	}
	consumer, err := readFrozenGenerationModule(ctx, cache, consumerDir)
	if err != nil {
		return v1GenerationModule{}, err
	}
	if consumer.module.Name == selection.source {
		return consumer, nil
	}
	lock, err := modules.ReadLock(filepath.Join(consumerDir, v1.LockFile))
	if err != nil {
		return v1GenerationModule{}, fmt.Errorf("ReadLock: %w", err)
	}
	// A dependency is validated against the consumer graph. Its checkout need
	// not contain a lock and dependency-owned replacement directives are ignored.
	for _, entry := range lock.Modules {
		if entry.Source != selection.source {
			continue
		}
		directory, module, err := cache.Cached(entry)
		if err != nil {
			return v1GenerationModule{}, fmt.Errorf("Cached: %w", err)
		}
		var others modules.SourceRoots
		for _, root := range consumer.dependencies {
			if root.Module != selection.source {
				others = append(others, root)
			}
		}
		return v1GenerationModule{directory: directory, resolutionDir: consumerDir, module: module, dependencies: others}, nil
	}
	localDir, err := findV1LocalModuleByName(repoRoot, selection.source)
	if err != nil {
		return v1GenerationModule{}, fmt.Errorf("findV1LocalModuleByName: %w", err)
	}
	if localDir != "" {
		return readFrozenGenerationModule(ctx, cache, localDir)
	}
	return v1GenerationModule{}, fmt.Errorf("frozen module %q is not selected in %s", selection.source, filepath.Join(consumerDir, v1.LockFile))
}
