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
	for _, scope := range ctx.Lineage() {
		// Inspect the flag set at each level: a command-local default must not
		// hide an explicitly selected format on the parent application.
		explicit := false
		for _, name := range scope.LocalFlagNames() {
			if name == Format.Name || name == "f" {
				explicit = true
				break
			}
		}
		if explicit && scope.IsSet(Format.Name) {
			return scope.String(Format.Name)
		}
	}
	return defaultFormat
}
