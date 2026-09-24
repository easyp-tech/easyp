package v1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPluginValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		plugin          Plugin
		expectedMessage string
	}{
		{
			name:   "bundled plugin",
			plugin: Plugin{Name: "go", Out: "gen"},
		},
		{
			name:   "pinned remote plugin with server port",
			plugin: Plugin{Remote: "localhost:8080/go", Version: "v1.2.3", Out: "gen"},
		},
		{
			name:   "binary path",
			plugin: Plugin{Path: "./tools/protoc-gen-custom", Out: "gen"},
		},
		{
			name:   "custom command",
			plugin: Plugin{Command: []string{"sh", "./tools/run-plugin"}, Out: "gen"},
		},
		{
			name:            "no source",
			plugin:          Plugin{Out: "gen"},
			expectedMessage: "exactly one of name, path, command or remote is required",
		},
		{
			name:            "both sources",
			plugin:          Plugin{Name: "go", Remote: "example.com/go", Out: "gen"},
			expectedMessage: "exactly one of name, path, command or remote is required",
		},
		{
			name:            "name and path",
			plugin:          Plugin{Name: "go", Path: "./tools/protoc-gen-custom", Out: "gen"},
			expectedMessage: "exactly one of name, path, command or remote is required",
		},
		{
			name:            "empty command",
			plugin:          Plugin{Command: []string{}, Out: "gen"},
			expectedMessage: "command executable is required",
		},
		{
			name:            "name and empty command",
			plugin:          Plugin{Name: "go", Command: []string{}, Out: "gen"},
			expectedMessage: "exactly one of name, path, command or remote is required",
		},
		{
			name:            "empty command executable",
			plugin:          Plugin{Command: []string{"", "./tools/run-plugin"}, Out: "gen"},
			expectedMessage: "command executable is required",
		},
		{
			name:            "missing output directory",
			plugin:          Plugin{Name: "go"},
			expectedMessage: "out is required",
		},
		{
			name:            "moving latest version",
			plugin:          Plugin{Remote: "example.com/go", Version: "go:latest", Out: "gen"},
			expectedMessage: "latest version is not reproducible",
		},
		{
			name:            "unverifiable local version",
			plugin:          Plugin{Name: "go", Version: "v1.2.3", Out: "gen"},
			expectedMessage: "version cannot be verified",
		},
		{
			name:            "unverifiable binary path version",
			plugin:          Plugin{Path: "./tools/protoc-gen-custom", Version: "v1.2.3", Out: "gen"},
			expectedMessage: "version cannot be verified",
		},
		{
			name:            "unverifiable custom command version",
			plugin:          Plugin{Command: []string{"sh", "./tools/run-plugin"}, Version: "v1.2.3", Out: "gen"},
			expectedMessage: "version cannot be verified",
		},
		{
			name:            "unpinned remote plugin",
			plugin:          Plugin{Remote: "example.com/go", Out: "gen"},
			expectedMessage: "remote plugin requires a pinned semantic version",
		},
		{
			name:            "version embedded in remote name",
			plugin:          Plugin{Remote: "example.com/go:v1.2.3", Version: "v1.2.3", Out: "gen"},
			expectedMessage: "specify the remote plugin version only in version",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.plugin.Validate()

			if tt.expectedMessage != "" {
				require.ErrorContains(t, err, tt.expectedMessage)
				return
			}
			require.NoError(t, err)
		})
	}
}
