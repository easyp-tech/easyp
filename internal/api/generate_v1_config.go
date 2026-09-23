package api

import (
	"fmt"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func prepareV1GeneratorConfig(configPath, moduleDir string, gen v1.Generate, module v1.Module) (config.Config, error) {
	cfg := config.Config{}
	for _, root := range module.Roots {
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
			return config.Config{}, fmt.Errorf("Rel: %w", err)
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
	return cfg, nil
}
