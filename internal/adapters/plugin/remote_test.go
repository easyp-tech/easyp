package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemotePluginTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		source  string
		target  string
		name    string
		version string
	}{
		{"localhost:8080/grpc/go:v1.6.2", "passthrough:///localhost:8080", "grpc/go", "v1.6.2"},
		{"http://127.0.0.1:8080/protocolbuffers/go", "passthrough:///127.0.0.1:8080", "protocolbuffers/go", "latest"},
		{"https://plugins.example.com/grpc/go:v1.6.2", "passthrough:///plugins.example.com:443", "grpc/go", "v1.6.2"},
		{"enterprise.plugins.beta.easyp.tech:443/bufbuild/es:v2.12.0", "passthrough:///enterprise.plugins.beta.easyp.tech:443", "bufbuild/es", "v2.12.0"},
	}

	e := &RemotePluginExecutor{}

	for _, tt := range tests {
		host, name, version, err := e.parsePluginURL(tt.source)
		require.NoError(t, err, tt.source)

		assert.Equal(t, tt.target, dialTarget(host), tt.source)
		assert.Equal(t, tt.name, name, tt.source)
		assert.Equal(t, tt.version, version, tt.source)
	}
}
