package config

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigFileName = "easyp.yaml"
)

// Plugin is the configuration of the plugin.
type Plugin struct {
	Name string            `json:"name" yaml:"name"`
	Out  string            `json:"out" yaml:"out"`
	Opts map[string]string `json:"opts" yaml:"opts"`
}

// DependencyEntryPoint part for generate code from dep
type DependencyEntryPoint struct {
	Dep  string `json:"dep" yaml:"dep"`
	Path string `json:"path" yaml:"path"`
}

// Generate is the configuration of the generate command.
type Generate struct {
	DependencyEntryPoint *DependencyEntryPoint `json:"dependency_entry_point" yaml:"dependency_entry_point"`
	Inputs               []Input               `json:"inputs" yaml:"inputs"`
	Plugins              []Plugin              `json:"plugins" yaml:"plugins"`
	ProtoRoot            string                `json:"proto_root" yaml:"proto_root"`
	GenerateOutDirs      bool                  `json:"generate_out_dirs" yaml:"generate_out_dirs"`
}

// Input source for generating code.
type Input struct {
	Directory string       `yaml:"directory"`
	GitRepo   InputGitRepo `yaml:"git_repo"`
}

// InputGitRepo is the configuration of the git repository.
type InputGitRepo struct {
	URL          string `yaml:"url"`
	Branch       string `yaml:"branch"`
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

	//fmt.Println(string(buf))

	cfg := &Config{}
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal: %w", err)
	}

	return cfg, nil
}
