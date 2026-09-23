package flags

import (
	"github.com/urfave/cli/v2"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

const (
	globalCategory = "global"
)

// Flags.
var (
	Config = &cli.StringFlag{
		Name:        "cfg",
		Category:    globalCategory,
		DefaultText: "specify the path to the configuration file",
		Usage:       "Specify the absolute or relative path to the configuration file for setting up the application.",
		Value:       v1.PolicyFile,
		Aliases:     []string{"config"},
		EnvVars:     []string{"EASYP_CFG"},
		TakesFile:   true,
	}

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
