package v1

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
)

func TestManagedValidationRejectsInvalidRules(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		rule string
		path string
	}{
		{name: "unknown_file_override", rule: "override: [{file_option: bogus_option, value: example.com/gen}]", path: "override[0].file_option"},
		{name: "unknown_field_override", rule: "override: [{field_option: json_name, value: name}]", path: "override[0].field_option"},
		{name: "unknown_file_disable", rule: "disable: [{file_option: bogus_option}]", path: "disable[0].file_option"},
		{name: "unknown_field_disable", rule: "disable: [{field_option: json_name}]", path: "disable[0].field_option"},
		{name: "boolean_as_string", rule: "override: [{file_option: java_multiple_files, value: 'true'}]", path: "override[0].value"},
		{name: "utf8_as_string", rule: "override: [{file_option: java_string_check_utf8, value: 'false'}]", path: "override[0].value"},
		{name: "arenas_as_number", rule: "override: [{file_option: cc_enable_arenas, value: 1}]", path: "override[0].value"},
		{name: "string_as_boolean", rule: "override: [{file_option: go_package_prefix, value: false}]", path: "override[0].value"},
		{name: "string_as_number", rule: "override: [{file_option: java_package, value: 17}]", path: "override[0].value"},
		{name: "string_as_map", rule: "override: [{file_option: ruby_package, value: {suffix: Foo}}]", path: "override[0].value"},
		{name: "string_as_list", rule: "override: [{file_option: swift_prefix, value: [Foo]}]", path: "override[0].value"},
		{name: "invalid_optimize_for", rule: "override: [{file_option: optimize_for, value: FAST}]", path: "override[0].value"},
		{name: "lowercase_optimize_for", rule: "override: [{file_option: optimize_for, value: speed}]", path: "override[0].value"},
		{name: "numeric_optimize_for", rule: "override: [{file_option: optimize_for, value: 1}]", path: "override[0].value"},
		{name: "invalid_jstype", rule: "override: [{field_option: jstype, value: STRING}]", path: "override[0].value"},
		{name: "boolean_jstype", rule: "override: [{field_option: jstype, value: true}]", path: "override[0].value"},
		{name: "null_value", rule: "override: [{file_option: go_package, value: null}]", path: "override[0].value"},
		{name: "missing_value", rule: "override: [{file_option: go_package}]", path: "override[0].value"},
		{name: "both_override_options", rule: "override: [{file_option: go_package, field_option: jstype, value: JS_STRING}]", path: "override[0]"},
		{name: "neither_override_option", rule: "override: [{module: example.com/api, value: value}]", path: "override[0]"},
		{name: "both_disable_options", rule: "disable: [{file_option: go_package, field_option: jstype}]", path: "disable[0]"},
		{name: "empty_disable", rule: "disable: [{}]", path: "disable[0]"},
		{name: "blank_disable_selector", rule: "disable: [{module: ''}]", path: "disable[0]"},
		{name: "field_without_option_disable", rule: "disable: [{field: acme.v1.Message.id}]", path: "disable[0]"},
		{name: "file_option_with_field_disable", rule: "disable: [{file_option: go_package, field: acme.v1.Message.id}]", path: "disable[0]"},
		{name: "file_option_with_field_override", rule: "override: [{file_option: go_package, field: acme.v1.Message.id, value: gen}]", path: "override[0]"},
		{name: "selector_wrong_type", rule: "override: [{file_option: go_package, module: true, value: gen}]", path: "override[0].module"},
		{name: "selector_unknown_key", rule: "disable: [{file_option: go_package, modules: [example.com/api]}]", path: "disable[0].modules"},
	}
	for _, tt := range tests {
		for _, enabled := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/enabled_%t", tt.name, enabled), func(t *testing.T) {
				t.Parallel()
				raw := fmt.Sprintf("generate:\n  managed:\n    enabled: %t\n    %s\nplugins:\n  - command: [must-not-run]\n    out: gen\n", enabled, tt.rule)
				issues := ValidateGenerateYAML([]byte(raw))
				require.True(t, config.HasErrors(issues), "%+v", issues)
				assert.Contains(t, issues[0].Message, "generate.managed."+tt.path)
				assert.Equal(t, 4, issues[0].Line)
				assert.Positive(t, issues[0].Column)

				got, err := ParseGenerate(strings.NewReader(raw))
				require.Error(t, err)
				assert.Equal(t, Generate{}, got, "invalid configuration must not escape to generation")
				path := filepath.Join(t.TempDir(), GenerateFile)
				require.NoError(t, os.WriteFile(path, []byte(raw), 0o600))
				fileIssues, err := ValidateFile(path)
				require.NoError(t, err)
				assert.Equal(t, issues, fileIssues)
			})
		}
	}
}

func TestManagedValidationAcceptsSupportedRules(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		field string
		value string
	}{
		{name: "go_package", value: "'example.com/api;apiv1'"},
		{name: "go_package_prefix", value: "'example.com/{{file_path_without:proto/}}'"},
		{name: "go_package_prefix", value: "'example.com/{{file_path}}'"},
		{name: "go_package_prefix", value: "'example.com/{{file_path_spec}}'"},
		{name: "go_package_prefix", value: "'example.com/{{file_dir}}'"},
		{name: "go_package_prefix", value: "'example.com/{{file_dir_without:proto/}}'"},
		{name: "java_package", value: "com.acme"},
		{name: "java_package_prefix", value: "com"},
		{name: "java_package_suffix", value: "api"},
		{name: "java_multiple_files", value: "false"},
		{name: "java_outer_classname", value: "API"},
		{name: "java_string_check_utf8", value: "true"},
		{name: "csharp_namespace", value: "Acme.API"},
		{name: "csharp_namespace_prefix", value: "Acme"},
		{name: "ruby_package", value: "Acme::API"},
		{name: "ruby_package_suffix", value: "API"},
		{name: "php_namespace", value: "'Acme\\API'"},
		{name: "php_metadata_namespace", value: "Metadata"},
		{name: "php_metadata_namespace_suffix", value: "Metadata"},
		{name: "objc_class_prefix", value: "ACM"},
		{name: "swift_prefix", value: "''"},
		{name: "cc_enable_arenas", value: "false"},
		{name: "optimize_for", value: "SPEED"},
		{name: "optimize_for", value: "CODE_SIZE"},
		{name: "optimize_for", value: "LITE_RUNTIME"},
		{name: "jstype", field: "field_option", value: "JS_NORMAL"},
		{name: "jstype", field: "field_option", value: "JS_STRING"},
		{name: "jstype", field: "field_option", value: "JS_NUMBER"},
	}
	for _, tt := range tests {
		t.Run(tt.name+"/"+tt.value, func(t *testing.T) {
			t.Parallel()
			field := tt.field
			if field == "" {
				field = "file_option"
			}
			raw := fmt.Sprintf("generate:\n  managed:\n    enabled: false\n    disable:\n      - %s: %s\n      - module: example.com/api\n      - package: acme.v1\n      - path: proto/\n      - field_option: jstype\n        field: acme.v1.Message.id\n    override:\n      - %s: %s\n        value: %s\n        module: example.com/api\n        package: acme.v1\n        path: proto/\n", field, tt.name, field, tt.name, tt.value)
			assert.Empty(t, ValidateGenerateYAML([]byte(raw)))
			got, err := ParseGenerate(strings.NewReader(raw))
			require.NoError(t, err)
			assert.Nil(t, got.Options.Go.PackagePrefix, "omitted Go prefix must still allow inheritance")
			assert.False(t, got.Generate.Managed.Enabled)
			require.Len(t, got.Generate.Managed.Override, 1)
			assert.Equal(t, "example.com/api", got.Generate.Managed.Override[0].Module)
		})
	}
}

func TestManagedValidationPreservesSelectorAndInheritanceValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  string
	}{
		{name: "no_selectors", raw: "generate:\n  managed:\n    override: [{file_option: go_package, value: gen}]\n"},
		{name: "empty_optional_selectors", raw: "generate:\n  managed:\n    disable: [{file_option: go_package, field: '', module: ''}]\n    override: [{file_option: go_package, value: '', field: '', path: ''}]\n"},
		{name: "field_with_context", raw: "generate:\n  managed:\n    override: [{field_option: jstype, value: JS_STRING, field: acme.v1.Message.id, module: example.com/api, package: acme.v1, path: proto/api.proto}]\n"},
		{name: "selectors_retain_existing_matching", raw: "generate:\n  managed:\n    override: [{file_option: go_package, value: gen, module: 'some/module', package: 'acme.*', path: 'api/**'}]\n"},
		{name: "empty_override_list", raw: "generate:\n  managed:\n    override: []\n"},
		{name: "explicit_empty_prefix", raw: "options:\n  go:\n    package_prefix: ''\n"},
		{name: "explicit_prefix", raw: "options:\n  go:\n    package_prefix: example.com/gen\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Empty(t, ValidateGenerateYAML([]byte(tt.raw)))
			got, err := ParseGenerate(strings.NewReader(tt.raw))
			require.NoError(t, err)
			var want Generate
			require.NoError(t, yaml.Unmarshal([]byte(tt.raw), &want))
			assert.Equal(t, want.Generate, got.Generate)
			assert.Equal(t, want.Options, got.Options)
		})
	}
}

func TestManagedValidationRejectsBeforeDescriptorApplication(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, rule string }{
		{name: "unknown_option", rule: "file_option: bogus_option, value: ignored"},
		{name: "wrong_value_type", rule: "file_option: java_multiple_files, value: 'false'"},
		{name: "invalid_enum", rule: "field_option: jstype, value: INVALID"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fd := &descriptorpb.FileDescriptorProto{Name: proto.String("api.proto"), Package: proto.String("acme.v1")}
			before := proto.Clone(fd)
			raw := "generate:\n  managed:\n    enabled: true\n    override: [{" + tt.rule + "}]\n"
			got, err := ParseGenerate(strings.NewReader(raw))
			if err == nil {
				// Follow the runtime boundary: application is only reachable after parsing succeeds.
				managed := core.ManagedModeConfig{Enabled: got.Generate.Managed.Enabled}
				for _, rule := range got.Generate.Managed.Override {
					managed.Override = append(managed.Override, core.ManagedOverrideRule{FileOption: core.FileOptionType(rule.FileOption), FieldOption: core.FieldOptionType(rule.FieldOption), Value: rule.Value})
				}
				require.NoError(t, core.ApplyManagedMode([]*descriptorpb.FileDescriptorProto{fd}, managed, nil))
			}
			require.Error(t, err)
			assert.True(t, proto.Equal(before, fd), "invalid rules must be rejected before managed defaults mutate descriptors")
		})
	}
}
