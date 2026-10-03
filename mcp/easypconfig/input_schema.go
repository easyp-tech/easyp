package easypconfig

import v1 "github.com/easyp-tech/easyp/internal/config/v1"

// describeInputSchema constrains the reference selectors, not filesystem access.
// The typed handler still validates bounds for direct Go callers.
func describeInputSchema() map[string]any {
	return map[string]any{
		"type": "object", "additionalProperties": false,
		"properties": map[string]any{
			"file":             map[string]any{"type": "string", "enum": []string{"", v1.PolicyFile, v1.GenerateFile, v1.ModuleFile, v1.LockFile}, "description": "Known format selector, never a filesystem path. Empty/omitted selects easyp.yaml."},
			"path":             map[string]any{"type": "string", "description": "Field or directive path. Root is empty or $. Array indices normalize to []."},
			"include_schema":   map[string]any{"type": "boolean", "description": "Include JSON Schema for YAML or grammar for text; default true."},
			"include_fields":   map[string]any{"type": "boolean", "description": "Include field or directive documentation; default true."},
			"include_examples": map[string]any{"type": "boolean", "description": "Include parseable examples; default true."},
			"include_children": map[string]any{"type": "boolean", "description": "Include descendants of the selected field or directive; default true."},
			"examples_limit":   map[string]any{"type": "integer", "minimum": 1, "maximum": 50, "description": "Maximum examples, 1 through 50; default 10."},
		},
	}
}
