package api

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/samber/lo"
	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/adapters/go_git"
	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/rules"
)

// getLogger extracts the logger.Logger from CLI context metadata.
// Falls back to a no-op logger if not found.
func getLogger(ctx *cli.Context) logger.Logger {
	if l, ok := ctx.App.Metadata["logger"].(logger.Logger); ok {
		return l
	}
	return logger.NewNop()
}

const (
	envEasypPath     = "EASYPPATH"
	defaultEasypPath = ".easyp"
	defaultVendorDir = "easyp_vendor"
)

func errExit(log logger.Logger, code int, msg string, attrs ...slog.Attr) {
	log.Error(context.Background(), msg, attrs...)
	os.Exit(code)
}

// getEasypPath return path for cache, modules storage
func getEasypPath(log logger.Logger) (string, error) {
	easypPath := os.Getenv(envEasypPath)
	if easypPath == "" {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("UserHomeDir: %w", err)
		}
		easypPath = filepath.Join(userHomeDir, defaultEasypPath)
	}

	easypPath, err := filepath.Abs(easypPath)
	if err != nil {
		return "", fmt.Errorf("Abs: %w", err)
	}

	log.Debug(context.Background(), "Use storage", slog.String("path", easypPath))

	return easypPath, nil
}

func buildCore(log logger.Logger, cfg config.Config) (*core.Core, error) {
	core.SetAllowCommentIgnores(cfg.Lint.AllowCommentIgnores)

	lintRules, ignoreOnly, err := rules.New(cfg.Lint)
	if err != nil {
		return nil, fmt.Errorf("New: %w", err)
	}

	plugins := make([]core.Plugin, 0, len(cfg.Generate.Plugins))
	for _, item := range cfg.Generate.Plugins {
		plugins = append(plugins, core.Plugin{
			Source: core.PluginSource{
				Name: item.Name, Remote: item.Remote, Path: item.Path, Command: item.Command,
			},
			Out: item.Out, Options: item.Opts, WithImports: item.WithImports,
		})
	}

	inputs := core.Inputs{}
	for _, item := range cfg.Generate.Inputs {
		if item.InputFilesDir.Path != "" {
			inputs.InputFilesDir = append(inputs.InputFilesDir, core.InputFilesDir{
				Path: item.InputFilesDir.Path,
				Root: item.InputFilesDir.Root,
			})
		}
	}

	return core.New(core.Options{
		Rules:                   lintRules,
		Ignore:                  append(append([]string(nil), cfg.Lint.Ignore...), defaultVendorDir),
		IgnoreOnly:              ignoreOnly,
		Logger:                  log,
		Plugins:                 plugins,
		Inputs:                  inputs,
		CurrentProjectGitWalker: go_git.New(),
		BreakingCheckConfig: core.BreakingCheckConfig{
			IgnoreDirs:    append(append([]string(nil), cfg.BreakingCheck.Ignore...), defaultVendorDir),
			AgainstGitRef: cfg.BreakingCheck.AgainstGitRef,
		},
		ManagedModeConfig: convertManagedModeConfig(cfg.Generate.Managed),
	}), nil
}

// convertManagedModeConfig converts config.ManagedMode to core.ManagedModeConfig.
func convertManagedModeConfig(cfg config.ManagedMode) core.ManagedModeConfig {
	return core.ManagedModeConfig{
		Enabled: cfg.Enabled,
		Disable: lo.Map(cfg.Disable, func(r config.ManagedDisableRule, _ int) core.ManagedDisableRule {
			return core.ManagedDisableRule{
				Module:      r.Module,
				Package:     r.Package,
				Path:        r.Path,
				FileOption:  core.FileOptionType(r.FileOption),
				FieldOption: core.FieldOptionType(r.FieldOption),
				Field:       r.Field,
			}
		}),
		Override: lo.Map(cfg.Override, func(r config.ManagedOverrideRule, _ int) core.ManagedOverrideRule {
			return core.ManagedOverrideRule{
				FileOption:  core.FileOptionType(r.FileOption),
				FieldOption: core.FieldOptionType(r.FieldOption),
				Value:       r.Value,
				Module:      r.Module,
				Package:     r.Package,
				Path:        r.Path,
				Field:       r.Field,
			}
		}),
	}
}
