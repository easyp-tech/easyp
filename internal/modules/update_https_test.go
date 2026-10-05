package modules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRewriteV1RequiredVersionsHTTPSIdentity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		original string
		updates  map[string]string
		want     string
	}{
		{
			name:     "preserve URL and indirect comment",
			original: "module https://example.com/user.git\nrequire https://example.com/common.git v1.0.0 // indirect\n",
			updates:  map[string]string{"https://example.com/common.git": "v1.1.0"},
			want:     "module https://example.com/user.git\nrequire https://example.com/common.git v1.1.0 // indirect\n",
		},
		{
			name:     "leave versionless dependency untouched",
			original: "module https://example.com/user.git\nrequire https://example.com/common.git // latest head\n",
			updates:  map[string]string{"https://example.com/common.git": "v1.1.0"},
			want:     "module https://example.com/user.git\nrequire https://example.com/common.git // latest head\n",
		},
		{
			name:     "leave unselected dependency untouched",
			original: "module https://example.com/user.git\nrequire https://example.com/common.git v1.0.0 // indirect\n",
			updates:  map[string]string{"https://example.com/other.git": "v1.1.0"},
			want:     "module https://example.com/user.git\nrequire https://example.com/common.git v1.0.0 // indirect\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			original := []byte(tt.original)

			updated := rewriteV1RequiredVersions(original, tt.updates)

			assert.Equal(t, tt.want, string(updated))
			assert.Equal(t, tt.original, string(original))
		})
	}
}
