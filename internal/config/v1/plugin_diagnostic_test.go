package v1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmbeddedRemoteVersionDiagnostic(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, remote, version string }{
		{name: "legacy URI", remote: "registry.example.test/go:v1.2.3"},
		{name: "legacy URI with port", remote: "localhost:8080/go:v1.2.3"},
		{name: "conflicting versions", remote: "registry.example.test/go:v1.2.3", version: "v2.0.0"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := (Plugin{Remote: tt.remote, Version: tt.version, Out: "gen"}).Validate()
			require.ErrorContains(t, err, "specify the remote plugin version only in version")
			require.ErrorContains(t, err, `version: "v1.2.3"`)
		})
	}
}
