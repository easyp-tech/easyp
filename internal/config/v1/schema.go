package v1

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type schema struct {
	Schema               string             `json:"$schema,omitempty"`
	Title                string             `json:"title,omitempty"`
	Description          string             `json:"description,omitempty"`
	Type                 string             `json:"type,omitempty"`
	Properties           map[string]*schema `json:"properties,omitempty"`
	AdditionalProperties any                `json:"additionalProperties,omitempty"`
	Items                *schema            `json:"items,omitempty"`
	OneOf                []*schema          `json:"oneOf,omitempty"`
	AnyOf                []*schema          `json:"anyOf,omitempty"`
	Not                  *schema            `json:"not,omitempty"`
	Required             []string           `json:"required,omitempty"`
	Enum                 []string           `json:"enum,omitempty"`
	Const                any                `json:"const,omitempty"`
	Pattern              string             `json:"pattern,omitempty"`
	MinLength            int                `json:"minLength,omitempty"`
}

// SchemaJSON returns the JSON Schema used to validate a v1 YAML file.
// schema-gen writes these same bytes to disk.
func SchemaJSON(name string) ([]byte, error) {
	document, ok := documents()[name]
	if !ok {
		return nil, fmt.Errorf("unknown v1 schema %q", name)
	}
	data, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("MarshalIndent: %w", err)
	}
	return append(data, '\n'), nil
}

func documents() map[string]*schema {
	policy := fromType(reflect.TypeFor[Policy]())
	policy.Properties["version"] = &schema{Type: "string", Const: "v1"}
	policy.Properties["linters"].Properties["default"] = &schema{Type: "string", Enum: []string{"MINIMAL", "BASIC", "STANDARD", "COMMENTS"}}
	policy.Properties["linters-settings"] = &schema{
		Type: "object",
		Properties: map[string]*schema{
			"ENUM_ZERO_VALUE_SUFFIX": suffixSettings(),
			"SERVICE_SUFFIX":         suffixSettings(),
		},
		AdditionalProperties: false,
	}
	policy.Properties["breaking"].Properties["baseline"].Pattern = "^git:.+$"

	generate := fromType(reflect.TypeFor[Generate]())
	generate.Properties["version"] = &schema{Type: "string", Const: "v1"}
	plugin := generate.Properties["plugins"].Items
	plugin.Required = []string{"out"}
	plugin.OneOf = []*schema{
		{Required: []string{"name"}},
		{Required: []string{"path"}},
		{Required: []string{"command"}},
		{Required: []string{"remote"}},
	}
	plugin.Properties["opts"] = &schema{OneOf: []*schema{
		{Type: "array", Items: &schema{Type: "string"}},
		{Type: "object", AdditionalProperties: &schema{OneOf: []*schema{
			{Type: "string"}, {Type: "number"}, {Type: "boolean"},
			{Type: "array", Items: &schema{OneOf: []*schema{{Type: "string"}, {Type: "number"}, {Type: "boolean"}}}},
		}}},
	}}
	managed := generate.Properties["generate"].Properties["managed"]
	disable := managed.Properties["disable"].Items
	disable.AnyOf = requiredAnyOf("module", "package", "path", "file_option", "field_option", "field")
	disable.Not = &schema{Required: []string{"file_option", "field_option"}}
	override := managed.Properties["override"].Items
	override.OneOf = requiredAnyOf("file_option", "field_option")
	override.Required = []string{"value"}

	lock := fromType(reflect.TypeFor[Lock]())
	lock.Required = []string{"version"}
	lock.Properties["version"] = &schema{Type: "integer", Const: 1}
	entry := lock.Properties["modules"].Items
	entry.Required = []string{"source", "version", "commit", "hash"}
	entry.Properties["source"].MinLength = 1
	entry.Properties["version"].Pattern = `^(v[0-9]+\.[0-9]+\.[0-9]+([+-][0-9A-Za-z.-]+)?|[0-9a-fA-F]{40}|[0-9a-fA-F]{64})$`
	entry.Properties["commit"].Pattern = `^([0-9a-fA-F]{40}|[0-9a-fA-F]{64})$`
	entry.Properties["hash"].Pattern = `^h1:[A-Za-z0-9+/]{43}=$`

	return map[string]*schema{
		"easyp":         document("easyp.yaml v1", policy),
		"easyp.gen":     document("easyp.gen.yaml v1", generate),
		"protobuf.lock": document("protobuf.lock v1", lock),
	}
}

func requiredAnyOf(fields ...string) []*schema {
	conditions := make([]*schema, 0, len(fields))
	for _, field := range fields {
		conditions = append(conditions, &schema{Required: []string{field}})
	}
	return conditions
}

func suffixSettings() *schema {
	return &schema{Type: "object", Properties: map[string]*schema{"suffix": {Type: "string"}}, AdditionalProperties: false}
}

func document(title string, body *schema) *schema {
	body.Schema = "https://json-schema.org/draft/2020-12/schema"
	body.Title = title
	body.Description = "Structural v1 YAML schema. EasyP CLI validation applies additional semantic checks."
	return body
}

func fromType(typ reflect.Type) *schema {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	switch typ.Kind() {
	case reflect.Struct:
		result := &schema{Type: "object", Properties: make(map[string]*schema), AdditionalProperties: false}
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			name := strings.Split(field.Tag.Get("yaml"), ",")[0]
			if name == "-" || !field.IsExported() {
				continue
			}
			if name == "" {
				name = strings.ToLower(field.Name)
			}
			result.Properties[name] = fromType(field.Type)
		}
		return result
	case reflect.Slice, reflect.Array:
		return &schema{Type: "array", Items: fromType(typ.Elem())}
	case reflect.Map:
		return &schema{Type: "object", AdditionalProperties: fromType(typ.Elem())}
	case reflect.String:
		return &schema{Type: "string"}
	case reflect.Bool:
		return &schema{Type: "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &schema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return &schema{Type: "number"}
	default:
		return &schema{}
	}
}
