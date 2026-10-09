package gitmodules

import (
	"archive/zip"
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitialMigrationRejectsAliasTargetMissingFromReleasedArchive(t *testing.T) {
	t.Parallel()
	repository := snapshotPolicyFixture(t, map[string]string{
		"data.proto/target.txt": "syntax = \"proto3\"; package alias; message Item {}\n",
		"data.proto/unused.txt": "unrelated bytes\n",
	}, map[string]string{"proto/alias.proto": "../data.proto/target.txt"})
	require.NoError(t, os.Remove(filepath.Join(repository, "protobuf.mod")))
	require.NoError(t, os.WriteFile(filepath.Join(repository, "easyp.yaml"), []byte("generate:\n  inputs: [{directory: {path: proto, root: proto}}]\n"), 0o644))
	runTestGit(t, repository, "add", "-A")
	runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "legacy metadata")
	_, err := New(t.TempDir()).FetchMigrationImports(t.Context(), repository, "", "")
	require.ErrorContains(t, err, "migrationArchiveLayout")
}

func TestMigrationSourceArchiveReadsOnlyRequiredAliasTarget(t *testing.T) {
	t.Parallel()
	var data bytes.Buffer
	writer := zip.NewWriter(&data)
	for _, name := range []string{"data.proto/target.txt", "data.proto/unrelated.txt"} {
		file, err := writer.Create(name)
		require.NoError(t, err)
		_, err = file.Write([]byte("contract bytes"))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	archive, err := zip.NewReader(bytes.NewReader(data.Bytes()), int64(data.Len()))
	require.NoError(t, err)
	nodes := []migrationArchiveNode{
		{name: "alias.proto", mode: fs.ModeSymlink, data: []byte("data.proto/target.txt")},
		{name: "data.proto/", mode: fs.ModeDir},
		{name: "data.proto/target.txt", mode: 0o644},
		{name: "data.proto/unrelated.txt", mode: 0o644},
	}
	require.NoError(t, loadMigrationArchiveAliasTargets(t.Context(), archive.File, nodes, nil))
	require.Equal(t, "contract bytes", string(nodes[2].data))
	require.Nil(t, nodes[3].data)
}
