package generation

import (
	"fmt"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core"
)

func prepareV1GeneratorConfig(configPath, moduleDir string, gen v1.Generate, module v1.Module) (core.Options, error) {
	cfg := core.Options{}
	for _, root := range module.Roots {
		cfg.Inputs.InputFilesDir = append(cfg.Inputs.InputFilesDir, core.InputFilesDir{Root: root, Path: "."})
	}
	for _, plugin := range gen.Plugins {
		remote := plugin.Remote
		if remote != "" {
			remote += ":" + plugin.Version
		}
		outAbs := filepath.Join(filepath.Dir(configPath), plugin.Out)
		outRel, err := filepath.Rel(moduleDir, outAbs)
		if err != nil {
			return core.Options{}, fmt.Errorf("Rel: %w", err)
		}
		cfg.Plugins = append(cfg.Plugins, core.Plugin{
			Source: core.PluginSource{Name: plugin.Name, Path: plugin.Path, Command: plugin.Command, Remote: remote},
			Out:    outRel, Options: plugin.Opts,
		})
	}
	cfg.ManagedModeConfig = convertManagedModeConfig(gen.Generate.Managed)
	if prefix := gen.Options.Go.PackagePrefix; prefix != nil && *prefix != "" {
		cfg.ManagedModeConfig.Enabled = true
		rule := core.ManagedOverrideRule{FileOption: "go_package_prefix", Value: *prefix}
		if gen.InheritedGoPackagePrefix {
			cfg.ManagedModeConfig.Override = append([]core.ManagedOverrideRule{rule}, cfg.ManagedModeConfig.Override...)
		} else {
			cfg.ManagedModeConfig.Override = append(cfg.ManagedModeConfig.Override, rule)
		}
	}
	return cfg, nil
}

func convertManagedModeConfig(cfg config.ManagedMode) core.ManagedModeConfig {
	managed := core.ManagedModeConfig{Enabled: cfg.Enabled}
	if cfg.Disable != nil {
		managed.Disable = make([]core.ManagedDisableRule, 0, len(cfg.Disable))
	}
	for _, rule := range cfg.Disable {
		managed.Disable = append(managed.Disable, core.ManagedDisableRule{
			Module: rule.Module, Package: rule.Package, Path: rule.Path,
			FileOption: core.FileOptionType(rule.FileOption), FieldOption: core.FieldOptionType(rule.FieldOption), Field: rule.Field,
		})
	}
	if cfg.Override != nil {
		managed.Override = make([]core.ManagedOverrideRule, 0, len(cfg.Override))
	}
	for _, rule := range cfg.Override {
		managed.Override = append(managed.Override, core.ManagedOverrideRule{
			FileOption: core.FileOptionType(rule.FileOption), FieldOption: core.FieldOptionType(rule.FieldOption), Value: rule.Value,
			Module: rule.Module, Package: rule.Package, Path: rule.Path, Field: rule.Field,
		})
	}
	return managed
}
