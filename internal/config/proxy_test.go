package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadEnvironmentProxy(t *testing.T) {
	// Environment changes affect the whole process, so these cases cannot run in parallel.
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr string
	}{
		{name: "empty", value: "", want: ""},
		{name: "whitespace", value: "  ", want: ""},
		{name: "host", value: "gitlab.com", want: "gitlab.com"},
		{name: "trimmed_host", value: "  gitlab.com  ", want: "gitlab.com"},
		{name: "host_and_port", value: "gitlab.example.com:8443", want: "gitlab.example.com:8443"},
		{name: "scheme", value: "https://gitlab.com", wantErr: "EASYP_PROXY"},
		{name: "path", value: "gitlab.com/mirrors", wantErr: "EASYP_PROXY"},
		{name: "user", value: "git@gitlab.com", wantErr: "EASYP_PROXY"},
		{name: "query", value: "gitlab.com?x=1", wantErr: "EASYP_PROXY"},
		{name: "invalid_port", value: "gitlab.com:abc", wantErr: "EASYP_PROXY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(EnvProxy, tt.value)

			var cfg Config
			err := cfg.LoadEnvironment()

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, cfg.Proxy)
		})
	}
}
