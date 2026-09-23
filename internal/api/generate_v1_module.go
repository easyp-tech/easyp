package api

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/urfave/cli/v2"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	"github.com/easyp-tech/easyp/internal/config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
)

func generateSelectedV1Module(ctx *cli.Context, log logger.Logger, configPath, repoRoot, selection string, gen v1.Generate) error {
	if filepath.IsAbs(selection) && !slices.Contains(gen.Generate.Modules, selection) {
		return generateV1Module(ctx, log, configPath, selection, gen)
	}
	localDir, err := findV1LocalModuleByName(repoRoot, selection)
	if err != nil {
		return err
	}
	if localDir != "" {
		return generateV1Module(ctx, log, configPath, localDir, gen)
	}
	consumerDirs, err := selectedV1ModuleDirs(repoRoot, filepath.Dir(configPath), nil)
	if err != nil {
		return fmt.Errorf("resolve selected module %q: %w", selection, err)
	}
	consumerDir := consumerDirs[0]
	manifest, err := os.Open(filepath.Join(consumerDir, "protobuf.mod"))
	if err != nil {
		return err
	}
	consumer, parseErr := v1.ParseModule(manifest)
	_ = manifest.Close()
	if parseErr != nil {
		return parseErr
	}
	if consumer.Name == selection {
		return generateV1Module(ctx, log, configPath, consumerDir, gen)
	}
	for _, replacement := range consumer.Replaces {
		if replacement.Module == selection {
			roots, err := resolveV1DependencySources(ctx.Context, log, consumerDir, consumer)
			if err != nil {
				return err
			}
			moduleDir := filepath.Clean(filepath.Join(consumerDir, replacement.Target))
			return generateV1DependencyModule(ctx, log, configPath, moduleDir, selection, gen, roots)
		}
	}
	cacheBase, err := getEasypPath(log)
	if err != nil {
		return err
	}
	callCtx := ctx.Context
	if callCtx == nil {
		callCtx = context.Background()
	}
	gitCacheRoot := filepath.Join(cacheBase, "v1", "git")
	dependencyRoots, err := lockedV1DependencySources(callCtx, consumerDir, consumer, gitCacheRoot)
	if err != nil {
		return err
	}
	lock, err := readV1Lock(filepath.Join(consumerDir, "protobuf.lock"))
	if err != nil {
		return err
	}
	for _, entry := range lock.Modules {
		if entry.Source == selection {
			moduleDir := v1ModuleCachePath(gitCacheRoot, entry)
			localRoots, err := localV1DependencySources(consumerDir, consumer, map[string]bool{})
			if err != nil {
				return err
			}
			dependencyRoots = append(dependencyRoots, localRoots...)
			return generateV1DependencyModule(ctx, log, configPath, moduleDir, selection, gen, dependencyRoots)
		}
	}
	return fmt.Errorf("module %q is not selected in %s/protobuf.lock", selection, consumerDir)
}

func findV1LocalModuleByName(repoRoot, name string) (string, error) {
	var found string
	err := filepath.WalkDir(repoRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != repoRoot && (strings.HasPrefix(entry.Name(), ".") || entry.Name() == "easyp_vendor") {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Name() != "protobuf.mod" {
			return nil
		}
		manifest, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !v1.IsModuleManifest(manifest) {
			return nil
		}
		module, parseErr := v1.ParseModule(strings.NewReader(string(manifest)))
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

func generateV1Module(ctx *cli.Context, log logger.Logger, configPath, moduleDir string, gen v1.Generate) error {
	fp, err := os.Open(filepath.Join(moduleDir, "protobuf.mod"))
	if err != nil {
		return err
	}
	module, parseErr := v1.ParseModule(fp)
	_ = fp.Close()
	if parseErr != nil {
		return fmt.Errorf("%s/protobuf.mod: %w", moduleDir, parseErr)
	}
	importRoots, err := resolveV1DependencySources(ctx.Context, log, moduleDir, module)
	if err != nil {
		return err
	}
	return generateV1ModuleWithRoots(ctx, log, configPath, moduleDir, gen, module, importRoots)
}

func resolveV1DependencySources(ctx context.Context, log logger.Logger, moduleDir string, module v1.Module) (v1SourceRoots, error) {
	roots, err := localV1DependencySources(moduleDir, module, map[string]bool{})
	if err != nil {
		return nil, fmt.Errorf("localV1DependencySources: %w", err)
	}
	cacheBase, err := getEasypPath(log)
	if err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	lockedRoots, err := lockedV1DependencySources(ctx, moduleDir, module, filepath.Join(cacheBase, "v1", "git"))
	if err != nil {
		return nil, fmt.Errorf("lockedV1DependencySources: %w", err)
	}
	return append(roots, lockedRoots...), nil
}

func generateV1DependencyModule(ctx *cli.Context, log logger.Logger, configPath, moduleDir, source string, gen v1.Generate, dependencyRoots v1SourceRoots) error {
	module, err := moduleconfig.ReadGitDependency(moduleDir, source)
	if err != nil {
		return fmt.Errorf("ReadGitDependency: %w", err)
	}
	otherRoots := make(v1SourceRoots, 0, len(dependencyRoots))
	for _, root := range dependencyRoots {
		if root.module != source {
			otherRoots = append(otherRoots, root)
		}
	}
	return generateV1ModuleWithRoots(ctx, log, configPath, moduleDir, gen, module, otherRoots)
}

func generateV1ModuleWithRoots(ctx *cli.Context, log logger.Logger, configPath, moduleDir string, gen v1.Generate, module v1.Module, importRoots v1SourceRoots) error {
	if err := checkV1ImportPathCollisions(moduleDir, module.Roots, importRoots.paths()); err != nil {
		return fmt.Errorf("module %s: %w", module.Name, err)
	}

	cfg := config.Config{}
	for _, root := range module.Roots {
		info, err := os.Stat(filepath.Join(moduleDir, root))
		if err != nil || !info.IsDir() {
			return fmt.Errorf("module %s: invalid root %q", module.Name, root)
		}
		cfg.Generate.Inputs = append(cfg.Generate.Inputs, config.Input{
			InputFilesDir: config.InputFilesDir{Root: root, Path: "."},
		})
	}
	for _, plugin := range gen.Plugins {
		remote := plugin.Remote
		if remote != "" {
			remote += ":" + plugin.Version
		}
		outAbs := filepath.Join(filepath.Dir(configPath), plugin.Out)
		outRel, err := filepath.Rel(moduleDir, outAbs)
		if err != nil {
			return err
		}
		cfg.Generate.Plugins = append(cfg.Generate.Plugins, config.Plugin{
			Name: plugin.Name, Remote: remote, Out: outRel, Opts: config.PluginOpts(plugin.Opts),
		})
	}
	cfg.Generate.Managed = gen.Generate.Managed
	if prefix := gen.Options.Go.PackagePrefix; prefix != nil && *prefix != "" {
		cfg.Generate.Managed.Enabled = true
		rule := config.ManagedOverrideRule{FileOption: "go_package_prefix", Value: *prefix}
		if gen.InheritedGoPackagePrefix {
			cfg.Generate.Managed.Override = append([]config.ManagedOverrideRule{rule}, cfg.Generate.Managed.Override...)
		} else {
			cfg.Generate.Managed.Override = append(cfg.Generate.Managed.Override, rule)
		}
	}
	app, err := buildCore(log, cfg)
	if err != nil {
		return fmt.Errorf("build v1 module %s: %w", module.Name, err)
	}
	app.SetImportRoots(importRoots.paths())
	if cfg.Generate.Managed.Enabled {
		moduleRoots := append(v1SourceRoots(nil), importRoots...)
		for _, root := range module.Roots {
			moduleRoots = append(moduleRoots, v1SourceRoot{path: filepath.Join(moduleDir, root), module: module.Name})
		}
		fileModules, err := moduleRoots.fileModules()
		if err != nil {
			return fmt.Errorf("fileModules: %w", err)
		}
		app.SetFileModules(fileModules)
	}
	if err := app.Generate(ctx.Context, moduleDir, ctx.String(flagGenerateDescriptorSetOut.Name), ctx.Bool(flagGenerateIncludeImports.Name)); err != nil {
		return fmt.Errorf("generate module %s: %w", module.Name, err)
	}
	return nil
}
