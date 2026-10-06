package v1

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// GenerateModule selects a module and optionally limits its generated sources.
// A string is shorthand for a module without paths or package filters.
type GenerateModule struct {
	Module   string   `yaml:"module"`
	Paths    []string `yaml:"paths,omitempty"`
	Packages []string `yaml:"packages,omitempty"`
}

// UnmarshalYAML accepts a module name or an object containing its selectors.
func (m *GenerateModule) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		if node.Tag != "!!str" {
			return fmt.Errorf("generate.modules entries must be strings or module objects")
		}
		*m = GenerateModule{Module: node.Value}
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			name := node.Content[i].Value
			if name != "module" && name != "paths" && name != "packages" {
				return fmt.Errorf("unknown generate.modules field %q", name)
			}
		}
		type plain GenerateModule
		if err := node.Decode((*plain)(m)); err != nil {
			return fmt.Errorf("Decode: %w", err)
		}
	default:
		return fmt.Errorf("generate.modules entries must be strings or module objects")
	}
	return nil
}

// MarshalYAML keeps the unfiltered shorthand compact.
func (m GenerateModule) MarshalYAML() (any, error) {
	if !m.HasSourceSelectors() {
		return m.Module, nil
	}
	type plain GenerateModule
	return plain(m), nil
}

// HasSourceSelectors reports whether this module restricts generated sources.
func (m GenerateModule) HasSourceSelectors() bool {
	return len(m.Paths) > 0 || len(m.Packages) > 0
}

// Validate checks the selected module name and optional source selectors.
func (m GenerateModule) Validate() error {
	if strings.TrimSpace(m.Module) == "" {
		return fmt.Errorf("module must not be empty")
	}
	if err := ValidatePackageSelectors(m.Packages); err != nil {
		return err
	}
	return ValidatePathSelectors(m.Paths)
}
