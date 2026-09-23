package v1

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/config"
)

// Generate is the consumer-side v1 file. Producer policy and legacy inputs
// remain outside this file; managed descriptor options retain their v0 shape.
type Generate struct {
	Version                  string `yaml:"version"`
	InheritedGoPackagePrefix bool   `yaml:"-"`
	Generate                 struct {
		Modules  []string           `yaml:"modules"`
		Packages []string           `yaml:"packages"`
		Managed  config.ManagedMode `yaml:"managed"`
	} `yaml:"generate"`
	Plugins []Plugin `yaml:"plugins"`
	Options struct {
		Go struct {
			PackagePrefix *string `yaml:"package_prefix"`
		} `yaml:"go"`
	} `yaml:"options"`
}

type Plugin struct {
	Name    string        `yaml:"name"`
	Remote  string        `yaml:"remote"`
	Version string        `yaml:"version"`
	Out     string        `yaml:"out"`
	Opts    PluginOptions `yaml:"opts"`
}

type PluginOptions map[string][]string

func (o *PluginOptions) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.MappingNode:
		var mapped config.PluginOpts
		if err := mapped.UnmarshalYAML(node); err != nil {
			return err
		}
		*o = PluginOptions(mapped)
		return nil
	case yaml.SequenceNode:
		result := make(PluginOptions)
		for i, item := range node.Content {
			if item.Kind != yaml.ScalarNode {
				return fmt.Errorf("opts[%d] must be a flag string", i)
			}
			key, value, _ := strings.Cut(item.Value, "=")
			if key == "" {
				return fmt.Errorf("opts[%d] has an empty flag name", i)
			}
			result[key] = append(result[key], value)
		}
		*o = result
		return nil
	default:
		return fmt.Errorf("plugin opts must be a mapping or string list")
	}
}

func ParseGenerate(r io.Reader) (Generate, error) {
	var result Generate
	decoder := yaml.NewDecoder(r)
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
	for i, plugin := range result.Plugins {
		if (plugin.Name == "") == (plugin.Remote == "") {
			return Generate{}, fmt.Errorf("plugins[%d]: exactly one of name or remote is required", i)
		}
		if plugin.Out == "" {
			return Generate{}, fmt.Errorf("plugins[%d]: out is required", i)
		}
		if strings.Contains(plugin.Version, ":latest") {
			return Generate{}, fmt.Errorf("plugins[%d]: latest version is not reproducible", i)
		}
		if plugin.Remote != "" {
			if !semver.IsValid(plugin.Version) {
				return Generate{}, fmt.Errorf("plugins[%d]: remote plugin requires a pinned semantic version", i)
			}
			lastSegment := plugin.Remote[strings.LastIndex(plugin.Remote, "/")+1:]
			if strings.Contains(lastSegment, ":") {
				return Generate{}, fmt.Errorf("plugins[%d]: specify the remote plugin version only in version", i)
			}
		} else if plugin.Version != "" {
			return Generate{}, fmt.Errorf("plugins[%d]: local and bundled plugin versions are selected by the executable; version cannot be verified", i)
		}
	}
	return result, nil
}
