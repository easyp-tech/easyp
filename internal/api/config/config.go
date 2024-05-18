package config

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

// Plugin is the configuration of the plugin.
type Plugin struct {
	Name string            `json:"name" yaml:"name"`
	Out  string            `json:"out" yaml:"out"`
	Opts map[string]string `json:"opts" yaml:"opts"`
}

// Generate is the configuration of the generate command.
type Generate struct {
	Plugins []Plugin `json:"plugins" yaml:"plugins"`
}

// Config is the configuration of easyp.
type Config struct {
	// LintConfig is the lint configuration.
	Lint LintConfig `json:"lint" yaml:"lint"`

	// Deps is the dependencies repositories
	Deps []string `json:"deps" yaml:"deps"`

	// Generate is the generate configuration.
	Generate Generate `json:"generate" yaml:"generate"`
}

// ReadConfig reads the configuration from the file.
func ReadConfig(ctx *cli.Context) (*Config, error) {
	cfgFileName := ctx.String(FlagCfg.Name)
	cfgFile, err := os.Open(cfgFileName)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("Error open config file: %s", cfgFileName)
		}

		return nil, fmt.Errorf("os.Open: %w", err)
	}
	defer func() {
		err := cfgFile.Close()
		if err != nil {
			log.Fatalf("Error close config file: %s", cfgFileName)
		}
	}()

	buf, err := io.ReadAll(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}

	cfg := &Config{}
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return cfg, nil
}
