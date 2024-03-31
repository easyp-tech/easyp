package api

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/api/config"
	"github.com/easyp-tech/easyp/internal/mod"
	moduleconfig "github.com/easyp-tech/easyp/internal/mod/adapters/module_config"
	"github.com/easyp-tech/easyp/internal/mod/adapters/storage"
)

var _ Handler = (*Mod)(nil)

// Mod is a handler for package manager
type Mod struct{}

func (m Mod) Command() *cli.Command {
	return &cli.Command{
		Name:         "mod",
		Aliases:      []string{"m"},
		Usage:        "package manager",
		UsageText:    "package manager",
		Description:  "package manager",
		ArgsUsage:    "",
		Category:     "",
		BashComplete: nil,
		Before:       nil,
		After:        nil,
		Action:       m.Action,
		OnUsageError: nil,
		Subcommands:  nil,
		Flags: []cli.Flag{},
		SkipFlagParsing:        false,
		HideHelp:               false,
		HideHelpCommand:        false,
		Hidden:                 false,
		UseShortOptionHandling: false,
		HelpName:               "help",
		CustomHelpTemplate:     "",
	}
}

const (
	envEasypPath     = "EASYPPATH"
	defaultEasypPath = ".easyp"
)

func (m Mod) Action(ctx *cli.Context) error {
	cfg, err := config.ReadConfig(ctx)
	if err != nil {
		return fmt.Errorf("ReadConfig: %w", err)
	}

	easypPath := os.Getenv(envEasypPath)
	if easypPath == "" {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("os.UserHomeDir: %w", err)
		}
		easypPath = filepath.Join(userHomeDir, defaultEasypPath)
	}

	slog.Info("Use storage", "path", easypPath)

	store := storage.New(easypPath)
	moduleConfig := moduleconfig.New()
	cmd := mod.New(store, moduleConfig)

	for _, dependency := range cfg.Deps {
		if err := cmd.Get(ctx.Context, dependency); err != nil {
			return fmt.Errorf("cmd.Get: %w", err)
		}
	}

	return nil
}
