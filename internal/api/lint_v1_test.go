package api

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/logger"
)

func TestLintV1LeavesVersionlessLegacyConfigToLegacyHandler(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	configPath := filepath.Join(root, "easyp.yaml")
	err := os.WriteFile(configPath, []byte("lint:\n  use: [FILE_LOWER_SNAKE_CASE]\n"), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(root, "file.proto"), []byte("syntax = \"proto3\";\npackage example;\n"), 0o644)
	require.NoError(t, err)
	ctx := cli.NewContext(&cli.App{}, flag.NewFlagSet("test", flag.ContinueOnError), nil)
	ctx.Context = t.Context()

	handled, err := (Lint{}).actionV1(ctx, logger.NewNop(), configPath, root, root)
	require.NoError(t, err)
	assert.False(t, handled)
}

func TestLintV1UsesSelectedConfigFile(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		header string
	}{
		{name: "explicit version", header: "version: v1\n"},
		{name: "default version"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			configPath := filepath.Join(root, "strict.yaml")
			err := os.WriteFile(configPath, []byte(tc.header+"issues:\n  exclude-rules:\n    - {}\n"), 0o644)
			require.NoError(t, err)
			// An unrelated default config must not override the explicitly selected one.
			err = os.WriteFile(filepath.Join(root, "easyp.yaml"), []byte("lint:\n  use: [FILE_LOWER_SNAKE_CASE]\n"), 0o644)
			require.NoError(t, err)
			err = os.WriteFile(filepath.Join(root, "file.proto"), []byte("syntax = \"proto3\";\npackage example;\n"), 0o644)
			require.NoError(t, err)
			ctx := cli.NewContext(&cli.App{}, flag.NewFlagSet("test", flag.ContinueOnError), nil)
			ctx.Context = t.Context()

			handled, err := (Lint{}).actionV1(ctx, logger.NewNop(), configPath, root, root)
			require.NoError(t, err)
			assert.True(t, handled)
		})
	}
}
