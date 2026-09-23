package v1

import (
	"fmt"
	"reflect"
	"strings"

	yamlvalidator "github.com/Yakwilik/go-yamlvalidator"
	"github.com/Yakwilik/go-yamlvalidator/pkg/valuevalidator"
	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/config"
)

// ValidatePolicyYAML reports structural errors with source coordinates.
// ParsePolicy remains responsible for semantic policy checks.
func ValidatePolicyYAML(raw []byte) []config.ValidationIssue {
	schema := yamlSchemaFor(reflect.TypeFor[Policy]())
	schema.AllowedKeys["version"].Validators = []yamlvalidator.ValueValidator{
		valuevalidator.EnumValidator{Allowed: []string{"v1"}},
	}
	schema.AllowedKeys["linters"].AllowedKeys["default"].Validators = []yamlvalidator.ValueValidator{
		valuevalidator.EnumValidator{Allowed: []string{"MINIMAL", "BASIC", "STANDARD", "COMMENTS"}},
	}
	return validateV1YAML(raw, schema)
}

// ValidateGenerateYAML reports structural errors with source coordinates.
// ParseGenerate remains responsible for semantic plugin and managed-mode checks.
func ValidateGenerateYAML(raw []byte) []config.ValidationIssue {
	schema := yamlSchemaFor(reflect.TypeFor[Generate]())
	schema.AllowedKeys["version"].Validators = []yamlvalidator.ValueValidator{
		valuevalidator.EnumValidator{Allowed: []string{"v1"}},
	}
	plugin := schema.AllowedKeys["plugins"].ItemSchema
	plugin.AllowedKeys["out"].Required = true
	plugin.ExactlyOneOf = []string{"name", "remote"}
	plugin.AllowedKeys["opts"] = &yamlvalidator.FieldSchema{
		Type:             yamlvalidator.TypeAny,
		UnknownKeyPolicy: yamlvalidator.UnknownKeyIgnore,
		Validators:       []yamlvalidator.ValueValidator{pluginOptionsValidator{}},
	}
	managed := schema.AllowedKeys["generate"].AllowedKeys["managed"]
	disable := managed.AllowedKeys["disable"].ItemSchema
	disable.AnyOf = [][]string{{"module"}, {"package"}, {"path"}, {"file_option"}, {"field_option"}, {"field"}}
	disable.MutuallyExclusive = []string{"file_option", "field_option"}
	override := managed.AllowedKeys["override"].ItemSchema
	override.ExactlyOneOf = []string{"file_option", "field_option"}
	override.AllowedKeys["value"].Required = true
	return validateV1YAML(raw, schema)
}

func validateV1YAML(raw []byte, schema *yamlvalidator.FieldSchema) []config.ValidationIssue {
	result := yamlvalidator.NewValidator(schema).ValidateWithOptions(raw, yamlvalidator.ValidationContext{
		StrictKeys:  true,
		StrictTypes: true,
	})
	issues := make([]config.ValidationIssue, 0, len(result.Collector.All()))
	for _, issue := range result.Collector.All() {
		severity := config.SeverityWarn
		if issue.Level == yamlvalidator.LevelError {
			severity = config.SeverityError
		}
		message := issue.Message
		if issue.Expected != "" {
			message += fmt.Sprintf(" (expected %s)", issue.Expected)
		}
		if issue.Got != "" {
			message += fmt.Sprintf(" (got %s)", issue.Got)
		}
		if issue.Path != "" {
			message += fmt.Sprintf(" (path: %s)", issue.Path)
		}
		issues = append(issues, config.ValidationIssue{
			Code: "yaml_validation", Message: message,
			Line: issue.Line, Column: issue.Column, Severity: severity,
		})
	}
	return issues
}

// yamlSchemaFor derives the field tree from the same YAML models used by the
// runtime parsers. File-specific rules are added by the two entry points above.
func yamlSchemaFor(typ reflect.Type) *yamlvalidator.FieldSchema {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	switch typ.Kind() {
	case reflect.Struct:
		fields := make(map[string]*yamlvalidator.FieldSchema)
		for i := range typ.NumField() {
			field := typ.Field(i)
			if !field.IsExported() {
				continue
			}
			name := strings.Split(field.Tag.Get("yaml"), ",")[0]
			if name == "-" {
				continue
			}
			if name == "" {
				name = strings.ToLower(field.Name)
			}
			fields[name] = yamlSchemaFor(field.Type)
		}
		return &yamlvalidator.FieldSchema{Type: yamlvalidator.TypeMap, AllowedKeys: fields}
	case reflect.Slice, reflect.Array:
		return &yamlvalidator.FieldSchema{Type: yamlvalidator.TypeSequence, ItemSchema: yamlSchemaFor(typ.Elem())}
	case reflect.Map:
		return &yamlvalidator.FieldSchema{Type: yamlvalidator.TypeMap, AdditionalProperties: yamlSchemaFor(typ.Elem())}
	case reflect.String:
		return &yamlvalidator.FieldSchema{Type: yamlvalidator.TypeString}
	case reflect.Bool:
		return &yamlvalidator.FieldSchema{Type: yamlvalidator.TypeBool}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &yamlvalidator.FieldSchema{Type: yamlvalidator.TypeInt}
	case reflect.Float32, reflect.Float64:
		return &yamlvalidator.FieldSchema{Type: yamlvalidator.TypeFloat}
	default:
		return &yamlvalidator.FieldSchema{Type: yamlvalidator.TypeAny, UnknownKeyPolicy: yamlvalidator.UnknownKeyIgnore}
	}
}

// pluginOptionsValidator mirrors the two YAML forms accepted by PluginOptions:
// a mapping of scalar/list values or a list of flag strings.
type pluginOptionsValidator struct{}

func (pluginOptionsValidator) Validate(node *yaml.Node, path string, ctx *yamlvalidator.ValidationContext) {
	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			key, value := node.Content[i], node.Content[i+1]
			optionPath := path + "." + key.Value
			switch value.Kind {
			case yaml.ScalarNode:
				// Scalar values are converted to plugin flag strings by the parser.
			case yaml.SequenceNode:
				for index, item := range value.Content {
					if item.Kind != yaml.ScalarNode {
						addYAMLError(ctx, item, fmt.Sprintf("%s[%d]", optionPath, index), "plugin option list items must be scalar")
					}
				}
			default:
				addYAMLError(ctx, value, optionPath, "plugin option must be a scalar or list of scalars")
			}
		}
	case yaml.SequenceNode:
		for index, item := range node.Content {
			if item.Kind != yaml.ScalarNode {
				addYAMLError(ctx, item, fmt.Sprintf("%s[%d]", path, index), "plugin option must be a flag string")
			}
		}
	default:
		addYAMLError(ctx, node, path, "plugin opts must be a mapping or list of flag strings")
	}
}

func addYAMLError(ctx *yamlvalidator.ValidationContext, node *yaml.Node, path, message string) {
	ctx.AddError(yamlvalidator.ValidationError{
		Level: yamlvalidator.LevelError, Path: path,
		Line: node.Line, Column: node.Column, Message: message,
	})
}
