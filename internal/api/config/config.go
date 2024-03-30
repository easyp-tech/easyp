package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/urfave/cli/v2"
)

// Config is the configuration of easyp.
type Config struct {
	// LintConfig is the lint configuration.
	Lint LintConfig `json:"lint" yaml:"lint" env:"EASYP_LINT"`

	// Deps is the dependencies repositories
	Deps []string `json:"deps" yaml:"deps" env:"EASYP_DEPS"`
}

func ReadConfig(ctx *cli.Context) (*Config, error) {
	cfgFile, err := os.Open(ctx.String(FlagCfg.Name))
	if err != nil {
		return nil, fmt.Errorf("os.Open: %w", err)
	}

	buf, err := io.ReadAll(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}

	jsBuf, err := yaml.YAMLToJSON(buf)
	if err != nil {
		return nil, fmt.Errorf("yaml.YAMLToJSON: %w", err)
	}

	cfg := &Config{}
	err = json.Unmarshal(jsBuf, &cfg)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return cfg, nil
}
