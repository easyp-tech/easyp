package api

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

// Handler is an interface for a handling command.
type Handler interface {
	// Command returns a command.
	Command() *cli.Command
}

// Config is the configuration of easyp.
type Config struct {
	// LintConfig is the lint configuration.
	Lint LintConfig `json:"lint" yaml:"lint" env:"EASYP_LINT"`

	// Deps is the dependencies repositories
	Deps []string `json:"deps" yaml:"deps" env:"EASYP_DEPS"`
}

// LintConfig contains linter configuration.
type LintConfig struct {
	Use []string `json:"use" yaml:"use" env:"USE"`
}

func readConfig(ctx *cli.Context) (*Config, error) {
	cfgFile, err := os.Open(ctx.String(FlagCfg.Name))
	if err != nil {
		return nil, fmt.Errorf("os.Open: %w", err)
	}

	cfg := &Config{}
	err = yaml.NewDecoder(cfgFile).Decode(&cfg)
	if err != nil {
		return nil, fmt.Errorf("yaml.NewDecoder.Decode: %w", err)
	}

	return cfg, nil
}
