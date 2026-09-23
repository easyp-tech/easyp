package api

import "github.com/urfave/cli/v2"

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
	tidyCmd := &cli.Command{
		Name:   "tidy",
		Usage:  "resolve protobuf.mod and write protobuf.lock",
		Action: m.Tidy,
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
		Subcommands:            []*cli.Command{downloadCmd, updateCmd, tidyCmd},
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
