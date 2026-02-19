package flags

import (
	"github.com/urfave/cli/v2"
)

// Format values shared across commands.
const (
	TextFormat = "text"
	JSONFormat = "json"
)

// GetFormat returns the format to use for the command, preferring the global
// --format flag when it is explicitly set, otherwise falling back to the
// command-specific default.
func GetFormat(ctx *cli.Context, defaultFormat string) string {
	if ctx.IsSet(Format.Name) {
		return ctx.String(Format.Name)
	}
	return defaultFormat
}
