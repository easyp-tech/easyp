package api

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestV1RequirementEditsRespectModuleBlocks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		original    string
		wantDirect  string
		wantUpdated string
	}{
		{
			name:        "versionless_looking_roots_are_not_requirements",
			original:    "module example.com/app\nroots ( // source directories\n  require example.com/dep\n)\n",
			wantDirect:  "module example.com/app\nroots ( // source directories\n  require example.com/dep\n)\nrequire example.com/dep v1.1.0\n",
			wantUpdated: "module example.com/app\nroots ( // source directories\n  require example.com/dep\n)\n",
		},
		{
			name:        "version_looking_root_is_not_rewritten",
			original:    "module example.com/app\nroots(\n  require example.com/dep v1.0.0 // directory names\n) // end roots\n",
			wantDirect:  "module example.com/app\nroots(\n  require example.com/dep v1.0.0 // directory names\n) // end roots\nrequire example.com/dep v1.1.0\n",
			wantUpdated: "module example.com/app\nroots(\n  require example.com/dep v1.0.0 // directory names\n) // end roots\n",
		},
		{
			name:        "roots_do_not_duplicate_standalone_requirement",
			original:    "module example.com/app\nroots (\n  require example.com/dep v1.0.0\n)\nrequire example.com/dep v1.0.0 // indirect\n",
			wantDirect:  "module example.com/app\nroots (\n  require example.com/dep v1.0.0\n)\nrequire example.com/dep v1.1.0\n",
			wantUpdated: "module example.com/app\nroots (\n  require example.com/dep v1.0.0\n)\nrequire example.com/dep v1.1.0 // indirect\n",
		},
		{
			name:        "require_block_after_replace_with_comments",
			original:    "module example.com/app\nreplace ( // local development\n  example.com/dep => ./local\n)\nrequire( // dependencies\n  example.com/dep v1.0.0 // indirect\n) // end dependencies\n",
			wantDirect:  "module example.com/app\nreplace ( // local development\n  example.com/dep => ./local\n)\nrequire( // dependencies\n  example.com/dep v1.1.0\n) // end dependencies\n",
			wantUpdated: "module example.com/app\nreplace ( // local development\n  example.com/dep => ./local\n)\nrequire( // dependencies\n  example.com/dep v1.1.0 // indirect\n) // end dependencies\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := []byte(tt.original)
			module, err := v1.ParseModule(bytes.NewReader(original))
			require.NoError(t, err)
			target := v1.Requirement{Module: "example.com/dep", Version: "v1.1.0"}

			direct, err := addDirectV1Requirement(original, target)
			updated := rewriteV1RequiredVersions(original, map[string]string{target.Module: target.Version})

			require.NoError(t, err)
			assert.Equal(t, tt.wantDirect, string(direct))
			assert.Equal(t, tt.wantUpdated, string(updated))
			assert.Equal(t, tt.original, string(original))
			for _, result := range [][]byte{direct, updated} {
				parsed, err := v1.ParseModule(bytes.NewReader(result))
				require.NoError(t, err)
				assert.Equal(t, module.Roots, parsed.Roots)
				assert.Equal(t, module.Replaces, parsed.Replaces)
			}
		})
	}
}
