package migration

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestConvertGenerateReportsAllUnpinnedRemotePlugins(t *testing.T) {
	t.Parallel()
	var cfg legacyConfig
	require.NoError(t, yaml.Unmarshal([]byte("generate:\n  plugins:\n    - {remote: registry/go, out: gen}\n    - {remote: registry/go-grpc:latest, out: gen}\n    - {remote: registry/valid:v1.2.3, out: gen}\n    - {remote: registry/validate, out: gen}\n"), &cfg))
	raw, err := convertGenerate(cfg, nil, nil, nil)
	require.Error(t, err)
	require.Nil(t, raw, "invalid plugins must not produce a partial candidate")
	for _, want := range []string{"generate.plugins[0]", "registry/go", "generate.plugins[1]", "registry/go-grpc:latest", "generate.plugins[3]", "registry/validate"} {
		require.ErrorContains(t, err, want)
	}
	require.NotContains(t, err.Error(), "generate.plugins[2]")
}
