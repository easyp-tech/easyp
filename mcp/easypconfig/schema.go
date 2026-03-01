package easypconfig

import (
	"encoding/json"
	"sort"
	"sync"

	invjsonschema "github.com/invopop/jsonschema"
)

var (
	schemaCacheOnce sync.Once
	cachedByPath    map[string]map[string]any
)

func SchemaByPath() map[string]map[string]any {
	ensureSchemaCache()
	return cloneSchemaByPath(cachedByPath)
}

func MarshalConfigJSONSchema() ([]byte, error) {
	schema := reflectConfigSchema()
	return json.MarshalIndent(schema, "", "  ")
}

func ensureSchemaCache() {
	schemaCacheOnce.Do(func() {
		root := buildRootSchemaMap()
		cachedByPath = buildSchemaByPath(root)
	})
}

func buildRootSchemaMap() map[string]any {
	root := invSchemaToMap(reflectConfigSchema())
	if len(root) == 0 {
		return map[string]any{}
	}
	return root
}

func reflectConfigSchema() *invjsonschema.Schema {
	reflector := &invjsonschema.Reflector{
		Anonymous:      true,
		DoNotReference: true,
	}
	return reflector.Reflect(configSchemaRoot{})
}

func buildSchemaByPath(root map[string]any) map[string]map[string]any {
	if len(root) == 0 {
		return map[string]map[string]any{}
	}

	index := map[string]map[string]any{
		"$": root,
	}
	walkSchemaPaths(index, "$", root)
	return index
}

func walkSchemaPaths(index map[string]map[string]any, basePath string, schema map[string]any) {
	for _, key := range []string{"allOf", "anyOf", "oneOf"} {
		branches, ok := asSchemaArray(schema[key])
		if !ok {
			continue
		}
		for _, branch := range branches {
			walkSchemaPaths(index, basePath, branch)
		}
	}

	props, ok := asSchemaMap(schema["properties"])
	if ok {
		names := make([]string, 0, len(props))
		for name := range props {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			child, ok := asSchemaMap(props[name])
			if !ok {
				continue
			}
			childPath := joinSchemaPath(basePath, name)
			if _, exists := index[childPath]; !exists {
				index[childPath] = child
			}
			walkSchemaPaths(index, childPath, child)
		}
	}

	if items, ok := asSchemaMap(schema["items"]); ok {
		arrayPath := basePath + "[]"
		if _, exists := index[arrayPath]; !exists {
			index[arrayPath] = items
		}
		walkSchemaPaths(index, arrayPath, items)
	}
}

func joinSchemaPath(base, child string) string {
	if base == "$" {
		return child
	}
	return base + "." + child
}

func asSchemaArray(v any) ([]map[string]any, bool) {
	arr, ok := v.([]any)
	if !ok {
		return nil, false
	}
	out := make([]map[string]any, 0, len(arr))
	for _, item := range arr {
		m, ok := asSchemaMap(item)
		if !ok {
			continue
		}
		out = append(out, m)
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

func asSchemaMap(v any) (map[string]any, bool) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, false
	}
	return m, true
}

func invSchemaToMap(schema *invjsonschema.Schema) map[string]any {
	if schema == nil {
		return nil
	}

	data, err := json.Marshal(schema)
	if err != nil {
		return map[string]any{}
	}

	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return map[string]any{}
	}
	return out
}

func cloneSchemaMap(in map[string]any) map[string]any {
	return cloneJSON(in, map[string]any{})
}

func cloneSchemaByPath(in map[string]map[string]any) map[string]map[string]any {
	return cloneJSON(in, map[string]map[string]any{})
}

func cloneJSON[T any](in T, fallback T) T {
	data, err := json.Marshal(in)
	if err != nil {
		return fallback
	}
	var out T
	if err := json.Unmarshal(data, &out); err != nil {
		return fallback
	}
	return out
}
