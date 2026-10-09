package go_git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func snapshotGit(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), "git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s", out)
}

func TestSnapshotRevisionUsesCommittedInputs(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for name, body := range map[string]string{"proto/item.proto": "old proto", "protobuf.mod": "module example.com/root\nroots proto\n", "protobuf.lock": "old lock", "easyp.lock": "historical legacy lock", ".deps/common.proto": "old dependency", "README.md": "not an input"} {
		target := filepath.Join(root, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(target), 0o755))
		require.NoError(t, os.WriteFile(target, []byte(body), 0o644))
	}
	snapshotGit(t, root, "init", "-q", "-b", "baseline")
	snapshotGit(t, root, "add", ".")
	snapshotGit(t, root, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "baseline")
	require.NoError(t, os.WriteFile(filepath.Join(root, "proto/item.proto"), []byte("new proto"), 0o644))
	snapshot, err := SnapshotRevision(t.Context(), root, "baseline")
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, snapshot.Close()) })
	raw, err := os.ReadFile(filepath.Join(snapshot.Root, "proto/item.proto"))
	require.NoError(t, err)
	assert.Equal(t, "old proto", string(raw))
	raw, err = os.ReadFile(filepath.Join(snapshot.Root, ".deps/common.proto"))
	require.NoError(t, err)
	assert.Equal(t, "old dependency", string(raw))
	raw, err = os.ReadFile(filepath.Join(snapshot.Root, "easyp.lock"))
	require.NoError(t, err)
	assert.Equal(t, "historical legacy lock", string(raw))
	assert.NoFileExists(t, filepath.Join(snapshot.Root, "README.md"))
	raw, err = os.ReadFile(filepath.Join(root, "proto/item.proto"))
	require.NoError(t, err)
	assert.Equal(t, "new proto", string(raw))
	require.NoError(t, snapshot.Close())
	assert.NoDirExists(t, snapshot.Root)
}
