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

	"github.com/easyp-tech/easyp/internal/adapters/console"
	"github.com/easyp-tech/easyp/internal/adapters/go_git"
	lockfile "github.com/easyp-tech/easyp/internal/adapters/lock_file"
	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	"github.com/easyp-tech/easyp/internal/adapters/storage"
	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/rules"
)

var (
	ErrPathNotAbsolute = errors.New("path is not absolute")
)

const (
	envEasypPath     = "EASYPPATH"
	defaultEasypPath = ".easyp"
)

func errExit(code int, msg string, args ...any) {
	slog.Info(msg, args...)
	os.Exit(code)
}

// getEasypPath return path for cache, modules storage
func getEasypPath() (string, error) {
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

	slog.Debug("Use storage", "path", easypPath)

	return easypPath, nil
}

func buildCore(_ context.Context, cfg config.Config, dirWalker core.DirWalker) (*core.Core, error) {
	lintRules, ignoreOnly, err := rules.New(cfg.Lint)
	if err != nil {
		return nil, fmt.Errorf("cfg.BuildLinterRules: %w", err)
	}

	lockFile := lockfile.New(dirWalker)
	easypPath, err := getEasypPath()
	if err != nil {
		return nil, fmt.Errorf("getEasypPath: %w", err)
	}

	store := storage.New(easypPath, lockFile)

	moduleCfg := moduleconfig.New()

	currentProjectGitWalker := go_git.New()

	breakingCheckConfig := core.BreakingCheckConfig{
		IgnoreDirs:    cfg.BreakingCheck.Ignore,
		AgainstGitRef: cfg.BreakingCheck.AgainstGitRef,
	}

	app := core.New(
		lintRules,
		cfg.Lint.Ignore,
		cfg.Deps,
		ignoreOnly,
		slog.Default(), // TODO: remove global state
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
	)

	return app, nil
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
