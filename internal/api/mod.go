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
		// FIXME: Use flags for mod
		// Flags: []cli.Flag{
		// 	flagCfg,
		// 	flagLintDirectoryPath,
		// },
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
	fmt.Println("Start mod")
	return nil
}
