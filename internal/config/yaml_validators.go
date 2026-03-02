package config

import (
	"fmt"

	v "github.com/yakwilikk/go-yamlvalidator"
	"gopkg.in/yaml.v3"
)

func yamlKindName(kind yaml.Kind) string {
	switch kind {
	case yaml.ScalarNode:
		return "ScalarNode"
	case yaml.SequenceNode:
		return "SequenceNode"
	case yaml.MappingNode:
		return "MappingNode"
	case yaml.AliasNode:
		return "AliasNode"
	default:
		return fmt.Sprintf("%d", kind)
	}
}

// directoryValidator allows directory to be either a string or a map with required path and optional root.
type directoryValidator struct{}

func (directoryValidator) Validate(node *yaml.Node, path string, ctx *v.ValidationContext) {
	switch node.Kind {
	case yaml.ScalarNode:
		// Only string scalars are allowed for the shorthand form.
		if node.Tag == "!!null" || node.Tag == "!!bool" || node.Tag == "!!int" || node.Tag == "!!float" {
			ctx.AddError(v.ValidationError{
				Level:   v.LevelError,
				Path:    path,
				Line:    node.Line,
				Column:  node.Column,
				Message: "directory must be a string or mapping",
				Got:     node.Tag,
			})
		}
	case yaml.MappingNode:
		requiredPath := false
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]
			switch keyNode.Value {
			case "path":
				requiredPath = true
				if valNode.Kind != yaml.ScalarNode {
					ctx.AddError(v.ValidationError{
						Level:   v.LevelError,
						Path:    path + ".path",
						Line:    valNode.Line,
						Column:  valNode.Column,
						Message: "path must be a string",
					})
				}
			case "root":
				if valNode.Kind != yaml.ScalarNode {
					ctx.AddError(v.ValidationError{
						Level:   v.LevelError,
						Path:    path + ".root",
						Line:    valNode.Line,
						Column:  valNode.Column,
						Message: "root must be a string",
					})
				}
			default:
				ctx.AddError(v.ValidationError{
					Level:   v.LevelWarning,
					Path:    path + "." + keyNode.Value,
					Line:    keyNode.Line,
					Column:  keyNode.Column,
					Message: "unknown field under directory",
					Got:     keyNode.Value,
				})
			}
		}
		if !requiredPath {
			ctx.AddError(v.ValidationError{
				Level:   v.LevelError,
				Path:    path + ".path",
				Line:    node.Line,
				Column:  node.Column,
				Message: "directory.path is required",
			})
		}
	default:
		ctx.AddError(v.ValidationError{
			Level:   v.LevelError,
			Path:    path,
			Line:    node.Line,
			Column:  node.Column,
			Message: "directory must be string or mapping",
			Got:     yamlKindName(node.Kind),
		})
	}
}

// pluginSourceValidator ensures exactly one plugin source field is set.
type pluginSourceValidator struct{}

func (pluginSourceValidator) Validate(node *yaml.Node, path string, ctx *v.ValidationContext) {
	if node.Kind != yaml.MappingNode {
		return
	}
	has := func(k string) bool {
		for i := 0; i < len(node.Content); i += 2 {
			if node.Content[i].Value == k {
				return true
			}
		}
		return false
	}
	var count int
	for _, k := range []string{"name", "remote", "path", "command"} {
		if has(k) {
			count++
		}
	}
	if count == 0 {
		ctx.AddError(v.ValidationError{
			Level:   v.LevelError,
			Path:    path,
			Line:    node.Line,
			Column:  node.Column,
			Message: "plugins item must have one of name/remote/path/command",
		})
	} else if count > 1 {
		ctx.AddError(v.ValidationError{
			Level:   v.LevelError,
			Path:    path,
			Line:    node.Line,
			Column:  node.Column,
			Message: "plugins item must not set multiple sources (name/remote/path/command)",
		})
	}
}

// pluginOptsValidator allows plugin opts values to be either scalar or sequence of scalars.
type pluginOptsValidator struct{}

func (pluginOptsValidator) Validate(node *yaml.Node, path string, ctx *v.ValidationContext) {
	if node.Kind != yaml.MappingNode {
		return
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valNode := node.Content[i+1]
		optsPath := path + "." + keyNode.Value

		switch valNode.Kind {
		case yaml.ScalarNode:
			continue
		case yaml.SequenceNode:
			for idx, item := range valNode.Content {
				if item.Kind != yaml.ScalarNode {
					ctx.AddError(v.ValidationError{
						Level:   v.LevelError,
						Path:    fmt.Sprintf("%s[%d]", optsPath, idx),
						Line:    item.Line,
						Column:  item.Column,
						Message: "opts array item must be a scalar value",
					})
				}
			}
		default:
			ctx.AddError(v.ValidationError{
				Level:   v.LevelError,
				Path:    optsPath,
				Line:    valNode.Line,
				Column:  valNode.Column,
				Message: "opts value must be a scalar or sequence of scalars",
				Expected: "scalar or sequence of scalars",
				Got:      yamlKindName(valNode.Kind),
			})
		}
	}
}

// managedDisableValidator validates generate.managed.disable entries.
type managedDisableValidator struct{}

func (managedDisableValidator) Validate(node *yaml.Node, path string, ctx *v.ValidationContext) {
	if node.Kind != yaml.MappingNode {
		return
	}
	has := func(k string) bool {
		for i := 0; i < len(node.Content); i += 2 {
			if node.Content[i].Value == k {
				return true
			}
		}
		return false
	}
	if !(has("module") || has("path") || has("file_option") || has("field_option") || has("field")) {
		ctx.AddError(v.ValidationError{
			Level:   v.LevelError,
			Path:    path,
			Line:    node.Line,
			Column:  node.Column,
			Message: "managed.disable entry must set at least one of module/path/file_option/field_option/field",
		})
	}
	if has("file_option") && has("field_option") {
		ctx.AddError(v.ValidationError{
			Level:   v.LevelError,
			Path:    path,
			Line:    node.Line,
			Column:  node.Column,
			Message: "managed.disable: file_option and field_option cannot both be set",
		})
	}
	if has("field") && !has("field_option") {
		ctx.AddError(v.ValidationError{
			Level:   v.LevelError,
			Path:    path,
			Line:    node.Line,
			Column:  node.Column,
			Message: "managed.disable.field requires field_option",
		})
	}
}

// managedOverrideValidator validates generate.managed.override entries.
type managedOverrideValidator struct{}

func (managedOverrideValidator) Validate(node *yaml.Node, path string, ctx *v.ValidationContext) {
	if node.Kind != yaml.MappingNode {
		return
	}
	has := func(k string) bool {
		for i := 0; i < len(node.Content); i += 2 {
			if node.Content[i].Value == k {
				return true
			}
		}
		return false
	}
	if !has("file_option") && !has("field_option") {
		ctx.AddError(v.ValidationError{
			Level:   v.LevelError,
			Path:    path,
			Line:    node.Line,
			Column:  node.Column,
			Message: "managed.override requires either file_option or field_option",
		})
	}
	if has("file_option") && has("field_option") {
		ctx.AddError(v.ValidationError{
			Level:   v.LevelError,
			Path:    path,
			Line:    node.Line,
			Column:  node.Column,
			Message: "managed.override cannot set both file_option and field_option",
		})
	}
	if has("field") && !has("field_option") {
		ctx.AddError(v.ValidationError{
			Level:   v.LevelError,
			Path:    path,
			Line:    node.Line,
			Column:  node.Column,
			Message: "managed.override.field can only be used with field_option",
		})
	}
}
