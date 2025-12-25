package flags

import "github.com/urfave/cli/v2"

const (
	globalCategory = "global"
)

const (
	// Max file size is 1mb.
	defaultConfigFilePath = "easyp.yaml"
)

// Flags.
var (
	Config = &cli.StringFlag{
		Name:        "cfg",
		Category:    globalCategory,
		DefaultText: "specify the path to the configuration file",
		FilePath:    "",
		Usage:       "Specify the absolute or relative path to the configuration file for setting up the application.",
		Required:    true,
		Hidden:      false,
		HasBeenSet:  true,
		Value:       defaultConfigFilePath,
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
