package api

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/api/config"
	"github.com/easyp-tech/easyp/internal/api/factories"
)

var _ Handler = (*Mod)(nil)

// Mod is a handler for package manager
type Mod struct{}

func (m Mod) Command() *cli.Command {
	getCmd := &cli.Command{
		Name:        "get",
		Usage:       "download and install package",
		UsageText:   "download and install package",
		Description: "download and install package",
	}
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
		Subcommands:            []*cli.Command{getCmd, downloadCmd},
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
		if err := cmd.Get(ctx.Context, dependency); err != nil {
			return fmt.Errorf("cmd.Get: %w", err)
		}
	}

	return nil
}

func (m Mod) Get(ctx cli.Context) error {
	return fmt.Errorf("not yet implemented")
}
