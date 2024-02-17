package api

import (
	"fmt"

	"github.com/urfave/cli/v2"
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
		Flags: []cli.Flag{
			flagCfg,
			// FIXME: Use flags for package_manager
			// flagLintDirectoryPath,
		},
		SkipFlagParsing:        false,
		HideHelp:               false,
		HideHelpCommand:        false,
		Hidden:                 false,
		UseShortOptionHandling: false,
		HelpName:               "help",
		CustomHelpTemplate:     "",
	}
}

func (m Mod) Action(ctx *cli.Context) error {
	cfg, err := readConfig(ctx)
	if err != nil {
		return fmt.Errorf("readConfig: %w", err)
	}

	// TODO: TEMO DEBUG
	modTst()
	// TODO: TEMO DEBUG

	_ = cfg
	return nil
}
