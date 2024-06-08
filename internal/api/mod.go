package api

import (
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/api/config"
	"github.com/easyp-tech/easyp/internal/api/factories"
	"github.com/easyp-tech/easyp/internal/mod/models"
)

var _ Handler = (*Mod)(nil)

// Mod is a handler for package manager
type Mod struct{}

func (m Mod) Command() *cli.Command {
	downloadCmd := &cli.Command{
		Name:        "download",
		Usage:       "download modules to local cache",
		UsageText:   "download modules to local cache",
		Description: "download modules to local cache",
		Action:      m.Download,
	}
	updateCmd := &cli.Command{
		Name:        "update",
		Usage:       "update modules version using version from config",
		UsageText:   "update modules version using version from config",
		Description: "update modules version using version from config",
		Action:      m.Update,
	}

	return &cli.Command{
		Name:                   "mod",
		Aliases:                []string{"m"},
		Usage:                  "package manager",
		UsageText:              "package manager",
		Description:            "package manager",
		ArgsUsage:              "",
		Category:               "",
		BashComplete:           nil,
		Before:                 nil,
		After:                  nil,
		Action:                 nil,
		OnUsageError:           nil,
		Subcommands:            []*cli.Command{downloadCmd, updateCmd},
		Flags:                  []cli.Flag{},
		SkipFlagParsing:        false,
		HideHelp:               false,
		HideHelpCommand:        false,
		Hidden:                 false,
		UseShortOptionHandling: false,
		HelpName:               "help",
		CustomHelpTemplate:     "",
	}
}

func (m Mod) Download(ctx *cli.Context) error {
	cfg, err := config.ReadConfig(ctx)
	if err != nil {
		return fmt.Errorf("ReadConfig: %w", err)
	}

	cmd, err := factories.NewMod()
	if err != nil {
		return fmt.Errorf("factories.NewMod: %w", err)
	}

	if err := cmd.Download(ctx.Context, cfg.Deps); err != nil {
		if errors.Is(err, models.ErrVersionNotFound) {
			os.Exit(1)
		}

		return fmt.Errorf("cmd.Download: %w", err)
	}
	return nil
}

func (m Mod) Update(ctx *cli.Context) error {
	cfg, err := config.ReadConfig(ctx)
	if err != nil {
		return fmt.Errorf("ReadConfig: %w", err)
	}

	cmd, err := factories.NewMod()
	if err != nil {
		return fmt.Errorf("factories.NewMod: %w", err)
	}

	if err := cmd.Update(ctx.Context, cfg.Deps); err != nil {
		if errors.Is(err, models.ErrVersionNotFound) {
			os.Exit(1)
		}

		return fmt.Errorf("cmd.Download: %w", err)
	}
	return nil
}
