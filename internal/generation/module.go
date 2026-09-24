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
	sources, err := modules.ModuleSources(moduleDir, module)
	if err != nil {
		return fmt.Errorf("ModuleSources: %w", err)
	}
	if err := modules.CheckImportCollisions(moduleDir, module.Roots, importRoots.Paths()); err != nil {
		return fmt.Errorf("CheckImportCollisions: %w", err)
	}
	options, err := prepareV1GeneratorConfig(configPath, moduleDir, gen, module)
	if err != nil {
		return fmt.Errorf("prepareV1GeneratorConfig: %w", err)
	}
	options.Logger = log
	options.PluginWorkDir = request.WorkDir
	options.ImportRoots = importRoots.Paths()
	if options.ManagedModeConfig.Enabled {
		roots := append(modules.SourceRoots(nil), importRoots...)
		roots = append(roots, sources...)
		options.FileModules, err = roots.FileModules()
		if err != nil {
			return fmt.Errorf("FileModules: %w", err)
		}
	}
	app := core.New(options)
	if err := app.Generate(ctx, moduleDir, request.DescriptorSetOut, request.IncludeImports); err != nil {
		return fmt.Errorf("Generate: %w", err)
	}
	return nil
}
