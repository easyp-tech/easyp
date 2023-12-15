package main

import (
	"log"
	"os"

	"github.com/samber/lo"
	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/api"
	"github.com/easyp-tech/easyp/internal/version"
)

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
		),
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
