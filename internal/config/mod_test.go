package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseModFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		envVars map[string]string
		want    []string
	}{
		{
			name: "basic_deps",
			content: `deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
  - github.com/bufbuild/protoc-gen-validate
`,
			want: []string{
				"github.com/googleapis/googleapis@common-protos-1_3_1",
				"github.com/bufbuild/protoc-gen-validate",
			},
		},
		{
			name:    "empty_deps",
			content: "deps: []\n",
			want:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mod, err := ParseModFile([]byte(tt.content))
			require.NoError(t, err)
			require.Equal(t, tt.want, mod.Deps)
		})
	}
}

func TestParseModFile_EnvExpansion(t *testing.T) {
	t.Setenv("DEP_URL", "github.com/googleapis/googleapis@v1.0.0")

	mod, err := ParseModFile([]byte(`deps:
  - ${DEP_URL}
`))
	require.NoError(t, err)
	require.Equal(t, []string{"github.com/googleapis/googleapis@v1.0.0"}, mod.Deps)
}

func TestNew_LoadsDepsFromProtobufMod(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, DefaultFileName)
	modPath := filepath.Join(dir, DefaultModFileName)

	err := os.WriteFile(cfgPath, []byte(`lint:
  use:
    - DIRECTORY_SAME_PACKAGE
`), 0o644)
	require.NoError(t, err)

	err = os.WriteFile(modPath, []byte(`deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
  - github.com/bufbuild/protoc-gen-validate
`), 0o644)
	require.NoError(t, err)

	cfg, err := New(context.Background(), cfgPath)
	require.NoError(t, err)
	require.Equal(t, []string{
		"github.com/googleapis/googleapis@common-protos-1_3_1",
		"github.com/bufbuild/protoc-gen-validate",
	}, cfg.Deps)
}

func TestNew_FallsBackToYamlDepsWhenModMissing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, DefaultFileName)

	err := os.WriteFile(cfgPath, []byte(`deps:
  - github.com/googleapis/googleapis@v1.0.0
lint:
  use:
    - DIRECTORY_SAME_PACKAGE
`), 0o644)
	require.NoError(t, err)

	cfg, err := New(context.Background(), cfgPath)
	require.NoError(t, err)
	require.Equal(t, []string{"github.com/googleapis/googleapis@v1.0.0"}, cfg.Deps)
}

func TestNew_ErrorsWhenDepsInBothFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, DefaultFileName)
	modPath := filepath.Join(dir, DefaultModFileName)

	err := os.WriteFile(cfgPath, []byte(`deps:
  - github.com/googleapis/googleapis@v1.0.0
lint:
  use:
    - DIRECTORY_SAME_PACKAGE
`), 0o644)
	require.NoError(t, err)

	err = os.WriteFile(modPath, []byte(`deps:
  - github.com/bufbuild/protoc-gen-validate
`), 0o644)
	require.NoError(t, err)

	_, err = New(context.Background(), cfgPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), DefaultModFileName)
}

func TestValidateModRaw(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name: "valid",
			content: `deps:
  - github.com/googleapis/googleapis@v1.0.0
`,
		},
		{
			name:    "empty_deps",
			content: "deps: []\n",
		},
		{
			name: "invalid_deps_type",
			content: `deps:
  not-a-list: true
`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			issues, err := ValidateModRaw([]byte(tt.content))
			require.NoError(t, err)
			require.Equal(t, tt.wantError, HasErrors(issues))
		})
	}
}
