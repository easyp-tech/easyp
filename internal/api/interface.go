package api

import (
	"github.com/urfave/cli/v2"
)

// Handler is an interface for a handling command.
type Handler interface {
	// Command returns a command.
	Command() *cli.Command
}
