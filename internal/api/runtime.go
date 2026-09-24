package api

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

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

func buildCore(log logger.Logger, cfg config.Config, importRoots []string) (*core.Core, error) {
	core.SetAllowCommentIgnores(cfg.Lint.AllowCommentIgnores)

	lintRules, ignoreOnly, err := rules.New(cfg.Lint)
	if err != nil {
		return nil, fmt.Errorf("New: %w", err)
	}

	return core.New(core.Options{
		Rules:                   lintRules,
		Ignore:                  append(append([]string(nil), cfg.Lint.Ignore...), defaultVendorDir),
		IgnoreOnly:              ignoreOnly,
		Logger:                  log,
		ImportRoots:             importRoots,
		CurrentProjectGitWalker: go_git.New(),
		BreakingCheckConfig: core.BreakingCheckConfig{
			IgnoreDirs:    append(append([]string(nil), cfg.BreakingCheck.Ignore...), defaultVendorDir),
			AgainstGitRef: cfg.BreakingCheck.AgainstGitRef,
		},
	}), nil
}
