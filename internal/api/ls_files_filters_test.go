package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/modules"
)

func TestListV1FilesRejectsExcludedAlias(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, alias, target, imported string }{
		{name: "directory_alias", alias: "alias", target: "dep/excluded", imported: "alias/b.proto"},
		{name: "file_alias", alias: "alias.proto", target: "dep/excluded/b.proto", imported: "alias.proto"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "protobuf.mod", "module example.test/app\nroots .\nrequire example.test/dep\nreplace example.test/dep => ./dep\n")
			writeV1GenerateFixture(t, root, "dep/buf.yaml", "version: v1\nbuild:\n  excludes: [excluded]\n")
			writeV1GenerateFixture(t, root, "dep/excluded/b.proto", "syntax = \"proto3\"; message B {}\n")
			writeV1GenerateFixture(t, root, "client.proto", "syntax = \"proto3\"; import \""+tt.imported+"\"; message Client {}\n")
			require.NoError(t, os.Symlink(filepath.Join(root, tt.target), filepath.Join(root, tt.alias)))
			_, module, err := modules.ReadManifest(root)
			require.NoError(t, err)
			listed, err := listV1Files(t.Context(), root, module, true, nil)
			require.NoError(t, err)
			require.Len(t, listed.Errors, 1)
			assert.Equal(t, "import_not_found", listed.Errors[0].Code)
			assert.Contains(t, listed.Errors[0].Message, tt.imported)
			require.Len(t, listed.Files, 1)
			assert.Equal(t, "client.proto", listed.Files[0].ImportPath)
		})
	}
}

func TestListV1FilesRejectsLexicallyExcludedAlias(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, alias, imported string
		directory, outside    bool
	}{
		{name: "file_alias", alias: "excluded/alias.proto", imported: "excluded/alias.proto"},
		{name: "directory_alias", alias: "excluded/alias", imported: "excluded/alias/model.proto", directory: true},
		{name: "file_alias_outside_roots", alias: "excluded/alias.proto", imported: "excluded/alias.proto", outside: true},
		{name: "directory_alias_outside_roots", alias: "excluded/alias", imported: "excluded/alias/model.proto", directory: true, outside: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "protobuf.mod", "module example.test/app\nroots api\nrequire example.test/dep\nreplace example.test/dep => ./dep\n")
			writeV1GenerateFixture(t, root, "dep/buf.yaml", "version: v1\nbuild:\n  excludes: [excluded]\n")
			writeV1GenerateFixture(t, root, "api/client.proto", "syntax = \"proto3\"; import \""+tt.imported+"\"; message Client {}\n")
			target := filepath.Join(root, "dep/selected")
			if tt.outside {
				target = t.TempDir()
			}
			writeV1GenerateFixture(t, target, "model.proto", "syntax = \"proto3\"; message Model {}\n")
			if !tt.directory {
				target = filepath.Join(target, "model.proto")
			}
			alias := filepath.Join(root, "dep", tt.alias)
			require.NoError(t, os.MkdirAll(filepath.Dir(alias), 0o755))
			require.NoError(t, os.Symlink(target, alias))
			_, module, err := modules.ReadManifest(root)
			require.NoError(t, err)
			listed, err := listV1Files(t.Context(), root, module, true, nil)
			require.NoError(t, err)
			require.Len(t, listed.Errors, 1)
			assert.Equal(t, "import_not_found", listed.Errors[0].Code)
			assert.Contains(t, listed.Errors[0].Message, tt.imported)
			require.Len(t, listed.Files, 1)
			assert.Equal(t, "client.proto", listed.Files[0].ImportPath)
		})
	}
}
