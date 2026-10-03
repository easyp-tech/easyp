package schemagen

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	yamlvalidator "github.com/Yakwilik/go-yamlvalidator"
	"github.com/stretchr/testify/require"
)

func TestRunWritesSeparateV1SchemasAndAliases(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, Run(Options{OutDir: dir}))
	for _, tc := range []struct {
		name       string
		wantField  string
		legacyName string
	}{
		{name: "easyp", wantField: "breaking", legacyName: "breaking_check"},
		{name: "easyp.gen", wantField: "plugins", legacyName: "inputs"},
		{name: "protobuf.lock", wantField: "modules", legacyName: "direct"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			versioned, latest := schemaNames(tc.name)
			versionedRaw, err := os.ReadFile(filepath.Join(dir, versioned))
			require.NoError(t, err)
			latestRaw, err := os.ReadFile(filepath.Join(dir, latest))
			require.NoError(t, err)
			require.True(t, bytes.Equal(versionedRaw, latestRaw))
			var document struct {
				Schema     string                     `json:"$schema"`
				Properties map[string]json.RawMessage `json:"properties"`
			}
			require.NoError(t, json.Unmarshal(versionedRaw, &document))
			require.Equal(t, "https://json-schema.org/draft/2020-12/schema", document.Schema)
			require.Contains(t, document.Properties, tc.wantField)
			require.NotContains(t, document.Properties, tc.legacyName)
		})
	}
}

func TestGeneratedSchemasValidateConfigRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		schema   string
		contents string
		valid    bool
	}{
		{name: "policy settings", schema: "easyp", contents: "linters-settings:\n  UNKNOWN:\n    suffix: BAD\n"},
		{name: "plugin without identity", schema: "easyp.gen", contents: "plugins:\n  - out: gen\n"},
		{name: "plugin with two identities", schema: "easyp.gen", contents: "plugins:\n  - name: go\n    remote: example.com/go:v1.0.0\n    out: gen\n"},
		{name: "managed disable with selector", schema: "easyp.gen", contents: "generate:\n  managed:\n    disable:\n      - module: example.com/repo\n", valid: true},
		{name: "managed disable without selector", schema: "easyp.gen", contents: "generate:\n  managed:\n    disable:\n      - {}\n"},
		{name: "managed disable with conflicting options", schema: "easyp.gen", contents: "generate:\n  managed:\n    disable:\n      - file_option: go_package\n        field_option: json_name\n"},
		{name: "managed override without value", schema: "easyp.gen", contents: "generate:\n  managed:\n    override:\n      - file_option: go_package_prefix\n"},
		{name: "valid generator", schema: "easyp.gen", contents: "plugins:\n  - name: go\n    out: gen\n", valid: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			versioned, _ := schemaNames(tt.schema)
			dir := t.TempDir()
			require.NoError(t, Run(Options{OutDir: dir}))
			rawSchema, err := os.ReadFile(filepath.Join(dir, versioned))
			require.NoError(t, err)
			compiled, err := yamlvalidator.CompileJSONSchema(rawSchema)
			require.NoError(t, err)

			issues := yamlvalidator.NewValidator(compiled).ValidateBytes([]byte(tt.contents))

			require.Equal(t, tt.valid, !issues.HasErrors(), issues.FormatAll(false))
		})
	}
}
