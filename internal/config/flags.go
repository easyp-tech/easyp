package config

import (
	"github.com/urfave/cli/v2"
)

var (
	FlagCfg = &cli.StringFlag{
		Name:       "cfg",
		Usage:      "set config file path",
		Required:   true,
		HasBeenSet: true,
		Value:      DefaultConfigFileName,
		Aliases:    []string{"config"},
		EnvVars:    []string{"EASYP_CFG"},
	}

	FlagDebug = &cli.BoolFlag{
		Name:       "debug",
		Usage:      "set config file path",
		Required:   false,
		HasBeenSet: false,
		Value:      false,
		Aliases:    []string{"d"},
		EnvVars:    []string{"DEBUG"},
	}
)
