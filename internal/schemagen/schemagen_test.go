package schemagen

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

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
