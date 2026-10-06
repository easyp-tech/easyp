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

// GenerateTargets selects module sources and their managed descriptor options.
type GenerateTargets struct {
	Modules  []GenerateModule   `yaml:"modules"`
	Packages []string           `yaml:"packages"`
	Paths    []string           `yaml:"paths"`
	Managed  config.ManagedMode `yaml:"managed"`
}

// GenerateOptions contains language-specific generation settings.
type GenerateOptions struct {
	Go GoOptions `yaml:"go"`
}

// HasSourceSelectors reports whether either the project or a selected module
// restricts generated sources.
func (g GenerateTargets) HasSourceSelectors() bool {
	if len(g.Packages) > 0 || len(g.Paths) > 0 {
		return true
	}
	for _, module := range g.Modules {
		if module.HasSourceSelectors() {
			return true
		}
	}
	return false
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
	if err := requireSingleYAMLDocument(decoder, GenerateFile); err != nil {
		return Generate{}, err
	}
	if result.Version == "" {
		result.Version = "v1"
	}
	if result.Version != "v1" {
		return Generate{}, fmt.Errorf("easyp.gen.yaml version must be v1")
	}
	for i, module := range result.Generate.Modules {
		if err := module.Validate(); err != nil {
			return Generate{}, fmt.Errorf("generate.modules[%d]: %w", i, err)
		}
	}
	if err := ValidatePackageSelectors(result.Generate.Packages); err != nil {
		return Generate{}, err
	}
	if err := ValidatePathSelectors(result.Generate.Paths); err != nil {
		return Generate{}, err
	}
	if err := result.Generate.Managed.Validate(); err != nil {
		return Generate{}, fmt.Errorf("generate.managed: %w", err)
	}
	for i, plugin := range result.Plugins {
		if err := plugin.Validate(); err != nil {
			return Generate{}, fmt.Errorf("plugins[%d]: %w", i, err)
		}
	}
	if err := yamlValidationError(GenerateFile, validateExpandedV1YAML(raw, generateSchema)); err != nil {
		return Generate{}, err
	}
	return result, nil
}
