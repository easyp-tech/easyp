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
	app, err := prepareV1ModuleCore(log, request, configPath, moduleDir, gen, module, importRoots)
	if err != nil {
		return fmt.Errorf("prepareV1ModuleCore: %w", err)
	}
	if err := app.Generate(ctx, moduleDir, request.DescriptorSetOut, request.IncludeImports); err != nil {
		return fmt.Errorf("Generate: %w", err)
	}
	return nil
}

func prepareV1ModuleCore(log logger.Logger, request Request, configPath, moduleDir string, gen v1.Generate, module v1.Module, importRoots modules.SourceRoots) (*core.Core, error) {
	sources, err := modules.ModuleSources(moduleDir, module)
	if err != nil {
		return nil, fmt.Errorf("ModuleSources: %w", err)
	}
	allSources := append(append(modules.SourceRoots(nil), sources...), importRoots...)
	if err := modules.CheckSourceCollisions(allSources); err != nil {
		return nil, fmt.Errorf("CheckSourceCollisions: %w", err)
	}
	options, err := prepareV1GeneratorConfig(configPath, moduleDir, gen, module)
	if err != nil {
		return nil, fmt.Errorf("prepareV1GeneratorConfig: %w", err)
	}
	options.Logger = log
	options.PluginWorkDir = request.WorkDir
	options.ImportRoots = importRoots.Paths()
	options.ImportFileAllowed = allSources.FileAllowed()
	if options.ManagedModeConfig.Enabled || options.ManagedModeConfig.GoPackageOnly {
		roots := append(modules.SourceRoots(nil), importRoots...)
		roots = append(roots, sources...)
		options.FileModules, err = roots.FileModules()
		if err != nil {
			return nil, fmt.Errorf("FileModules: %w", err)
		}
	}
	return core.New(options), nil
}
