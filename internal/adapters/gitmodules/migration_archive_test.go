package gitmodules

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchMigrationVerifiesV0ProtoArchiveHash(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, config     string
		files, installed map[string]string
	}{
		{
			name:      "stripped import root",
			config:    "generate:\n  inputs: [{directory: {path: proto, root: proto}}]\n",
			files:     map[string]string{"proto/example.proto": "syntax = \"proto3\";", "proto/README.md": "not archived", "LICENSE": "not archived"},
			installed: map[string]string{"example.proto": "syntax = \"proto3\";"},
		},
		{
			name:      "default import root",
			config:    "generate:\n  inputs: [{directory: {path: mcp, root: .}}]\n",
			files:     map[string]string{"mcp/options.proto": "syntax = \"proto3\";", "README.md": "not archived"},
			installed: map[string]string{"mcp/options.proto": "syntax = \"proto3\";"},
		},
		{
			name:      "non-proto collision is outside the released archive",
			config:    "generate:\n  inputs: [{directory: {path: proto, root: proto}}]\n",
			files:     map[string]string{"proto/example.proto": "syntax = \"proto3\";", "proto/README.md": "inside root", "README.md": "outside root"},
			installed: map[string]string{"example.proto": "syntax = \"proto3\";"},
		},
		{
			name:      "proto-suffixed directory keeps import name",
			files:     map[string]string{"api.proto/message.proto": "syntax = \"proto3\";", "api.proto/README.md": "archived directory content", "README.md": "not archived"},
			installed: map[string]string{"api.proto/message.proto": "syntax = \"proto3\";"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.config != "" {
				tt.files["easyp.yaml"] = tt.config
			}
			repository, commit := migrationTestRepository(t, tt.files)
			legacyHash := migrationTestHash(t, tt.installed)
			fetched, err := (&Cache{root: t.TempDir()}).FetchMigration(t.Context(), repository, commit, legacyHash)
			require.NoError(t, err)
			assert.Equal(t, commit, fetched.Lock.Commit)
			assert.Equal(t, commit, fetched.Lock.Version)
			assert.Equal(t, migrationTestHash(t, tt.files), fetched.Lock.Hash)
			assert.NotEqual(t, legacyHash, fetched.Lock.Hash)
		})
	}
}

func TestHashMigrationProtoArchiveRejectsCollisionsAndCleansUp(t *testing.T) {
	// Keep this sequential because os.MkdirTemp uses the process temp environment.
	for _, tt := range []struct {
		name      string
		files     map[string]string
		wantError string
	}{
		{name: "file collision", files: map[string]string{"a/file.proto": "first", "b/file.proto": "second"}, wantError: "legacy file collision"},
		{name: "directory and file collision", files: map[string]string{"a/nested.proto": "file", "b/nested.proto/file.proto": "directory"}, wantError: "legacy directory/file collision"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.files["buf.work.yaml"] = "version: v1\ndirectories: [a, b]\n"
			repository, _ := migrationTestRepository(t, tt.files)
			files, err := trackedV1Files(t.Context(), repository)
			require.NoError(t, err)
			archiveTemp := t.TempDir()
			for _, name := range []string{"TMPDIR", "TMP", "TEMP"} {
				t.Setenv(name, archiveTemp)
			}
			_, err = hashMigrationProtoArchive(t.Context(), repository, files)
			require.ErrorContains(t, err, tt.wantError)
			entries, err := os.ReadDir(archiveTemp)
			require.NoError(t, err)
			assert.Empty(t, entries, "failed verification must remove its temporary archive")
		})
	}
}

func TestFetchMigrationRejectsChangedV0ArchiveSources(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, attribute, source string
	}{
		{name: "export ignored proto", attribute: "proto/ignored.proto export-ignore\n", source: "syntax = \"proto3\";"},
		{name: "export substituted proto", attribute: "proto/ignored.proto export-subst\n", source: "// $Format:%H$\nsyntax = \"proto3\";"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			files := map[string]string{
				"easyp.yaml":          "generate:\n  inputs: [{directory: {path: proto, root: proto}}]\n",
				".gitattributes":      tt.attribute,
				"proto/kept.proto":    "syntax = \"proto3\";",
				"proto/ignored.proto": tt.source,
			}
			repository, commit := migrationTestRepository(t, files)
			installed := map[string]string{"kept.proto": files["proto/kept.proto"]}
			if tt.name == "export substituted proto" {
				installed["ignored.proto"] = strings.ReplaceAll(tt.source, "$Format:%H$", commit)
			}
			_, err := (&Cache{root: t.TempDir()}).FetchMigration(t.Context(), repository, commit, migrationTestHash(t, installed))
			require.ErrorContains(t, err, "source selection")
			require.ErrorContains(t, err, "manual migration")
		})
	}
}
