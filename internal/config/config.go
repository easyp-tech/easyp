package config

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/a8m/envsubst"
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

// ManagedDisableRule defines a rule to disable managed mode for specific conditions.
type ManagedDisableRule struct {
	// Module disables managed mode for all files in the specified module.
	Module string `json:"module,omitempty" yaml:"module,omitempty"`
	// Path disables managed mode for files matching the specified path (directory or file).
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	// FileOption disables a specific file option from being modified.
	FileOption string `json:"file_option,omitempty" yaml:"file_option,omitempty"`
	// FieldOption disables a specific field option from being modified.
	FieldOption string `json:"field_option,omitempty" yaml:"field_option,omitempty"`
	// Field disables a specific field (fully qualified name: package.Message.field).
	Field string `json:"field,omitempty" yaml:"field,omitempty"`
}

// ManagedOverrideRule defines a rule to override file or field options.
type ManagedOverrideRule struct {
	// FileOption specifies which file option to override.
	FileOption string `json:"file_option,omitempty" yaml:"file_option,omitempty"`
	// FieldOption specifies which field option to override.
	FieldOption string `json:"field_option,omitempty" yaml:"field_option,omitempty"`
	// Value is the value to set for the option.
	Value any `json:"value,omitempty" yaml:"value,omitempty"`
	// Module applies this override only to files in the specified module.
	Module string `json:"module,omitempty" yaml:"module,omitempty"`
	// Path applies this override only to files matching the specified path.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	// Field applies this override only to the specified field (fully qualified name).
	Field string `json:"field,omitempty" yaml:"field,omitempty"`
}

// ManagedMode is the configuration for managed mode which automatically
// sets file and field options without modifying the original proto files.
type ManagedMode struct {
	// Enabled activates managed mode.
	Enabled bool `json:"enabled" yaml:"enabled"`
	// Disable contains rules to disable managed mode for specific conditions.
	Disable []ManagedDisableRule `json:"disable,omitempty" yaml:"disable,omitempty"`
	// Override contains rules to override file and field options.
	Override []ManagedOverrideRule `json:"override,omitempty" yaml:"override,omitempty"`
}

// Generate is the configuration of the generate command.
type Generate struct {
	Inputs  []Input     `json:"inputs" yaml:"inputs"`
	Plugins []Plugin    `json:"plugins" yaml:"plugins"`
	Managed ManagedMode `json:"managed,omitempty" yaml:"managed,omitempty"`
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

	return ParseConfig(buf)
}

// ParseConfig parses configuration from bytes with environment variable expansion.
// Supports escaping via $$ (e.g., $$var becomes $var, $${VAR} becomes ${VAR})
// This is the unified function for parsing easyp.yaml used throughout the codebase.
func ParseConfig(buf []byte) (*Config, error) {
	// Expand environment variables in the config file
	expanded, err := envsubst.String(string(buf))
	if err != nil {
		return nil, fmt.Errorf("envsubst.String: %w", err)
	}
	buf = []byte(expanded)

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

	// Validate managed mode
	if err := c.Generate.Managed.Validate(); err != nil {
		return fmt.Errorf("managed mode validation: %w", err)
	}

	return nil
}

// Validate validates the managed mode configuration.
func (m *ManagedMode) Validate() error {
	// Validate disable rules
	for i, rule := range m.Disable {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("disable rule %d: %w", i, err)
		}
	}

	// Validate override rules
	for i, rule := range m.Override {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("override rule %d: %w", i, err)
		}
	}

	return nil
}

// Validate validates a disable rule.
func (r *ManagedDisableRule) Validate() error {
	// At least one field must be set
	if r.Module == "" && r.Path == "" && r.FileOption == "" && r.FieldOption == "" && r.Field == "" {
		return errors.New("disable rule must have at least one field set")
	}

	// Cannot have both file_option and field_option
	if r.FileOption != "" && r.FieldOption != "" {
		return errors.New("disable rule cannot have both file_option and field_option")
	}

	// Field can only be used with field_option
	if r.Field != "" && r.FieldOption == "" {
		return errors.New("field can only be used with field_option in disable rule")
	}

	return nil
}

// Validate validates an override rule.
func (r *ManagedOverrideRule) Validate() error {
	// Must have either file_option or field_option
	if r.FileOption == "" && r.FieldOption == "" {
		return errors.New("override rule must have either file_option or field_option")
	}

	// Cannot have both file_option and field_option
	if r.FileOption != "" && r.FieldOption != "" {
		return errors.New("override rule cannot have both file_option and field_option")
	}

	// Must have a value
	if r.Value == nil {
		return errors.New("override rule must have a value")
	}

	// Field can only be used with field_option
	if r.Field != "" && r.FieldOption == "" {
		return errors.New("field can only be used with field_option in override rule")
	}

	return nil
}
