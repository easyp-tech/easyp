// Package schemagen writes JSON Schemas derived from the public v1 YAML models.
package schemagen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

const DefaultOutDir = "schemas"

// Options controls where v1 schema artifacts are written.
type Options struct {
	OutDir string
}

type schema struct {
	Schema               string             `json:"$schema,omitempty"`
	ID                   string             `json:"$id,omitempty"`
	Title                string             `json:"title,omitempty"`
	Description          string             `json:"description,omitempty"`
	Type                 string             `json:"type,omitempty"`
	Properties           map[string]*schema `json:"properties,omitempty"`
	AdditionalProperties any                `json:"additionalProperties,omitempty"`
	Items                *schema            `json:"items,omitempty"`
	OneOf                []*schema          `json:"oneOf,omitempty"`
	Required             []string           `json:"required,omitempty"`
	Enum                 []string           `json:"enum,omitempty"`
	Const                any                `json:"const,omitempty"`
	Pattern              string             `json:"pattern,omitempty"`
	MinLength            int                `json:"minLength,omitempty"`
}

// Run creates versioned schemas and latest aliases for the three YAML files.
func Run(opts Options) error {
	dir := opts.OutDir
	if dir == "" {
		dir = DefaultOutDir
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("MkdirAll: %w", err)
	}
	for name, document := range documents() {
		versioned, latest := schemaNames(name)
		data, err := json.MarshalIndent(document, "", "  ")
		if err != nil {
			return fmt.Errorf("MarshalIndent: %w", err)
		}
		data = append(data, '\n')
		for _, filename := range []string{versioned, latest} {
			if err := os.WriteFile(filepath.Join(dir, filename), data, 0o644); err != nil {
				return fmt.Errorf("WriteFile: %w", err)
			}
		}
	}
	return nil
}

func schemaNames(name string) (string, string) {
	return name + "-v1.schema.json", name + ".schema.json"
}

func documents() map[string]*schema {
	policy := fromType(reflect.TypeOf(v1.Policy{}))
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

	generate := fromType(reflect.TypeOf(v1.Generate{}))
	generate.Properties["version"] = &schema{Type: "string", Const: "v1"}
	plugin := generate.Properties["plugins"].Items
	plugin.Required = []string{"out"}
	plugin.Properties["opts"] = &schema{OneOf: []*schema{
		{Type: "array", Items: &schema{Type: "string"}},
		{Type: "object", AdditionalProperties: &schema{OneOf: []*schema{
			{Type: "string"}, {Type: "number"}, {Type: "boolean"},
			{Type: "array", Items: &schema{OneOf: []*schema{{Type: "string"}, {Type: "number"}, {Type: "boolean"}}}},
		}}},
	}}

	lock := fromType(reflect.TypeOf(v1.Lock{}))
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
