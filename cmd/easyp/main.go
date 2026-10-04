package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/samber/lo"
	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/api"
	"github.com/easyp-tech/easyp/internal/flags"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/version"
)

func initLogger(isDebug bool) logger.Logger {
	// use info as default level
	level := slog.LevelInfo

	if isDebug {
		level = slog.LevelDebug
	}

	return logger.New(slog.New(
		slog.NewTextHandler(
			os.Stderr,
			&slog.HandlerOptions{
				AddSource: false,
				Level:     level,
			},
		),
	))
}

func main() {
	app := &cli.App{
		Name:        "easyp",
		HelpName:    "easyp",
		Usage:       "modern Protocol Buffers toolkit",
		UsageText:   "easyp [global options] command [command options]",
		Version:     version.System(),
		Description: "Manage protobuf dependencies, lint APIs, generate code, and check compatibility.",
		Commands: buildCommand(
			api.Lint{},
			api.Mod{},
			api.Get{},
			api.Completion{},
			api.Init{},
			api.Migrate{},
			api.Generate{},
			api.LsFiles{},
			api.SchemaGen{},
			api.Validate{},
			api.BreakingCheck{},
		),
		Flags: []cli.Flag{
			flags.Frozen(),
			flags.Config,
			flags.DebugMode,
			flags.Format,
		},
		Before: func(ctx *cli.Context) error {
			log := initLogger(ctx.Bool(flags.DebugMode.Name))
			ctx.App.Metadata = map[string]interface{}{
				"logger": log,
			}
			return nil
		},
		EnableBashCompletion: true,
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
