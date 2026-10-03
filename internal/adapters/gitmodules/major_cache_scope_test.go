package gitmodules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestMajorCachedModuleDoesNotParseUnrelatedManifest(t *testing.T) {
	t.Parallel()
	remote := t.TempDir()
	source := filepath.ToSlash(filepath.Join(remote, "sub/v2"))
	require.NoError(t, os.MkdirAll(filepath.Join(remote, "sub/proto"), 0o755))
	// Another independently versioned module must not affect the selected one.
	require.NoError(t, os.WriteFile(filepath.Join(remote, "protobuf.mod"), []byte("module "+filepath.ToSlash(remote)+"\nrequire example.com/unrelated v2.0.0\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(remote, "sub/protobuf.mod"), []byte("module "+source+"\nroots proto\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(remote, "sub/proto/item.proto"), []byte("syntax = \"proto3\"; package sub.v2; message Item {}\n"), 0o644))
	runTestGit(t, remote, "init", "-q")
	runTestGit(t, remote, "add", ".")
	runTestGit(t, remote, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "independent modules")
	runTestGit(t, remote, "tag", "sub/v2.0.0")
	cache := New(t.TempDir())
	commit := strings.TrimSpace(runTestGit(t, remote, "rev-parse", "HEAD"))
	for _, version := range []string{"v2.0.0", commit, ""} {
		fetched, err := cache.Fetch(t.Context(), source, version)
		require.NoError(t, err)
		require.NoError(t, cache.Install(t.Context(), v1.Lock{Version: 1, Modules: []v1.LockedModule{fetched.Lock}}))
		_, module, err := cache.Cached(fetched.Lock)
		require.NoError(t, err)
		assert.Equal(t, source, module.Name)
		assert.Equal(t, []string{filepath.Join("sub", "proto")}, module.Roots)
	}
}
