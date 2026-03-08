package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseConfig_ManagedModePackageSelectors(t *testing.T) {
	content := `lint:
  use:
    - DIRECTORY_SAME_PACKAGE
generate:
  inputs:
    - directory: proto
  plugins:
    - name: go
      out: .
  managed:
    enabled: true
    disable:
      - package: acme.weather.v1
        file_option: java_package_prefix
    override:
      - file_option: go_package_prefix
        package: acme.weather.v1
        value: github.com/acme/gen/go
`

	cfg, err := ParseConfig([]byte(content))
	require.NoError(t, err)

	require.Len(t, cfg.Generate.Managed.Disable, 1)
	require.Equal(t, "acme.weather.v1", cfg.Generate.Managed.Disable[0].Package)
	require.Equal(t, "java_package_prefix", cfg.Generate.Managed.Disable[0].FileOption)

	require.Len(t, cfg.Generate.Managed.Override, 1)
	require.Equal(t, "acme.weather.v1", cfg.Generate.Managed.Override[0].Package)
	require.Equal(t, "go_package_prefix", cfg.Generate.Managed.Override[0].FileOption)
	require.Equal(t, "github.com/acme/gen/go", cfg.Generate.Managed.Override[0].Value)
}
