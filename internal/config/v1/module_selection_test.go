package v1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParseGenerateModuleSelectors(t *testing.T) {
	t.Parallel()
	gen, err := ParseGenerate(strings.NewReader("generate:\n  modules:\n    - proto/user\n    - module: example.com/contracts\n      paths: [proto/api]\n      packages: [api.v1]\n  paths: [.]\n  packages: [api.v1]\n"))
	require.NoError(t, err)
	require.Len(t, gen.Generate.Modules, 2)
	raw, err := yaml.Marshal(gen.Generate.Modules[1])
	require.NoError(t, err)
	var module map[string]any
	require.NoError(t, yaml.Unmarshal(raw, &module))
	assert.Equal(t, "example.com/contracts", module["module"])
	assert.Equal(t, []any{"proto/api"}, module["paths"])
	assert.Equal(t, []any{"api.v1"}, module["packages"])
	assert.Equal(t, []string{"."}, gen.Generate.Paths)
	assert.Equal(t, []string{"api.v1"}, gen.Generate.Packages)
}

func TestParseGenerateRejectsInvalidModuleSelectors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, entry string }{
		{name: "empty_module", entry: "{module: ''}"},
		{name: "missing_module", entry: "{paths: [proto]}"},
		{name: "unknown_field", entry: "{module: proto, path: proto/api}"},
		{name: "invalid_path", entry: "{module: proto, paths: [../outside]}"},
		{name: "invalid_package", entry: "{module: proto, packages: [api/*]}"},
		{name: "numeric_module", entry: "42"},
		{name: "boolean_module", entry: "{module: true}"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseGenerate(strings.NewReader("generate:\n  modules: [" + tt.entry + "]\n"))
			require.Error(t, err)
		})
	}
}
