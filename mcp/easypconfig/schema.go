package easypconfig

import (
	"encoding/json"
	"fmt"
	"sort"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// MarshalConfigJSONSchema returns the v1 easyp.yaml schema.
func MarshalConfigJSONSchema() ([]byte, error) {
	return v1.SchemaJSON("easyp")
}

// SchemaByPath returns independently owned easyp.yaml schema fragments.
func SchemaByPath() map[string]map[string]any {
	index, err := SchemaByPathFor(v1.PolicyFile)
	if err != nil {
		return nil
	}
	return index
}

// SchemaByPathFor indexes the JSON Schema of a v1 EasyP YAML configuration file.
func SchemaByPathFor(file string) (map[string]map[string]any, error) {
	_, schemaName, err := configFile(file)
	if err != nil {
		return nil, err
	}
	raw, err := v1.SchemaJSON(schemaName)
	if err != nil {
		return nil, fmt.Errorf("SchemaJSON: %w", err)
	}
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, fmt.Errorf("Unmarshal: %w", err)
	}
	index := map[string]map[string]any{"$": root}
	indexSchema(index, "$", root)
	return index, nil
}

func configFile(file string) (string, string, error) {
	switch file {
	case "", v1.PolicyFile:
		return v1.PolicyFile, "easyp", nil
	case v1.GenerateFile:
		return v1.GenerateFile, "easyp.gen", nil
	default:
		return "", "", fmt.Errorf("unknown config file %q", file)
	}
}

func indexSchema(index map[string]map[string]any, base string, schema map[string]any) {
	properties, ok := schema["properties"].(map[string]any)
	if ok {
		names := make([]string, 0, len(properties))
		for name := range properties {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			child, ok := properties[name].(map[string]any)
			if !ok {
				continue
			}
			path := name
			if base != "$" {
				path = base + "." + name
			}
			index[path] = child
			indexSchema(index, path, child)
		}
	}
	if item, ok := schema["items"].(map[string]any); ok {
		path := base + "[]"
		index[path] = item
		indexSchema(index, path, item)
	}
	for _, keyword := range []string{"oneOf", "anyOf", "allOf"} {
		branches, ok := schema[keyword].([]any)
		if !ok {
			continue
		}
		for _, branch := range branches {
			if child, ok := branch.(map[string]any); ok {
				indexSchema(index, base, child)
			}
		}
	}
}
