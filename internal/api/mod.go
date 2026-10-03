package api

import (
	"github.com/easyp-tech/easyp/internal/flags"
	"github.com/urfave/cli/v2"
)

var _ Handler = (*Mod)(nil)

// Mod is a handler for package manager
type Mod struct{}

func (m Mod) Command() *cli.Command {
	downloadCmd := &cli.Command{
		Flags:       []cli.Flag{flags.Frozen()},
		Name:        "download",
		Usage:       "download modules to local cache",
		UsageText:   "download modules to local cache",
		Description: "download modules to local cache",
		Action:      m.Download,
	}
	updateCmd := &cli.Command{
		Flags:       []cli.Flag{flags.Frozen()},
		Name:        "update",
		Usage:       "refresh requirements within their current major versions and rewrite protobuf.mod/protobuf.lock",
		UsageText:   "refresh requirements within their current major versions and rewrite protobuf.mod/protobuf.lock",
		Description: "refresh requirements within their current major versions and rewrite protobuf.mod/protobuf.lock",
		Action:      m.Update,
	}
	tidyCmd := &cli.Command{
		Flags:  []cli.Flag{flags.Frozen()},
		Name:   "tidy",
		Usage:  "resolve protobuf.mod and write protobuf.lock",
		Action: m.Tidy,
	}
	vendorCmd := &cli.Command{
		Flags:  []cli.Flag{flags.Frozen()},
		Name:   "vendor",
		Usage:  "copy locked protobuf imports into easyp_vendor",
		Action: m.Vendor,
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
		Subcommands:            []*cli.Command{downloadCmd, updateCmd, tidyCmd, vendorCmd},
		Flags:                  []cli.Flag{flags.Frozen()},
		SkipFlagParsing:        false,
		HideHelp:               false,
		HideHelpCommand:        false,
		Hidden:                 false,
		UseShortOptionHandling: false,
		HelpName:               "help",
		CustomHelpTemplate:     "",
	}
}
