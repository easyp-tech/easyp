package api

import (
	"github.com/urfave/cli/v2"
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
}

// LintConfig contains linter configuration.
type LintConfig struct {
	Use []string `json:"use" yaml:"use" env:"USE"`
}
