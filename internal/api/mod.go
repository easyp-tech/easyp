package api

import (
	"errors"
	"fmt"
	"log/slog"

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
		Subcommands:            []*cli.Command{downloadCmd},
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

	for _, dependency := range cfg.Deps {
		err := cmd.Get(ctx.Context, dependency)
		if err != nil {
			if errors.Is(err, models.ErrVersionNotFound) {
				slog.Warn("Version not found", "dependency", dependency)
				continue
			}

			return fmt.Errorf("cmd.Get: %w", err)
		}
	}

	return nil
}
