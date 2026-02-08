package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/adapters/console"
	"github.com/easyp-tech/easyp/internal/adapters/go_git"
	lockfile "github.com/easyp-tech/easyp/internal/adapters/lock_file"
	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	"github.com/easyp-tech/easyp/internal/adapters/storage"
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

var (
	ErrPathNotAbsolute = errors.New("path is not absolute")
)

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
			return "", fmt.Errorf("os.UserHomeDir: %w", err)
		}
		easypPath = filepath.Join(userHomeDir, defaultEasypPath)
	}

	easypPath, err := filepath.Abs(easypPath)
	if err != nil {
		return "", ErrPathNotAbsolute
	}

	log.Debug(context.Background(), "Use storage", slog.String("path", easypPath))

	return easypPath, nil
}

func buildCore(_ context.Context, log logger.Logger, cfg config.Config, dirWalker core.DirWalker) (*core.Core, error) {
	vendorPath := defaultVendorDir // TODO: read from config

	lintRules, ignoreOnly, err := rules.New(cfg.Lint)
	if err != nil {
		return nil, fmt.Errorf("cfg.BuildLinterRules: %w", err)
	}

	lockFile := lockfile.New(dirWalker)
	easypPath, err := getEasypPath(log)
	if err != nil {
		return nil, fmt.Errorf("getEasypPath: %w", err)
	}

	store := storage.New(easypPath, lockFile, log)

	moduleCfg := moduleconfig.New(log)

	currentProjectGitWalker := go_git.New()

	// always ignore vendor dir for linter and breakinf check
	linterIgnoreDirs := append(cfg.Lint.Ignore, vendorPath)
	breakingCheckIgnoreDirs := append(cfg.BreakingCheck.Ignore, vendorPath)

	breakingCheckConfig := core.BreakingCheckConfig{
		IgnoreDirs:    breakingCheckIgnoreDirs,
		AgainstGitRef: cfg.BreakingCheck.AgainstGitRef,
	}

	// Convert managed mode configuration
	managedMode := convertManagedModeConfig(cfg.Generate.Managed)

	app := core.New(
		lintRules,
		linterIgnoreDirs,
		cfg.Deps,
		ignoreOnly,
		log,
		lo.Map(cfg.Generate.Plugins, func(p config.Plugin, _ int) core.Plugin {
			return core.Plugin{
				Source: core.PluginSource{
					Name:    p.Name,
					Remote:  p.Remote,
					Path:    p.Path,
					Command: p.Command,
				},
				Out:         p.Out,
				Options:     p.Opts,
				WithImports: p.WithImports,
			}
		}),
		core.Inputs{
			InputGitRepos: lo.Filter(lo.Map(cfg.Generate.Inputs, func(i config.Input, _ int) core.InputGitRepo {
				return core.InputGitRepo{
					URL:          i.GitRepo.URL,
					SubDirectory: i.GitRepo.SubDirectory,
					Out:          i.GitRepo.Out,
					Root:         i.GitRepo.Root,
				}
			}), func(i core.InputGitRepo, _ int) bool {
				return i.URL != ""
			}),
			InputFilesDir: lo.Filter(lo.Map(cfg.Generate.Inputs, func(i config.Input, _ int) core.InputFilesDir {
				return core.InputFilesDir{
					Path: i.InputFilesDir.Path,
					Root: i.InputFilesDir.Root,
				}
			}), func(i core.InputFilesDir, _ int) bool {
				return i.Path != "" && IsExistingDir(i.Root)
			}),
		},
		console.New(),
		store,
		moduleCfg,
		lockFile,
		currentProjectGitWalker,
		breakingCheckConfig,
		managedMode,
		vendorPath, // vendorDir
	)

	return app, nil
}

// convertManagedModeConfig converts config.ManagedMode to core.ManagedModeConfig.
func convertManagedModeConfig(cfg config.ManagedMode) core.ManagedModeConfig {
	return core.ManagedModeConfig{
		Enabled: cfg.Enabled,
		Disable: lo.Map(cfg.Disable, func(r config.ManagedDisableRule, _ int) core.ManagedDisableRule {
			return core.ManagedDisableRule{
				Module:      r.Module,
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
				Path:        r.Path,
				Field:       r.Field,
			}
		}),
	}
}

func IsExistingDir(path string) bool {
	if path == "" || strings.ContainsRune(path, '\x00') {
		return false
	}

	// Очистим путь от лишнего (типа "./../")
	cleanPath := filepath.Clean(path)

	info, err := os.Stat(cleanPath)
	if err != nil {
		return false
	}

	if !info.IsDir() {
		return false
	}

	return true
}
