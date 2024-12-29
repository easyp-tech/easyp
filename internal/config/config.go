package config

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

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
	Inputs  []Input  `json:"inputs" yaml:"inputs"`
	Plugins []Plugin `json:"plugins" yaml:"plugins"`
}

// Input source for generating code.
type Input struct {
	Directory string       `yaml:"directory"`
	GitRepo   InputGitRepo `yaml:"git_repo"`
}

// InputGitRepo is the configuration of the git repository.
type InputGitRepo struct {
	URL          string `yaml:"url"`
	SubDirectory string `yaml:"sub_directory"`
}

// InputDirectory is the configuration of the directory.
type InputDirectory struct {
	Path string `yaml:"path"`
}

// Config is the configuration of easyp.
type Config struct {
	// LintConfig is the lint configuration.
	Lint LintConfig `json:"lint" yaml:"lint"`

	// Deps is the dependencies repositories
	Deps []string `json:"deps" yaml:"deps"`

	// Generate is the generate configuration.
	Generate Generate `json:"generate" yaml:"generate"`

	// BreakingCheck `breaking` command's configuration
	BreakingCheck BreakingCheck `json:"breaking" yaml:"breaking"`
}

var errFileNotFound = errors.New("config file not found")

// New creates a new configuration from the file.
func New(_ context.Context, filepath string) (*Config, error) {
	cfgFile, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errFileNotFound
		}

		return nil, fmt.Errorf("os.Open: %w", err)
	}

	defer func() {
		err := cfgFile.Close()
		if err != nil {
			slog.Debug("cfgFile.Close", slog.String("filepath", filepath))
		}
	}()

	buf, err := io.ReadAll(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}

	cfg := &Config{}
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal: %w", err)
	}

	return cfg, nil
}
