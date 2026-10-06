package migration

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestMigrationOmitsUnconfiguredGoPrefix(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		managed string
		enabled bool
	}{
		{name: "managed_omitted"},
		{name: "managed_enabled", managed: "  managed:\n    enabled: true\n    override: [{file_option: go_package_prefix, value: example.com/managed}]\n", enabled: true},
		{name: "managed_disabled", managed: "  managed:\n    enabled: false\n    override: [{file_option: go_package_prefix, value: example.com/managed}]\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeFixture(t, root, "easyp.yaml", "generate:\n  inputs: [{directory: .}]\n  plugins: [{name: go, out: gen}]\n"+tt.managed)
			writeFixture(t, root, "item.proto", "syntax = \"proto3\"; package item.v1; option go_package = \"example.com/original/item\"; message Item {}")

			plan, err := Build(t.Context(), Options{Dir: root, Module: "example.com/app"})
			require.NoError(t, err)
			raw := outputContent(t, plan, v1.GenerateFile)
			var document map[string]any
			require.NoError(t, yaml.Unmarshal(raw, &document))
			assert.NotContains(t, document, "options", "migration must not manufacture new Go prefix settings")
			gen, err := v1.ParseGenerate(bytes.NewReader(raw))
			require.NoError(t, err)
			assert.Nil(t, gen.Options.Go.PackagePrefix, "an omitted prefix must remain optional")
			assert.Equal(t, tt.enabled, gen.Generate.Managed.Enabled)
			if tt.managed != "" {
				require.Len(t, gen.Generate.Managed.Override, 1)
				assert.Equal(t, "go_package_prefix", gen.Generate.Managed.Override[0].FileOption)
				assert.Equal(t, "example.com/managed", gen.Generate.Managed.Override[0].Value)
			}
		})
	}
}
