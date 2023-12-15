package api

import (
	"github.com/urfave/cli/v2"
)

var (
	flagCfg = &cli.StringFlag{
		Name:       "cfg",
		Usage:      "set config file path",
		Required:   true,
		HasBeenSet: true,
		Value:      "easyp.yaml",
		Aliases:    []string{"c"},
		EnvVars:    []string{"EASYP_CFG"},
	}
)
