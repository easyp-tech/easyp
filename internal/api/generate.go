package api

import (
	"github.com/urfave/cli/v2"
)

var _ Handler = (*Generate)(nil)

// Generate is a command for generating files.
type Generate struct{}

// Command implements Handler.
func (g Generate) Command() *cli.Command {
	return &cli.Command{
		Name:        "generate",
		Aliases:     []string{"g"},
		Usage:       "generate files",
		UsageText:   "generate files",
		Description: "generate files",
		Action:      g.Action,
	}
}

// Action is a handler for generate command.
func (g Generate) Action(ctx *cli.Context) error {
	return nil
}
