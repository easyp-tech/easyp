package gitmodules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestSnapshotSHA256RepositoryAliases(t *testing.T) {
	t.Parallel()
	repository := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repository, "real"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repository, "real/file.proto"), []byte("syntax = \"proto3\"; package sha.v1;\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repository, "protobuf.mod"), []byte("module "+repository+"\nroots api\n"), 0o644))
	require.NoError(t, os.Symlink("real", filepath.Join(repository, "api")))
	runTestGit(t, repository, "init", "--quiet", "--object-format=sha256")
	runTestGit(t, repository, "add", ".")
	runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "SHA256 alias")
	commit := runTestGit(t, repository, "rev-parse", "HEAD")
	require.Len(t, commit[:len(commit)-1], 64)
	runTestGit(t, repository, "tag", "v1.0.0")
	for _, ref := range []string{"", "v1.0.0"} {
		t.Run("ref="+ref, func(t *testing.T) {
			t.Parallel()
			cache := &Cache{root: t.TempDir()}
			fetched, err := cache.Fetch(t.Context(), repository, ref)
			require.NoError(t, err)
			assert.Len(t, fetched.Lock.Commit, 64)
			assert.Contains(t, fetched.Lock.Hash, "h1:")
			lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}
			require.NoError(t, cache.Install(t.Context(), lock))
			directory, _, err := cache.Cached(fetched.Lock)
			require.NoError(t, err)
			content, err := os.ReadFile(filepath.Join(directory, "api/file.proto"))
			require.NoError(t, err)
			assert.Equal(t, "syntax = \"proto3\"; package sha.v1;\n", string(content))
			require.NoError(t, cache.VerifyCached(t.Context(), lock))
		})
	}
}
