package generation

import (
	"context"
	"fmt"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/modules"
)

func generateV1ModuleWithRoots(ctx context.Context, log logger.Logger, request Request, configPath, moduleDir string, gen v1.Generate, module v1.Module, importRoots modules.SourceRoots) error {
	app, _, err := prepareV1ModuleCore(log, request, configPath, moduleDir, gen, module, importRoots, nil)
	if err != nil {
		return fmt.Errorf("prepareV1ModuleCore: %w", err)
	}
	if err := app.Generate(ctx, moduleDir, request.DescriptorSetOut, request.IncludeImports); err != nil {
		return fmt.Errorf("Generate: %w", err)
	}
	return nil
}

func prepareV1ModuleCore(log logger.Logger, request Request, configPath, moduleDir string, gen v1.Generate, module v1.Module, importRoots, sources modules.SourceRoots) (*core.Core, map[string]string, error) {
	if sources == nil {
		var err error
		sources, err = modules.ModuleSources(moduleDir, module)
		if err != nil {
			return nil, nil, fmt.Errorf("ModuleSources: %w", err)
		}
	}
	allSources := append(append(modules.SourceRoots(nil), sources...), importRoots...)
	// The consumer retains ownership when the same physical root is also
	// present in its dependency graph. Indexing also validates every import.
	indexRoots := append(append(modules.SourceRoots(nil), importRoots...), sources...)
	owners, err := indexRoots.IndexFileModules()
	if err != nil {
		return nil, nil, fmt.Errorf("IndexFileModules: %w", err)
	}
	options, err := prepareV1GeneratorConfig(configPath, moduleDir, gen, module)
	if err != nil {
		return nil, nil, fmt.Errorf("prepareV1GeneratorConfig: %w", err)
	}
	options.Logger = log
	options.PluginWorkDir = request.WorkDir
	options.ImportRoots = importRoots.Paths()
	options.ImportFileAllowed = allSources.FileAllowed()
	options.OpenSourceFile = allSources.OpenSourceFile
	options.WalkSourceFiles = sources.WalkSnapshotFiles
	if options.ManagedModeConfig.Enabled || options.ManagedModeConfig.GoPackageOnly {
		options.FileModules = owners
	}
	return core.New(options), owners, nil
}
