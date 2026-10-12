package flags

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// ConfigName is the name of the command-local configuration file selector.
const ConfigName = "config"

// Config returns a fresh command-local --config flag. Each command decides
// which configuration format the selected file must contain.
func Config(usage string) *cli.StringFlag {
	return &cli.StringFlag{Name: ConfigName, Usage: usage, TakesFile: true}
}

// ExplicitConfig returns the --config value when it was set and rejects an
// explicitly empty selection.
func ExplicitConfig(ctx *cli.Context) (string, bool, error) {
	if !ctx.IsSet(ConfigName) {
		return "", false, nil
	}
	path := ctx.String(ConfigName)
	if path == "" {
		return "", false, fmt.Errorf("--%s must not be empty", ConfigName)
	}
	return path, true, nil
}

// Flags.
var (
	DebugMode = &cli.BoolFlag{
		Name:     "debug",
		Usage:    "Enable debug mode to get more detailed information in logs.",
		Required: false,
		Value:    false,
		Aliases:  []string{"d"},
		EnvVars:  []string{"EASYP_DEBUG"},
	}

	Format = &cli.GenericFlag{
		Name:       "format",
		Usage:      "set output format for commands that support multiple formats",
		Required:   false,
		HasBeenSet: false,
		Value: &EnumValue{
			Enum:    []string{TextFormat, JSONFormat},
			Default: "text",
		},
		Aliases: []string{"f"},
		EnvVars: []string{"EASYP_FORMAT"},
	}
)
