package v1

import (
	"bytes"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/config"
)

// Generate is the consumer-side v1 file. Producer policy and legacy inputs
// remain outside this file; managed descriptor options retain their v0 shape.
type Generate struct {
	Version                  string          `yaml:"version"`
	InheritedGoPackagePrefix bool            `yaml:"-"`
	Generate                 GenerateTargets `yaml:"generate"`
	Plugins                  []Plugin        `yaml:"plugins"`
	Options                  GenerateOptions `yaml:"options"`
}

// GenerateTargets selects modules and their managed descriptor options.
type GenerateTargets struct {
	Modules  []string           `yaml:"modules"`
	Packages []string           `yaml:"packages"`
	Managed  config.ManagedMode `yaml:"managed"`
}

// GenerateOptions contains language-specific generation settings.
type GenerateOptions struct {
	Go GoOptions `yaml:"go"`
}

// GoOptions controls the Go package prefix. A nil prefix allows inheritance.
type GoOptions struct {
	PackagePrefix *string `yaml:"package_prefix"`
}

// ParseGenerate reads and validates a v1 generation configuration.
func ParseGenerate(r io.Reader) (Generate, error) {
	raw, err := expandConfigYAML(r)
	if err != nil {
		return Generate{}, fmt.Errorf("expandConfigYAML: %w", err)
	}
	var result Generate
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	if err := decoder.Decode(&result); err != nil {
		return Generate{}, fmt.Errorf("decode easyp.gen.yaml: %w", err)
	}
	if result.Version == "" {
		result.Version = "v1"
	}
	if result.Version != "v1" {
		return Generate{}, fmt.Errorf("easyp.gen.yaml version must be v1")
	}
	if err := result.Generate.Managed.Validate(); err != nil {
		return Generate{}, fmt.Errorf("generate.managed: %w", err)
	}
	for i, plugin := range result.Plugins {
		if err := plugin.Validate(); err != nil {
			return Generate{}, fmt.Errorf("plugins[%d]: %w", i, err)
		}
	}
	return result, nil
}
