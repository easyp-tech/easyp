package migration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestManifestBackupAndCombinedDependencies(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	legacy := "// exact original formatting\ndirect (\n  example.com/acme/a@v1.0.0 // preserve me\n)\nindirect (\n  example.com/acme/b@v1.0.0\n)\n"
	writeFixture(t, root, "protobuf.mod", legacy)
	require.NoError(t, os.Chmod(filepath.Join(root, "protobuf.mod"), 0o640))
	writeFixture(t, root, "easyp.yaml", "deps: [example.com/acme/a@v1.0.0]\ngenerate:\n  inputs:\n    - git_repo: {url: example.com/acme/c@v1.0.0}\n")
	repo := &mockRepository{fetched: make(map[string]modules.Fetched)}
	for _, source := range []string{"example.com/acme/a", "example.com/acme/b", "example.com/acme/c"} {
		repo.fetched[source+"@v1.0.0"] = modules.Fetched{
			Module: v1.Module{Name: source, Roots: []string{"."}},
			Lock:   v1.LockedModule{Source: source, Version: "v1.0.0", Commit: testCommit, Hash: testNewHash},
		}
	}
	plan, err := Build(context.Background(), Options{Dir: root, Module: "example.com/acme/api", ResolveLock: true, Repository: repo})
	require.NoError(t, err)
	var generation v1.Generate
	require.NoError(t, yaml.Unmarshal(outputContent(t, plan, "easyp.gen.yaml"), &generation))
	assert.Equal(t, []v1.GenerateModule{{Module: "example.com/acme/c"}}, generation.Generate.Modules)
	candidate := string(outputContent(t, plan, "protobuf.mod"))
	assert.Contains(t, candidate, "require example.com/acme/a v1.0.0")
	assert.Contains(t, candidate, "require example.com/acme/b v1.0.0 // indirect")
	assert.Contains(t, candidate, "require example.com/acme/c v1.0.0")
	require.NoError(t, plan.Apply())
	assert.Equal(t, legacy, string(mustRead(t, root, "protobuf.mod.v0.bak")))
	for _, name := range []string{"protobuf.mod", "protobuf.mod.v0.bak"} {
		info, err := os.Stat(filepath.Join(root, name))
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o640), info.Mode().Perm())
	}
	second, err := Build(context.Background(), Options{Dir: root, Module: "example.com/acme/api"})
	require.NoError(t, err)
	assert.True(t, second.AlreadyV1())
	require.NoError(t, second.Apply())
}

func TestDirectoryPathIsRelativeToRoot(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, extra, wantError string }{
		{name: "same_import_names"},
		{name: "path_selection_preserves_scope", extra: "proto/other.proto"},
		{name: "hidden_scope_narrowing", extra: "proto/api/.hidden/extra.proto", wantError: "scope"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeFixture(t, root, "easyp.yaml", "generate:\n  inputs: [{directory: {root: proto, path: api}}]\n")
			writeFixture(t, root, "proto/api/a.proto", "syntax = \"proto3\";\n")
			if tt.extra != "" {
				writeFixture(t, root, tt.extra, "syntax = \"proto3\";\n")
			}
			plan, err := Build(context.Background(), Options{Dir: root, Module: "example.com/acme/api"})
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, map[string]string{"api/a.proto": filepath.Join("proto", "api", "a.proto")}, plan.sources)
			assert.Equal(t, []string{"proto"}, plan.roots)
			assert.Equal(t, []string{"proto/api"}, plan.paths)
		})
	}
}

func TestOutputCopiesCannotChangeApplication(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFixture(t, root, "easyp.yaml", "lint: {}\n")
	plan, err := Build(context.Background(), Options{Dir: root, Module: "example.com/acme/api"})
	require.NoError(t, err)
	for _, output := range plan.Outputs() {
		for i := range output.Content {
			output.Content[i] = 'X'
		}
	}
	require.NoError(t, plan.Apply())
	assert.Contains(t, string(mustRead(t, root, "easyp.yaml")), "version: v1")
}
