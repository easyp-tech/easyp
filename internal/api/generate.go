package api

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/flags"
)

var _ Handler = (*Generate)(nil)

// Generate is a handler for generate command.
type Generate struct{}

var (
	flagGenerateDirectoryPath = &cli.StringFlag{
		Name:       "path",
		Usage:      "set path to directory with proto files",
		Required:   true,
		HasBeenSet: true,
		Value:      ".",
		Aliases:    []string{"p"},
		EnvVars:    []string{"EASYP_ROOT_GENERATE_PATH"},
	}
)

// Command implements Handler.
func (g Generate) Command() *cli.Command {
	return &cli.Command{
		Name:        "generate",
		Aliases:     []string{"g"},
		Usage:       "generate code from proto files",
		UsageText:   "generate code from proto files",
		Description: "generate code from proto files",
		Action:      g.Action,
		Flags: []cli.Flag{
			flagGenerateDirectoryPath,
		},
		HelpName: "help",
	}
}

// Action implements Handler.
func (g Generate) Action(ctx *cli.Context) error {
	cfg, err := config.New(ctx.Context, ctx.String(flags.Config.Name))
	if err != nil {
		return fmt.Errorf("config.New: %w", err)
	}
	app, err := buildCore(ctx.Context, *cfg)
	if err != nil {
		return fmt.Errorf("buildCore: %w", err)
	}

	dir := ctx.String(flagGenerateDirectoryPath.Name)
	err = app.Generate(ctx.Context, ".", dir)
	if err != nil {
		return fmt.Errorf("generator.Generate: %w", err)
	}

	return nil
}
