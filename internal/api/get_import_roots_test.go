package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestGetImportRootValidationPrecedesDependencyAccess(t *testing.T) {
	for _, tt := range []struct{ name, root string }{
		{name: "escaping", root: "../outside"},
		{name: "absolute", root: "/outside"},
		{name: "noncanonical", root: "api//svc"},
		{name: "backslash", root: "api\\svc"},
		{name: "glob", root: "api/*"},
		{name: "empty", root: ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			manifest := []byte("module example.com/consumer\n")
			require.NoError(t, os.WriteFile(filepath.Join(root, "protobuf.mod"), manifest, 0o644))
			cache := filepath.Join(root, "cache-that-must-not-be-created")
			t.Setenv("EASYPPATH", cache)
			t.Chdir(root)
			app := &cli.App{Commands: []*cli.Command{(Get{}).Command()}, Metadata: map[string]any{}}
			err := app.Run([]string{"easyp", "get", "--import-root", tt.root, "example.invalid/unreachable/repo@named-tag"})
			require.Error(t, err)
			require.ErrorContains(t, err, "root")
			require.NotContains(t, err.Error(), "flag provided but not defined", "the root flag must reach value validation")
			require.NotContains(t, err.Error(), "ResolveTag", "invalid roots must be rejected before querying a tag")
			current, readErr := os.ReadFile(filepath.Join(root, "protobuf.mod"))
			require.NoError(t, readErr)
			require.Equal(t, manifest, current)
			require.NoFileExists(t, filepath.Join(root, "protobuf.lock"))
			require.NoDirExists(t, cache)
		})
	}
}

func TestGetImportRootFrozenRejectsWithoutDependencyAccess(t *testing.T) {
	root := t.TempDir()
	manifest := []byte("module example.com/consumer\n")
	require.NoError(t, os.WriteFile(filepath.Join(root, "protobuf.mod"), manifest, 0o644))
	t.Chdir(root)
	app := &cli.App{Commands: []*cli.Command{(Get{}).Command()}, Metadata: map[string]any{}}
	err := app.Run([]string{"easyp", "get", "--frozen", "--import-root", "api", "example.invalid/unreachable/repo"})
	require.ErrorContains(t, err, "get is not allowed in frozen mode")
	current, readErr := os.ReadFile(filepath.Join(root, "protobuf.mod"))
	require.NoError(t, readErr)
	require.Equal(t, manifest, current)
	require.NoFileExists(t, filepath.Join(root, "protobuf.lock"))
}
