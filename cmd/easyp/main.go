package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/samber/lo"
	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/api"
	"github.com/easyp-tech/easyp/internal/version"
)

func initLogger(isDebug bool) {
	// use info as default level
	level := slog.LevelInfo

	if isDebug {
		level = slog.LevelDebug
	}

	slog.SetLogLoggerLevel(level)
}

func main() {
	app := &cli.App{
		Name:        "easyp info",
		HelpName:    "easyp",
		Usage:       "usage info",
		UsageText:   "usage text info",
		ArgsUsage:   "args usage info",
		Version:     version.System(),
		Description: "description info",
		Commands: buildCommand(
			api.Lint{},
			api.Mod{},
		),
		Flags: []cli.Flag{
			api.FlagDebug,
			api.FlagCfg,
		},
		Before: func(ctx *cli.Context) error {
			initLogger(ctx.Bool(api.FlagDebug.Name))
			return nil
		},
		BashComplete: cli.DefaultAppComplete,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func buildCommand(handlers ...api.Handler) []*cli.Command {
	return lo.Map(handlers, func(handler api.Handler, _ int) *cli.Command {
		return handler.Command()
	})
}
