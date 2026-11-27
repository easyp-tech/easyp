package go_git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Baselines are staged below the system temp directory, so its host filesystem
// decides whether distinct Git spellings collide. Validated directories are
// cached while staging, so the colliding path comes after siblings that already
// populated the cache.
func TestSnapshotRevisionRejectsHostCaseCollisions(t *testing.T) {
	t.Parallel()
	probe := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(probe, "probe"), nil, 0o644))
	_, caseErr := os.Stat(filepath.Join(probe, "Probe"))
	caseInsensitive := caseErr == nil

	for _, tt := range []struct {
		name  string
		files []string
	}{
		{name: "file", files: []string{"proto/A.proto", "proto/a.proto", "proto/b.proto"}},
		{name: "directory", files: []string{"Proto/a.proto", "Proto/z.proto", "proto/b.proto"}},
		{name: "nested directory", files: []string{"x/Proto/a.proto", "x/b.proto", "x/proto/c.proto"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository := t.TempDir()
			snapshotGit(t, repository, "init", "-q", "-b", "baseline")
			for _, name := range tt.files {
				require.NoError(t, os.WriteFile(filepath.Join(repository, "blob"), []byte(name), 0o644))
				cmd := exec.CommandContext(t.Context(), "git", "hash-object", "-w", "--", "blob")
				cmd.Dir = repository
				hash, err := cmd.Output()
				require.NoError(t, err)
				snapshotGit(t, repository, "update-index", "--add", "--cacheinfo", "100644,"+strings.TrimSpace(string(hash))+","+name)
			}
			snapshotGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "case-sensitive Git names")

			snapshot, err := SnapshotRevision(t.Context(), repository, "baseline")

			if caseInsensitive {
				require.ErrorContains(t, err, "snapshot path collision")
				return
			}
			require.NoError(t, err)
			t.Cleanup(func() { assert.NoError(t, snapshot.Close()) })
			for _, name := range tt.files {
				raw, err := os.ReadFile(filepath.Join(snapshot.Root, name))
				require.NoError(t, err)
				assert.Equal(t, name, string(raw))
			}
		})
	}
}
