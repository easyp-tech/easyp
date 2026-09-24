package easypconfig

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

const (
	// ToolName is the stable MCP tool name.
	ToolName = "easyp_config_describe"
	// SchemaVersion identifies the v1 configuration model returned by the tool.
	SchemaVersion = "v1"
)

var arrayIndex = regexp.MustCompile(`\[(?:\d+|\*)\]`)

// DescribeInput selects a v1 configuration file and a dot path within it.
type DescribeInput struct {
	File            string `json:"file,omitempty" jsonschema:"Config filename: easyp.yaml or easyp.gen.yaml. Defaults to easyp.yaml."`
	Path            string `json:"path,omitempty" jsonschema:"Dot path to a section or field. Empty means the full file."`
	IncludeSchema   *bool  `json:"include_schema,omitempty" jsonschema:"Include the JSON Schema fragment. Default true."`
	IncludeFields   *bool  `json:"include_fields,omitempty" jsonschema:"Include field documentation. Default true."`
	IncludeExamples *bool  `json:"include_examples,omitempty" jsonschema:"Include valid v1 examples. Default true."`
	IncludeChildren *bool  `json:"include_children,omitempty" jsonschema:"Include descendants of the selected path. Default true."`
	ExamplesLimit   *int   `json:"examples_limit,omitempty" jsonschema:"Maximum examples, from 1 to 50. Default 10."`
}

// FieldDoc describes one field from the current v1 configuration schema.
type FieldDoc struct {
	Path          string   `json:"path"`
	Type          string   `json:"type"`
	Required      bool     `json:"required"`
	Description   string   `json:"description"`
	AllowedValues []string `json:"allowed_values,omitempty"`
	DefaultValue  string   `json:"default_value,omitempty"`
	Examples      []string `json:"examples,omitempty"`
	Notes         []string `json:"notes,omitempty"`
}

// Example is a v1 configuration snippet that can be parsed by EasyP.
type Example struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	YAML        string   `json:"yaml"`
	Paths       []string `json:"paths,omitempty"`
}

// DescribeOutput contains the selected schema, field details, and examples.
type DescribeOutput struct {
	SchemaVersion string         `json:"schema_version"`
	File          string         `json:"file"`
	SelectedPath  string         `json:"selected_path"`
	Schema        map[string]any `json:"schema,omitempty"`
	Fields        []FieldDoc     `json:"fields,omitempty"`
	Examples      []Example      `json:"examples,omitempty"`
	Notes         []string       `json:"notes,omitempty"`
}

// Describe explains a v1 config path using the same JSON Schema as schema-gen.
func Describe(input DescribeInput) (DescribeOutput, error) {
	file, _, err := configFile(input.File)
	if err != nil {
		return DescribeOutput{}, err
	}
	index, err := SchemaByPathFor(file)
	if err != nil {
		return DescribeOutput{}, fmt.Errorf("SchemaByPathFor: %w", err)
	}
	path := normalizePath(input.Path)
	schema, ok := index[path]
	if !ok {
		return DescribeOutput{}, fmt.Errorf("unknown path %q in %s", input.Path, file)
	}

	out := DescribeOutput{SchemaVersion: SchemaVersion, File: file, SelectedPath: path}
	if enabled(input.IncludeSchema) {
		out.Schema = schema
	}
	paths := selectedPaths(index, path, enabled(input.IncludeChildren))
	if enabled(input.IncludeFields) {
		out.Fields = describeFields(file, index, paths)
	}
	if enabled(input.IncludeExamples) {
		out.Examples = selectExamples(file, path, exampleLimit(input.ExamplesLimit))
	}
	out.Notes = notesFor(file, path)
	return out, nil
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || path == "$" || strings.EqualFold(path, "root") {
		return "$"
	}
	path = strings.TrimPrefix(path, "$.")
	path = strings.TrimPrefix(path, ".")
	path = arrayIndex.ReplaceAllString(path, "[]")
	return strings.TrimSuffix(path, ".")
}

func enabled(value *bool) bool {
	return value == nil || *value
}

func exampleLimit(value *int) int {
	if value == nil {
		return 10
	}
	if *value < 1 {
		return 1
	}
	if *value > 50 {
		return 50
	}
	return *value
}

func selectedPaths(index map[string]map[string]any, path string, children bool) []string {
	if !children {
		return []string{path}
	}
	paths := make([]string, 0, len(index))
	for candidate := range index {
		if within(path, candidate) {
			paths = append(paths, candidate)
		}
	}
	sort.Strings(paths)
	return paths
}

func within(parent, path string) bool {
	return parent == "$" || path == parent || strings.HasPrefix(path, parent+".") || strings.HasPrefix(path, parent+"[]")
}

func describeFields(file string, index map[string]map[string]any, paths []string) []FieldDoc {
	fields := make([]FieldDoc, 0, len(paths))
	for _, path := range paths {
		if path == "$" || strings.HasSuffix(path, "[]") {
			continue
		}
		fields = append(fields, schemaField(file, index, path))
	}
	return fields
}

func schemaField(file string, index map[string]map[string]any, path string) FieldDoc {
	schema := index[path]
	field := FieldDoc{Path: path, Type: schemaType(schema), Description: fieldDescription(file, path)}
	field.Notes = notesFor(file, path)
	if description, ok := schema["description"].(string); ok && description != "" {
		field.Description = description
	}
	if values, ok := schema["enum"].([]any); ok {
		for _, value := range values {
			if name, ok := value.(string); ok {
				field.AllowedValues = append(field.AllowedValues, name)
			}
		}
	}
	parent := "$"
	name := path
	if lastDot := strings.LastIndex(path, "."); lastDot >= 0 {
		parent, name = path[:lastDot], path[lastDot+1:]
	}
	if parentSchema, ok := index[parent]; ok {
		if required, ok := parentSchema["required"].([]any); ok {
			for _, item := range required {
				field.Required = field.Required || item == name
			}
		}
	}
	return field
}

func schemaType(schema map[string]any) string {
	if typ, ok := schema["type"].(string); ok {
		return typ
	}
	if _, ok := schema["oneOf"]; ok {
		return "oneOf"
	}
	if _, ok := schema["anyOf"]; ok {
		return "anyOf"
	}
	return "any"
}

func fieldDescription(file, path string) string {
	if description := descriptions[file][path]; description != "" {
		return description
	}
	return "Configuration value."
}
