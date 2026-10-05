package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/config"
)

func TestGenerateSchemaPluginVersionParity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		raw       string
		wantError bool
	}{
		{
			name:      "local name rejects version structurally",
			raw:       "version: v1\nplugins:\n  - name: go\n    version: v1.2.3\n    out: gen\n",
			wantError: true,
		},
		{
			name:      "path rejects version structurally",
			raw:       "version: v1\nplugins:\n  - path: ./protoc-gen-x\n    version: v1.2.3\n    out: gen\n",
			wantError: true,
		},
		{
			name:      "command rejects version structurally",
			raw:       "version: v1\nplugins:\n  - command: [sh, ./plugin.sh]\n    version: v1.2.3\n    out: gen\n",
			wantError: true,
		},
		{
			name:      "remote requires version structurally",
			raw:       "version: v1\nplugins:\n  - remote: plugins.example.test/protobuf/go\n    out: gen\n",
			wantError: true,
		},
		{
			name: "remote with version is structurally valid",
			raw:  "version: v1\nplugins:\n  - remote: plugins.example.test/protobuf/go\n    version: v1.2.3\n    out: gen\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			issues := ValidateGenerateYAML([]byte(tt.raw))

			if tt.wantError {
				require.True(t, config.HasErrors(issues), "%+v", issues)
				return
			}
			assert.False(t, config.HasErrors(issues), "%+v", issues)
		})
	}
}
