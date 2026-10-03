package core

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
)

func TestCompareBreakingProfiles_FieldAndWireMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		categories []string
		baseline   string
		current    string
		wantRules  []string
	}{
		{
			name:       "implicit_enum_default_changes",
			baseline:   `syntax = "proto2"; package sample.v1; enum State { ZERO = 0; ONE = 1; } message Item { optional State state = 1; }`,
			current:    `syntax = "proto2"; package sample.v1; enum State { ONE = 1; ZERO = 0; } message Item { optional State state = 1; }`,
			categories: []string{"WIRE"}, wantRules: []string{"FIELD_SAME_DEFAULT"},
		},
		{
			name:       "implicit_enum_default_unchanged",
			baseline:   `syntax = "proto2"; package sample.v1; enum State { ZERO = 0; ONE = 1; } message Item { optional State state = 1; }`,
			current:    `syntax = "proto2"; package sample.v1; enum State { ZERO = 0; ONE = 1; TWO = 2; } message Item { optional State state = 1; }`,
			categories: []string{"WIRE"},
		},
		{
			name:       "field_rename_is_source_and_json_breaking",
			baseline:   `syntax = "proto3"; package sample.v1; message Item { string before = 1; }`,
			current:    `syntax = "proto3"; package sample.v1; message Item { string after = 1; }`,
			categories: []string{"FILE"}, wantRules: []string{"FIELD_SAME_NAME"},
		},
		{
			name:       "field_rename_is_json_breaking",
			baseline:   `syntax = "proto3"; package sample.v1; message Item { string before = 1; }`,
			current:    `syntax = "proto3"; package sample.v1; message Item { string after = 1; }`,
			categories: []string{"WIRE_JSON"}, wantRules: []string{"FIELD_SAME_NAME"},
		},
		{
			name:       "field_rename_is_not_binary_breaking",
			baseline:   `syntax = "proto3"; package sample.v1; message Item { string before = 1; }`,
			current:    `syntax = "proto3"; package sample.v1; message Item { string after = 1; }`,
			categories: []string{"WIRE"},
		},
		{
			name:       "json_name_change_is_json_breaking",
			baseline:   `syntax = "proto3"; package sample.v1; message Item { string value = 1 [json_name = "before"]; }`,
			current:    `syntax = "proto3"; package sample.v1; message Item { string value = 1 [json_name = "after"]; }`,
			categories: []string{"WIRE_JSON"}, wantRules: []string{"FIELD_SAME_JSON_NAME"},
		},
		{
			name:       "int32_uint32_is_compatible_in_both_wire_profiles",
			baseline:   `syntax = "proto3"; package sample.v1; message Item { int32 value = 1; }`,
			current:    `syntax = "proto3"; package sample.v1; message Item { uint32 value = 1; }`,
			categories: []string{"WIRE"},
		},
		{
			name:       "int32_sint32_is_wire_incompatible",
			baseline:   `syntax = "proto3"; package sample.v1; message Item { int32 value = 1; }`,
			current:    `syntax = "proto3"; package sample.v1; message Item { sint32 value = 1; }`,
			categories: []string{"WIRE"}, wantRules: []string{"FIELD_WIRE_COMPATIBLE_TYPE"},
		},
		{
			name:       "int32_uint32_is_json_compatible",
			baseline:   `syntax = "proto3"; package sample.v1; message Item { int32 value = 1; }`,
			current:    `syntax = "proto3"; package sample.v1; message Item { uint32 value = 1; }`,
			categories: []string{"WIRE_JSON"},
		},
		{
			name:       "int32_sint32_is_json_incompatible",
			baseline:   `syntax = "proto3"; package sample.v1; message Item { int32 value = 1; }`,
			current:    `syntax = "proto3"; package sample.v1; message Item { sint32 value = 1; }`,
			categories: []string{"WIRE_JSON"}, wantRules: []string{"FIELD_WIRE_JSON_COMPATIBLE_TYPE"},
		},
		{
			name:       "string_to_bytes_is_binary_compatible",
			baseline:   `syntax = "proto3"; package sample.v1; message Item { string value = 1; }`,
			current:    `syntax = "proto3"; package sample.v1; message Item { bytes value = 1; }`,
			categories: []string{"WIRE"},
		},
		{
			name:       "bytes_to_string_is_binary_incompatible",
			baseline:   `syntax = "proto3"; package sample.v1; message Item { bytes value = 1; }`,
			current:    `syntax = "proto3"; package sample.v1; message Item { string value = 1; }`,
			categories: []string{"WIRE"}, wantRules: []string{"FIELD_WIRE_COMPATIBLE_TYPE"},
		},
		{
			name:       "implicit_explicit_optional_is_source_only",
			baseline:   `syntax = "proto3"; package sample.v1; message Item { string value = 1; }`,
			current:    `syntax = "proto3"; package sample.v1; message Item { optional string value = 1; }`,
			categories: []string{"PACKAGE"}, wantRules: []string{"FIELD_SAME_CARDINALITY"},
		},
		{
			name:       "implicit_explicit_optional_is_wire_compatible",
			baseline:   `syntax = "proto3"; package sample.v1; message Item { string value = 1; }`,
			current:    `syntax = "proto3"; package sample.v1; message Item { optional string value = 1; }`,
			categories: []string{"WIRE"},
		},
		{
			name:       "real_oneof_membership_is_binary_breaking",
			baseline:   `syntax = "proto3"; package sample.v1; message Item { string value = 1; }`,
			current:    `syntax = "proto3"; package sample.v1; message Item { oneof choice { string value = 1; } }`,
			categories: []string{"WIRE"}, wantRules: []string{"FIELD_SAME_ONEOF"},
		},
		{
			name:     "deleting_field_requires_number_reservation_for_wire",
			baseline: `syntax = "proto3"; package sample.v1; message Item { string value = 1; }`,
			current:  `syntax = "proto3"; package sample.v1; message Item {}`, categories: []string{"WIRE"},
			wantRules: []string{"FIELD_NO_DELETE_UNLESS_NUMBER_RESERVED"},
		},
		{
			name:       "wire_json_also_requires_deleted_field_name_reservation",
			baseline:   `syntax = "proto3"; package sample.v1; message Item { string value = 1; }`,
			current:    `syntax = "proto3"; package sample.v1; message Item { reserved 1; }`,
			categories: []string{"WIRE_JSON"}, wantRules: []string{"FIELD_NO_DELETE_UNLESS_NAME_RESERVED"},
		},
		{
			name:       "wire_json_accepts_deleted_field_when_name_and_number_are_reserved",
			baseline:   `syntax = "proto3"; package sample.v1; message Item { string value = 1; }`,
			current:    `syntax = "proto3"; package sample.v1; message Item { reserved 1; reserved "value"; }`,
			categories: []string{"WIRE_JSON"},
		},
		{
			name:       "wire_accepts_repeated_to_map_with_matching_value_type",
			baseline:   `syntax = "proto3"; package sample.v1; message Entry { int32 key = 1; int32 value = 2; } message Item { repeated Entry value = 1; }`,
			current:    `syntax = "proto3"; package sample.v1; message Item { map<int32, int32> value = 1; }`,
			categories: []string{"WIRE"},
		},
		{
			name:       "wire_json_rejects_repeated_to_map",
			baseline:   `syntax = "proto3"; package sample.v1; message Entry { int32 key = 1; int32 value = 2; } message Item { repeated Entry value = 1; }`,
			current:    `syntax = "proto3"; package sample.v1; message Item { map<int32, int32> value = 1; }`,
			categories: []string{"WIRE_JSON"}, wantRules: []string{"FIELD_WIRE_JSON_COMPATIBLE_CARDINALITY"},
		},
		{
			name:       "default_change_breaks_wire",
			baseline:   `syntax = "proto2"; package sample.v1; message Item { optional int32 value = 1 [default = 2]; }`,
			current:    `syntax = "proto2"; package sample.v1; message Item { optional int32 value = 1 [default = 3]; }`,
			categories: []string{"WIRE"}, wantRules: []string{"FIELD_SAME_DEFAULT"},
		},
		{
			name:       "new_required_field_breaks_wire",
			baseline:   `syntax = "proto2"; package sample.v1; message Item { optional int32 value = 1; }`,
			current:    `syntax = "proto2"; package sample.v1; message Item { optional int32 value = 1; required int32 added = 2; }`,
			categories: []string{"WIRE"}, wantRules: []string{"MESSAGE_SAME_REQUIRED_FIELDS"},
		},
		{
			name:       "package_change_is_binary_breaking",
			baseline:   `syntax = "proto3"; package sample.v1; message Item { string value = 1; }`,
			current:    `syntax = "proto3"; package sample.v2; message Item { string value = 1; }`,
			categories: []string{"WIRE"}, wantRules: []string{"FILE_SAME_PACKAGE"},
		},
		{
			name:       "moving_message_within_package_passes_package_profile",
			baseline:   `syntax = "proto3"; package sample.v1; message Item { string value = 1; }`,
			current:    `syntax = "proto3"; package sample.v1; message Item { string value = 1; }`,
			categories: []string{"PACKAGE"},
		},
		{
			name:       "file_option_change_is_source_breaking",
			baseline:   `syntax = "proto3"; package sample.v1; option go_package = "example.com/old"; message Item { string value = 1; }`,
			current:    `syntax = "proto3"; package sample.v1; option go_package = "example.com/new"; message Item { string value = 1; }`,
			categories: []string{"FILE"}, wantRules: []string{"FILE_SAME_GO_PACKAGE"},
		},
		{
			name:       "mixed_profiles_report_each_issue_once",
			baseline:   `syntax = "proto3"; package sample.v1; message Item { int32 before = 1; }`,
			current:    `syntax = "proto3"; package sample.v1; message Item { uint32 after = 1; }`,
			categories: []string{"WIRE", "WIRE_JSON", "PACKAGE", "FILE"},
			wantRules:  []string{"FIELD_SAME_NAME", "FIELD_SAME_TYPE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			baseline := writeProfileProto(t, "baseline", tt.baseline)
			current := writeProfileProto(t, "current", tt.current)
			app := New(Options{Logger: logger.NewNop(), BreakingCheckConfig: BreakingCheckConfig{Categories: tt.categories}})

			issues, err := app.CompareBreaking(t.Context(), fs.NewFSWalker(current, "."), fs.NewFSWalker(baseline, "."), nil)

			require.NoError(t, err)
			gotRules := make([]string, len(issues))
			for i, issue := range issues {
				gotRules[i] = issue.RuleName
				require.NotEmpty(t, issue.Path)
				require.Positive(t, issue.Position.Line)
			}
			require.ElementsMatch(t, tt.wantRules, gotRules, "%+v", issues)
			require.Len(t, gotRules, len(slices.Compact(slices.Clone(gotRules))), "duplicate diagnostics: %+v", issues)
		})
	}
}

func writeProfileProto(t *testing.T, name, contents string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.MkdirAll(root, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "schema.proto"), []byte(contents), 0o600))
	return root
}
