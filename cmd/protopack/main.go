package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/samber/lo"
	"github.com/urfave/cli/v2"

	"go.redsock.ru/protopack/internal/api"
	"go.redsock.ru/protopack/internal/flags"
	"go.redsock.ru/protopack/internal/version"
)

func initLogger(isDebug bool) *slog.Logger {
	// use info as default level
	level := slog.LevelInfo

	if isDebug {
		level = slog.LevelDebug
	}

	logger := slog.New(
		slog.NewTextHandler(
			os.Stderr,
			&slog.HandlerOptions{
				AddSource: false,
				Level:     level,
			},
		),
	)

	slog.SetDefault(logger) // TODO: remove global state

	return logger
}

func main() {
	app := &cli.App{
		Name:        "easyp",
		HelpName:    "easyp",
		Usage:       "usage info",
		UsageText:   "usage text info",
		ArgsUsage:   "args usage info",
		Version:     version.System(),
		Description: "description info",
		Commands: buildCommand(
			api.Lint{},
			api.Mod{},
			api.Completion{},
			api.Init{},
			api.Generate{},
			api.BreakingCheck{},
		),
		Flags: []cli.Flag{
			flags.Config,
			flags.DebugMode,
		},
		Before: func(ctx *cli.Context) error {
			initLogger(ctx.Bool(flags.DebugMode.Name))
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
