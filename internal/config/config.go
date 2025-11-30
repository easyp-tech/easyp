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
	// Sources

	Name    string   `json:"name,omitempty" yaml:"name,omitempty"`
	Remote  string   `json:"remote,omitempty" yaml:"remote,omitempty"`
	Path    string   `json:"path,omitempty" yaml:"path,omitempty"`
	Command []string `json:"command,omitempty" yaml:"command,omitempty"`

	Out         string            `json:"out" yaml:"out"`
	Opts        map[string]string `json:"opts,omitempty" yaml:"opts,omitempty"`
	WithImports bool              `json:"with_imports,omitempty" yaml:"with_imports,omitempty"`
}

// Generate is the configuration of the generate command.
type Generate struct {
	Inputs  []Input  `json:"inputs" yaml:"inputs"`
	Plugins []Plugin `json:"plugins" yaml:"plugins"`
}

// Input source for generating code.
type Input struct {
	InputFilesDir InputFilesDir `yaml:"directory"`
	GitRepo       InputGitRepo  `yaml:"git_repo"`
}

// InputGitRepo is the configuration of the git repository.
type InputGitRepo struct {
	URL          string `yaml:"url"`
	SubDirectory string `yaml:"sub_directory"`
	Out          string `yaml:"out"`
	Root         string `yaml:"root"`
}

// InputDirectory is the configuration of the directory.
type InputDirectory struct {
	Path string `yaml:"path"`
}

// InputFilesDir is the configuration of the directory with additional functionality.
type InputFilesDir struct {
	Path string `yaml:"path"`
	Root string `yaml:"root"`
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

	err = cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return cfg, nil
}

func (d *InputFilesDir) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		// строка — старый формат
		var path string
		if err := value.Decode(&path); err != nil {
			return err
		}
		d.Path = path
		d.Root = "."
	case yaml.MappingNode:
		// структура — новый формат
		type raw InputFilesDir
		var r raw
		if err := value.Decode(&r); err != nil {
			return err
		}
		*d = InputFilesDir(r)
		if d.Root == "" {
			d.Root = "."
		}
	default:
		return fmt.Errorf("unsupported type for directory: %v", value.Kind)
	}
	return nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("config is nil")
	}

	// Validate plugins
	for _, plugin := range c.Generate.Plugins {
		// Only one source allowed.
		var sourceCount int
		if plugin.Name != "" {
			sourceCount++
		}
		if plugin.Remote != "" {
			sourceCount++
		}
		if plugin.Path != "" {
			sourceCount++
		}
		if len(plugin.Command) > 0 {
			sourceCount++
		}

		if sourceCount > 1 {
			return fmt.Errorf("plugin has multiple sources (name, remote, path, or command)")
		}

		if sourceCount == 0 {
			return fmt.Errorf("plugin must have one source: name, remote, path, or command")
		}
	}

	return nil
}
