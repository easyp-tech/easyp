package v1

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/rules"
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
	AllOf                []*schema          `json:"allOf,omitempty"`
	If                   *schema            `json:"if,omitempty"`
	Then                 *schema            `json:"then,omitempty"`
	Not                  *schema            `json:"not,omitempty"`
	Required             []string           `json:"required,omitempty"`
	Enum                 []string           `json:"enum,omitempty"`
	Const                any                `json:"const,omitempty"`
	Pattern              string             `json:"pattern,omitempty"`
	MinLength            int                `json:"minLength,omitempty"`
	MaxItems             *int               `json:"maxItems,omitempty"`
}

// policyExtendsPattern accepts a local ./ or ../ policy path, an exact declared
// module identity with an optional # policy fragment, and the RFC form that
// appends a relative policy path to a declared module identity. Whitespace, a
// repeated version and repeated fragments are rejected structurally; the CLI
// additionally resolves the identity against the consumer module graph.
const policyExtendsPattern = `^(|\.\.?|\.\.?/[^\r\n\\@]+|[^\s@#]+/[^\s@#]*(#[^\s@#]*)?)$`

// policyExtendsDescription documents the delivery contract shared by both
// sections. Only linters and linters-settings inherit from a lint base, and only
// the breaking section inherits from a breaking base; issues.exclude-rules is
// never inherited.
const policyExtendsDescription = "Base policy for this section only, applied before local adjustments. " +
	"Use ./ or ../ for a policy file or directory relative to this file, or a module identity already declared in protobuf.mod and locked in protobuf.lock, optionally with #<policy-file-or-directory>. " +
	"A module reference without # loads that module's easyp.yaml. Versions are not repeated here and no new module or revision is acquired. " +
	"Remote policies are read from the verified module contents of the consuming module; a base's own relative references stay inside its module. " +
	"Only the referenced section and linters-settings are inherited: issues.exclude-rules, generation settings and the other section are not. " +
	"Base breaking.ignore paths and the baseline are applied at this file's location, not in the module cache."

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
	for _, field := range []string{"enable", "disable"} {
		policy.Properties["linters"].Properties[field].Items.Enum = rules.AllLintUseValues()
	}
	policy.Properties["issues"].Properties["exclude-rules"].Items.Properties["linters"].Items.Enum = rules.AllLintUseValues()
	policy.Properties["linters"].Properties["default"] = &schema{Type: "string", Enum: []string{"MINIMAL", "BASIC", "STANDARD", "COMMENTS"}}
	policy.Properties["linters-settings"] = &schema{
		Type: "object",
		Properties: map[string]*schema{
			"ENUM_ZERO_VALUE_SUFFIX": suffixSettings(),
			"SERVICE_SUFFIX":         suffixSettings(),
		},
		AdditionalProperties: false,
	}
	policy.Properties["breaking"].Properties["baseline"].Pattern = "^(git:.+)?$"
	policy.Properties["breaking"].Properties["categories"].Items.Enum = []string{breakingCategoryFile}
	policy.Properties["breaking"].Properties["categories"].Description = "FILE adds checks for declarations moved between files; existing compatibility checks remain enabled."

	for _, section := range []string{"linters", "breaking"} {
		extends := policy.Properties[section].Properties["extends"]
		extends.Const = nil
		extends.Pattern = policyExtendsPattern
		extends.Description = policyExtendsDescription
	}
	policy.Properties["issues"].Properties["exclude-rules"].Items.Properties["path"].Description = "Relative to the policy providing this issues section. Supports *, ?, character classes and ** directory segments; a literal directory includes its subtree."
	policy.Properties["breaking"].Properties["ignore_unstable"].Description = "Ignore declarations in packages with unstable version suffixes; stable package checks remain enabled."

	generate := fromType(reflect.TypeFor[Generate]())
	generate.Properties["version"] = &schema{Type: "string", Const: "v1"}
	zero := 0
	generate.Properties["generate"].Properties["packages"].MaxItems = &zero
	generate.Properties["generate"].Properties["packages"].Description = "Reserved package selector. Select complete modules; only an omitted or empty list is accepted."
	generate.Properties["options"].Properties["go"].Properties["package_prefix"].Description = "Sets go_package without enabling managed defaults for other languages; full managed mode requires generate.managed.enabled."
	plugin := generate.Properties["plugins"].Items
	plugin.Required = []string{"out"}
	plugin.Properties["with_imports"].Description = "Generate code for transitive imports with this plugin only; defaults to false. Independent from descriptor --include_imports."
	policy.Properties["linters"].Properties["allow_comment_ignores"].Description = "Enable scoped easyp:disable, nolint and buf:lint:ignore comments; defaults to true. State is local to each effective lint policy."
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
	for _, selector := range disable.AnyOf {
		selector.Properties = map[string]*schema{selector.Required[0]: {MinLength: 1}}
	}
	disable.Not = &schema{Required: []string{"file_option", "field_option"}}
	override := managed.Properties["override"].Items
	override.OneOf = requiredAnyOf("file_option", "field_option")
	override.Required = []string{"value"}
	for _, rule := range []*schema{disable, override} {
		rule.AllOf = append(rule.AllOf, &schema{
			If:   &schema{Required: []string{"field"}, Properties: map[string]*schema{"field": {MinLength: 1}}},
			Then: &schema{Required: []string{"field_option"}},
		})
	}
	for _, options := range []struct {
		field    string
		metadata []core.ManagedOptionMetadata
	}{
		{field: "file_option", metadata: core.ManagedFileOptionMetadata()},
		{field: "field_option", metadata: core.ManagedFieldOptionMetadata()},
	} {
		for _, option := range options.metadata {
			disable.Properties[options.field].Enum = append(disable.Properties[options.field].Enum, option.Name)
			override.Properties[options.field].Enum = append(override.Properties[options.field].Enum, option.Name)
			override.AllOf = append(override.AllOf, &schema{
				If: &schema{Required: []string{options.field}, Properties: map[string]*schema{
					options.field: {Const: option.Name},
				}},
				Then: &schema{Properties: map[string]*schema{
					"value": {Type: option.ValueType, Enum: option.Enum},
				}},
			})
		}
	}

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
