package v1

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/config"
)

// PluginOptions preserves repeated values for each plugin flag.
type PluginOptions map[string][]string

// UnmarshalYAML accepts an options mapping or a sequence of flag strings.
func (o *PluginOptions) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.MappingNode:
		var mapped config.PluginOpts
		if err := mapped.UnmarshalYAML(node); err != nil {
			return fmt.Errorf("UnmarshalYAML: %w", err)
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
