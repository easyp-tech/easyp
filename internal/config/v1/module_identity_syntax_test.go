package v1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseModuleRejectsEmbeddedVersionsInIdentities(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "module",
			raw:  "module example.test/root@v1.0.0\n",
			want: "module module identity \"example.test/root@v1.0.0\" must not embed version \"v1.0.0\"",
		},
		{
			name: "require",
			raw:  "module example.test/root\nrequire example.test/dep@v1.0.0\n",
			want: "use require example.test/dep v1.0.0",
		},
		{
			name: "replace",
			raw:  "module example.test/root\nrequire example.test/dep v1.0.0\nreplace example.test/dep@v1.0.0 => ./dep\n",
			want: "use replace example.test/dep => <local-path>",
		},
		{
			name: "empty suffix",
			raw:  "module example.test/root\nrequire example.test/dep@\n",
			want: "must not end with @",
		},
		{
			name: "URL path tag",
			raw:  "module example.test/root\nrequire ssh://git@example.test/acme/dep@v1.0.0\n",
			want: "use require ssh://git@example.test/acme/dep v1.0.0",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseModule(strings.NewReader(tt.raw))

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestParseModuleAllowsAtInURLAuthority(t *testing.T) {
	t.Parallel()

	module, err := ParseModule(strings.NewReader("module example.test/root\nrequire ssh://git@example.test/acme/dep v1.0.0\nreplace ssh://git@example.test/acme/dep => ./dep\n"))

	require.NoError(t, err)
	require.Len(t, module.Requires, 1)
	assert.Equal(t, "ssh://git@example.test/acme/dep", module.Requires[0].Module)
	require.Len(t, module.Replaces, 1)
	assert.Equal(t, "ssh://git@example.test/acme/dep", module.Replaces[0].Module)
}
