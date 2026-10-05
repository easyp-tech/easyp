package modules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestAddDirectV1RequirementPreservesFormatting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		original string
		version  string
		want     string
	}{
		{
			name:     "update preserves spacing and comment",
			original: "require\thttps://example.com/dep\t v1.0.0  // keep this\n",
			version:  "v1.1.0",
			want:     "require\thttps://example.com/dep\t v1.1.0  // keep this\n",
		},
		{
			name:     "versionless stays versionless",
			original: "require (\n\thttps://example.com/dep\t// keep this\n)\n",
			want:     "require (\n\thttps://example.com/dep\t// keep this\n)\n",
		},
		{
			name:     "pin versionless and retain comment",
			original: "require https://example.com/dep  // indirect needed by clients\n",
			version:  "v1.1.0",
			want:     "require https://example.com/dep v1.1.0  // needed by clients\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			original := []byte(tt.original)
			target := v1.Requirement{Module: "https://example.com/dep", Version: tt.version}

			updated, err := addDirectV1Requirement(original, target)

			require.NoError(t, err)
			assert.Equal(t, tt.want, string(updated))
		})
	}
}

func TestAddDirectV1RequirementAddsAndRejectsDuplicates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		original string
		version  string
		want     string
		wantErr  string
	}{
		{
			name:     "promote indirect requirement without changing other entries",
			original: "module example.com/app\nrequire ( // dependencies\n  example.com/dep v1.0.0 // indirect\n  example.com/other v1.0.0 // indirect\n)\n",
			version:  "v1.1.0",
			want:     "module example.com/app\nrequire ( // dependencies\n  example.com/dep v1.1.0\n  example.com/other v1.0.0 // indirect\n)\n",
		},
		{
			name:     "append versionless with missing final newline",
			original: "module example.com/root",
			want:     "module example.com/root\nrequire example.com/dep\n",
		},
		{
			name:     "append version without editing other directives",
			original: "module example.com/root\nroots proto\nrequire example.com/other v1.0.0 // keep\n",
			version:  "v1.1.0",
			want:     "module example.com/root\nroots proto\nrequire example.com/other v1.0.0 // keep\nrequire example.com/dep v1.1.0\n",
		},
		{
			name:     "duplicate in block and standalone requirement",
			original: "require (\n example.com/dep v1.0.0\n)\nrequire example.com/dep\n",
			wantErr:  "duplicate require example.com/dep",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			original := []byte(tt.original)

			updated, err := addDirectV1Requirement(original, v1.Requirement{Module: "example.com/dep", Version: tt.version})

			assert.Equal(t, tt.original, string(original))
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				assert.Nil(t, updated)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(updated))
		})
	}
}
