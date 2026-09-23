package api

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
)

func resolveV1DependencySources(ctx context.Context, log logger.Logger, moduleDir string, module v1.Module) (v1SourceRoots, error) {
	roots, err := localV1DependencySources(moduleDir, module, map[string]bool{})
	if err != nil {
		return nil, fmt.Errorf("localV1DependencySources: %w", err)
	}
	cacheRoot, err := gitCachePath(log)
	if err != nil {
		return nil, fmt.Errorf("gitCachePath: %w", err)
	}
	lockedRoots, err := lockedV1DependencySources(ctx, moduleDir, module, cacheRoot)
	if err != nil {
		return nil, fmt.Errorf("lockedV1DependencySources: %w", err)
	}
	return append(roots, lockedRoots...), nil
}

func generateV1ModuleWithRoots(ctx *cli.Context, log logger.Logger, configPath, moduleDir string, gen v1.Generate, module v1.Module, importRoots v1SourceRoots) error {
	sources, err := moduleV1SourceRoots(moduleDir, module)
	if err != nil {
		return fmt.Errorf("moduleV1SourceRoots: %w", err)
	}
	if err := checkV1ImportPathCollisions(moduleDir, module.Roots, importRoots.paths()); err != nil {
		return fmt.Errorf("checkV1ImportPathCollisions: %w", err)
	}

	cfg, err := prepareV1GeneratorConfig(configPath, moduleDir, gen, module)
	if err != nil {
		return fmt.Errorf("prepareV1GeneratorConfig: %w", err)
	}
	app, err := buildCore(log, cfg)
	if err != nil {
		return fmt.Errorf("buildCore: %w", err)
	}
	app.SetImportRoots(importRoots.paths())
	if cfg.Generate.Managed.Enabled {
		moduleRoots := append(v1SourceRoots(nil), importRoots...)
		moduleRoots = append(moduleRoots, sources...)
		fileModules, err := moduleRoots.fileModules()
		if err != nil {
			return fmt.Errorf("fileModules: %w", err)
		}
		app.SetFileModules(fileModules)
	}
	if err := app.Generate(ctx.Context, moduleDir, ctx.String(flagGenerateDescriptorSetOut.Name), ctx.Bool(flagGenerateIncludeImports.Name)); err != nil {
		return fmt.Errorf("Generate: %w", err)
	}
	return nil
}
