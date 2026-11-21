package protoc

import (
	"context"
	"strings"
	"testing"
)

// TestSemanticChecksMatchProtoc compiles one-file cases that protoc rejects
// for semantic reasons, plus a valid control, and expects protoc's message.
// An empty message only requires the file to be rejected.
func TestSemanticChecksMatchProtoc(t *testing.T) {
	cases := map[string]string{
		"custom_option_ok.proto":              "-",
		"custom_option_unknown.proto":         `Option "(nope)" unknown. Ensure that your proto definition file imports the proto which defines the option.`,
		"custom_option_wrong_type.proto":      `Value must be integer for int32 option "p.my".`,
		"duplicate_extension_number.proto":    `Extension number 50000 has already been used in "google.protobuf.MessageOptions" by extension "p.x".`,
		"enum_value_scope_conflict.proto":     "",
		"extension_number_out_of_range.proto": `"google.protobuf.MessageOptions" does not declare 5 as an extension number.`,
		"field_and_message_same_name.proto":   "",
		"json_name_conflict.proto":            `The default JSON name of field "aB" ("aB") conflicts with the default JSON name of field "a_b".`,
		"proto3_default.proto":                `Explicit default values are not allowed in proto3.`,
		"proto3_enum_first_nonzero.proto":     `The first enum value must be zero for open enums.`,
		"proto3_extension_range.proto":        `Extension ranges are not allowed in proto3.`,
		"proto3_group.proto":                  `Groups are not supported in proto3 syntax.`,
		"proto3_required.proto":               `Required fields are not allowed in proto3.`,
		"reserved_name_used.proto":            `Field name "a" is reserved.`,
		"reserved_number_used.proto":          `Field "a" uses reserved number 1.`,
		"unknown_field_option.proto":          `Option "custom_opt" unknown.`,
		"unknown_file_option.proto":           `Option "custom_opt" unknown.`,
		"unknown_message_option.proto":        `Option "nope" unknown.`,
		"wrong_option_type.proto":             `Value must be identifier for boolean option "google.protobuf.FieldOptions.deprecated".`,
	}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			var diags []string
			_, err := Compile(context.Background(), []string{name},
				WithImportPaths("../testdata/semantic"), WithStandardImports(true),
				WithErrorHandler(func(d Diagnostic) { diags = append(diags, d.String()) }))
			if want == "-" {
				if err != nil {
					t.Fatalf("valid file rejected: %v %v", err, diags)
				}
				return
			}
			if err == nil {
				t.Fatalf("accepted, protoc rejects it")
			}
			if want != "" && !strings.Contains(strings.Join(diags, "\n"), want) {
				t.Fatalf("diagnostics %q do not contain %q", diags, want)
			}
		})
	}
}
