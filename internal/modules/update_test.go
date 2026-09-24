package modules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestRewriteV1RequiredVersionsBlockWhitespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opening string
		want    string
	}{
		{
			name:    "single space",
			opening: "require (",
			want:    "module example.com/root\nrequire ( // dependencies\n\thttps://example.com/common.git v1.1.0 // indirect\n)\n",
		},
		{
			name:    "no space",
			opening: "require(",
			want:    "module example.com/root\nrequire( // dependencies\n\thttps://example.com/common.git v1.1.0 // indirect\n)\n",
		},
		{
			name:    "tab",
			opening: "require\t(",
			want:    "module example.com/root\nrequire\t( // dependencies\n\thttps://example.com/common.git v1.1.0 // indirect\n)\n",
		},
		{
			name:    "multiple spaces",
			opening: "require   (",
			want:    "module example.com/root\nrequire   ( // dependencies\n\thttps://example.com/common.git v1.1.0 // indirect\n)\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := "module example.com/root\n" + tt.opening + " // dependencies\n\thttps://example.com/common.git v1.0.0 // indirect\n)\n"
			_, err := v1.ParseModule(strings.NewReader(original))
			require.NoError(t, err)

			updated := rewriteV1RequiredVersions([]byte(original), map[string]string{
				"https://example.com/common.git": "v1.1.0",
			})

			assert.Equal(t, tt.want, string(updated))
		})
	}
}
