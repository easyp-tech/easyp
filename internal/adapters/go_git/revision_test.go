package go_git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSnapshotSupportsTaggedAndPinnedBaseline(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "item.proto"), []byte("old"), 0o644))
	snapshotGit(t, root, "init", "-q", "-b", "baseline")
	snapshotGit(t, root, "add", ".")
	snapshotGit(t, root, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
	snapshotGit(t, root, "tag", "v1.0.0")
	snapshotGit(t, root, "-c", "user.name=Test", "-c", "user.email=test@example.test", "tag", "-a", "v1.0.1", "-m", "annotated")
	cmd := exec.CommandContext(t.Context(), "git", "-C", root, "rev-parse", "HEAD")
	out, err := cmd.Output()
	require.NoError(t, err)
	for _, ref := range []string{"baseline", "v1.0.0", "v1.0.1", strings.TrimSpace(string(out)), "HEAD"} {
		t.Run(ref, func(t *testing.T) {
			t.Parallel()
			snapshot, err := SnapshotRevision(t.Context(), root, ref)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, snapshot.Close()) })
			raw, err := os.ReadFile(filepath.Join(snapshot.Root, "item.proto"))
			require.NoError(t, err)
			require.Equal(t, "old", string(raw))
		})
	}
}
